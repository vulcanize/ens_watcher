// VulcanizeDB
// Copyright © 2018 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package transformer

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vulcanize/vulcanizedb/pkg/core"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"
	"github.com/vulcanize/vulcanizedb/pkg/omni/light/converter"
	"github.com/vulcanize/vulcanizedb/pkg/omni/light/fetcher"
	"github.com/vulcanize/vulcanizedb/pkg/omni/light/repository"
	"github.com/vulcanize/vulcanizedb/pkg/omni/light/retriever"
	"github.com/vulcanize/vulcanizedb/pkg/omni/shared/constants"
	"github.com/vulcanize/vulcanizedb/pkg/omni/shared/contract"
	"github.com/vulcanize/vulcanizedb/pkg/omni/shared/getter"
	"github.com/vulcanize/vulcanizedb/pkg/omni/shared/parser"
	"github.com/vulcanize/vulcanizedb/pkg/omni/shared/types"

	"github.com/vulcanize/ens_watcher/transformer/models"
	trep "github.com/vulcanize/ens_watcher/transformer/repository"
	"github.com/vulcanize/ens_watcher/transformer/utils"
)

// Requires a light synced vDB (headers) and a running eth node (or infura)
type transformer struct {
	// Database interfaces
	trep.ENSRepository          // Repository for ENS domain records
	repository.HeaderRepository // Interface for interaction with header repositories

	// Pre-processing interfaces
	parser.Parser            // Parses events and methods out of contract abi fetched using contract address
	getter.InterfaceGetter   // Used to check the interface of resolvers
	retriever.BlockRetriever // Retrieves first block for contract and current block height

	// Processing interfaces
	fetcher.Fetcher     // Fetches event logs, using header hashes
	converter.Converter // Converts watched event logs into custom log

	// Ethereum network name
	// Default "" network is mainnet
	Network string

	// Registry contract
	Registry             *contract.Contract
	registryEventIds     []string
	registryEventFilters []common.Hash

	// Resolver addresses and contracts
	ResolverAddresses    map[string]bool
	Resolvers            map[string]*contract.Contract
	resolverEventIds     map[string][]string
	resolverEventFilters map[string][]common.Hash
	invalidResolvers     map[string]bool
}

// Order-of-operations:
// 1. Create new transformer
// 2. Initialize registry contract
// 3. Execute

// Transformer takes in config for blockchain, database, and network id
func NewTransformer(network string, bc core.BlockChain, db *postgres.DB) (*transformer, error) {
	if network != "" && network != "ropsten" {
		return nil, errors.New(`invalid network id; only mainnet ("") and ropsten ("ropsten") are allowed`)
	}

	return &transformer{
		Network:          network,
		Fetcher:          fetcher.NewFetcher(bc),
		Parser:           parser.NewParser(network),
		HeaderRepository: repository.NewHeaderRepository(db),
		Converter:        converter.NewConverter(&contract.Contract{}),
		Resolvers:        map[string]*contract.Contract{},
		ENSRepository:    trep.NewENSRepository(db),
		InterfaceGetter:  getter.NewInterfaceGetter(bc),
		BlockRetriever:   retriever.NewBlockRetriever(db),
	}, nil
}

// Initializes transformer with the registry contract info
func (tr *transformer) Init() error {
	// Get registry abi (mainnet and ropsten contracts have same abi)
	err := tr.Parser.Parse(constants.EnsContractAddress)
	if err != nil {
		return err
	}

	var addr string
	var start int64
	if tr.Network == "ropsten" {
		addr = "0x112234455C3a32FD11230C42E7Bccd4A84e02010"
		start = 25409
	} else {
		addr = constants.EnsContractAddress
		start = 3327417
		tr.Network = "mainnet"
	}

	// Aggregate info into registry contract object and store for execution
	tr.Registry = &contract.Contract{
		Name:          "ENS-Registry",
		Network:       tr.Network,
		Address:       addr,
		Abi:           tr.Parser.Abi(),
		ParsedAbi:     tr.Parser.ParsedAbi(),
		StartingBlock: start,
		LastBlock:     -1,
		Events:        tr.Parser.GetEvents([]string{}), // Watch all events (NewOwner, Transfer, NewTTL, and NewResolver)
		Methods:       nil,
		FilterArgs:    map[string]bool{},
		MethodArgs:    map[string]bool{},
	}

	tr.registryEventIds = make([]string, 0, 4)
	tr.registryEventFilters = make([]common.Hash, 0, 4)
	tr.resolverEventIds = make(map[string][]string)
	tr.resolverEventFilters = make(map[string][]common.Hash)

	for _, e := range tr.Registry.Events {
		// Generate eventID and use it to create a checked_header column if one does not already exist
		eventId := strings.ToLower(e.Name + "_" + addr)
		err := tr.HeaderRepository.AddCheckColumn(eventId)
		if err != nil {
			return err
		}
		tr.registryEventIds = append(tr.registryEventIds, eventId)
		tr.registryEventFilters = append(tr.registryEventFilters, e.Sig())
	}

	tr.ResolverAddresses = make(map[string]bool)
	tr.Resolvers = make(map[string]*contract.Contract)
	tr.resolverEventIds = make(map[string][]string)
	tr.resolverEventFilters = make(map[string][]common.Hash)
	tr.invalidResolvers = make(map[string]bool)
	tr.invalidResolvers["0x0000000000000000000000000000000000000000"] = true

	return nil
}

// Executes over registry contract
// Also finds new resolver contracts emitted from NewResolver events and executes over them
func (tr *transformer) Execute() error {
	// Configure converter with the registry contract
	tr.Converter.Update(tr.Registry)

	// This is to make sure that we use the same range for all calls to MissingHeadersForAll in this pass
	lastBlock, err := tr.BlockRetriever.RetrieveMostRecentBlock()
	if err != nil {
		return err
	}

	// Retrieve unchecked headers for the registry
	missingHeaders, err := tr.HeaderRepository.MissingHeadersForAll(tr.Registry.StartingBlock, lastBlock, tr.registryEventIds)
	if err != nil {
		return err
	}

	// Iterate over headers
	for _, header := range missingHeaders {
		// And collect registry event logs
		logs, err := tr.Fetcher.FetchLogs([]string{tr.Registry.Address}, tr.registryEventFilters, header)
		if err != nil {
			return err
		}

		// If no logs are found mark the header checked for all of these eventIDs and continue
		if len(logs) < 1 {
			err = tr.HeaderRepository.MarkHeaderCheckedForAll(header.Id, tr.registryEventIds)
			if err != nil {
				return err
			}
			continue
		}

		// Convert logs into batches of log mappings (eventName => []types.Log)
		convertedLogs, err := tr.Converter.ConvertBatch(logs, tr.Registry.Events, header.Id)
		if err != nil {
			return err
		}

		// Process the registry log data into our domain records
		err = tr.processRegistryLogs(convertedLogs, header.BlockNumber)
		if err != nil {
			return err
		}

		// Mark this header checked for registry events
		err = tr.HeaderRepository.MarkHeaderCheckedForAll(header.Id, tr.registryEventIds)
		if err != nil {
			return err
		}

		// Configure any new resolver contracts that were seen in NewResolver events
		err = tr.configResolvers(header.BlockNumber)
		if err != nil {
			return err
		}
	}

	// Watch resolver contracts for the same block range
	err = tr.watchResolvers(lastBlock)
	if err != nil {
		return err
	}

	return nil
}

// Process the log data from Registry events into domain record objects
// Keeps track of Resolver addresses that are seen emitted so that we can watch them downstream
func (tr *transformer) processRegistryLogs(logs map[string][]types.Log, blockNumber int64) error {
	// Process registry NewOwner logs
	for _, newOwner := range logs["NewOwner"] {
		parentHash := newOwner.Values["node"]
		labelHash := newOwner.Values["label"]
		subnode := utils.CreateSubnode(parentHash, labelHash)
		var record *models.DomainModel
		exists, err := tr.ENSRepository.RecordExists(subnode)
		if err != nil {
			return err
		}
		if exists { // If a record already exists for this subdomain, retrieve it for updating
			record, err = tr.ENSRepository.GetRecord(subnode, blockNumber)
			if err != nil {
				return err
			}
		} else { // If no previous record exists for this subdomain, create a new one
			record = &models.DomainModel{}
		}
		// Update the new or retrieved record with values emitted from this log
		record.NameHash = subnode
		record.ParentHash = parentHash
		record.LabelHash = labelHash
		record.Owner = newOwner.Values["owner"]
		record.BlockNumber = blockNumber
		// Persist the new or updated record
		err = tr.ENSRepository.CreateRecord(*record)
		if err != nil {
			return err
		}
	}

	// Note that for all other logs a record should already exist (NewOwner event from domain's creation must have already occurred)
	// Process registry Transfer logs
	for _, transfer := range logs["Transfer"] {
		// Get most recent/current record
		lastRecord, err := tr.ENSRepository.GetRecord(transfer.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed owner and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.Owner = transfer.Values["owner"]
		// Persist updated record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
	}

	// Process registry NewTTL logs
	for _, ttl := range logs["NewTTL"] {
		// Get most recent state
		lastRecord, err := tr.ENSRepository.GetRecord(ttl.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed ttl and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.TTL = ttl.Values["ttl"]
		// Persist new record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
	}

	// Process registry NewResolver logs
	for _, newResolver := range logs["NewResolver"] {
		// Get most recent state
		lastRecord, err := tr.ENSRepository.GetRecord(newResolver.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed resolver address and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.ResolverAddr = newResolver.Values["resolver"]
		// Persist new record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
		// Add resolver address to list of resolver addresses
		tr.ResolverAddresses[newResolver.Values["resolver"]] = true
	}

	return nil
}

// Configures contracts for watching Resolvers we found emitted from the Registry's NewResolver events
func (tr *transformer) configResolvers(blockNumber int64) error {
	for resolverAddr := range tr.ResolverAddresses {
		_, ok := tr.Resolvers[resolverAddr]
		if ok { // Resolver contract has either already been setup or we already know it is invalid
			continue
		}
		_, ok = tr.invalidResolvers[resolverAddr]
		if ok {
			continue
		}
		// Construct the abi for this resolver
		abiStr := tr.InterfaceGetter.GetABI(resolverAddr, blockNumber)
		if abiStr == "" {
			// If abi is empty and we don't support any of the desired interfaces, skip configuring this resolver and add it to the list of invalid resolver so
			// we don't keep checking the domain records that use this resolver will be incomplete, but we can continue to collect their data from the registry
			tr.invalidResolvers[resolverAddr] = true
			continue
		}

		// If it does, use the standard ABI
		err := tr.Parser.ParseAbiStr(abiStr)
		if err != nil {
			return err
		}

		// Aggregate info into resolver contract object and store for execution
		tr.Resolvers[resolverAddr] = &contract.Contract{
			Name:          "ENS-Resolver",
			Network:       tr.Network,
			Address:       resolverAddr,
			Abi:           tr.Parser.Abi(),
			ParsedAbi:     tr.Parser.ParsedAbi(),
			StartingBlock: blockNumber, // Start the resolver contract at the blockheight it was first seen emitted by the Registry from NewResolver events
			LastBlock:     -1,
			Events:        tr.Parser.GetEvents([]string{}), // Watch all resolver events
			Methods:       nil,
			FilterArgs:    map[string]bool{},
			MethodArgs:    map[string]bool{},
		}

		// Create checked_headers columns, event ids, and event sigs for this resolver
		for _, e := range tr.Resolvers[resolverAddr].Events {
			eventId := strings.ToLower(e.Name + "_" + resolverAddr)
			err := tr.HeaderRepository.AddCheckColumn(eventId)
			if err != nil {
				return err
			}
			tr.resolverEventIds[resolverAddr] = append(tr.resolverEventIds[resolverAddr], eventId)
			tr.resolverEventFilters[resolverAddr] = append(tr.resolverEventFilters[resolverAddr], e.Sig())
		}
	}

	return nil
}

// Watches the configured Resolvers
func (tr *transformer) watchResolvers(lastBlock int64) error {
	// Iterate over resolver contracts
	for addr, resolver := range tr.Resolvers {
		// Update converter with this contract
		tr.Converter.Update(resolver)

		// Retrieve unchecked headers for this resolver
		missingHeaders, err := tr.HeaderRepository.MissingHeadersForAll(resolver.StartingBlock, lastBlock, tr.resolverEventIds[addr])
		if err != nil {
			return err
		}

		// Iterate over headers
		for _, header := range missingHeaders {
			// And collect event logs for this resolver
			logs, err := tr.Fetcher.FetchLogs([]string{addr}, tr.resolverEventFilters[addr], header)
			if err != nil {
				return err
			}

			// If no logs are found mark the header checked for all of these eventIDs and continue
			if len(logs) < 1 {
				err = tr.HeaderRepository.MarkHeaderCheckedForAll(header.Id, tr.resolverEventIds[addr])
				if err != nil {
					return err
				}
				continue
			}

			// Convert logs into batches of log mappings (eventName => []types.Log)
			convertedLogs, err := tr.Converter.ConvertBatch(logs, resolver.Events, header.Id)
			if err != nil {
				return err
			}

			// Process the resolver log data into our domain records
			err = tr.processResolverLogs(convertedLogs, header.BlockNumber)
			if err != nil {
				return err
			}

			// Mark this header checked for resolver events
			err = tr.HeaderRepository.MarkHeaderCheckedForAll(header.Id, tr.resolverEventIds[addr])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Processes Resolver event log data into our domain records
func (tr *transformer) processResolverLogs(logs map[string][]types.Log, blockNumber int64) error {
	// Process resolver AddrChanged logs
	for _, addrChanged := range logs["AddrChanged"] {
		// Get most recent state
		lastRecord, err := tr.ENSRepository.GetRecord(addrChanged.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed address and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.PointsToAddr = addrChanged.Values["a"]
		// Persist new record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
	}

	// Process resolver NameChanged logs
	for _, nameChanged := range logs["NameChanged"] {
		// Get most recent state
		lastRecord, err := tr.ENSRepository.GetRecord(nameChanged.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed name and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.Name = nameChanged.Values["name"]
		// Persist new record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
	}

	// Process resolver ContentChanged logs
	for _, contentChanged := range logs["ContentChanged"] {
		// Get most recent state
		lastRecord, err := tr.ENSRepository.GetRecord(contentChanged.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed content hash and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.Content = contentChanged.Values["hash"]
		// Persist new record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
	}

	// Process resolver AbiChanged logs
	for _, abiChanged := range logs["AbiChanged"] {
		// Get most recent state
		lastRecord, err := tr.ENSRepository.GetRecord(abiChanged.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed content type and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.ContentType = abiChanged.Values["contentType"]
		// Persist new record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
	}

	// Process resolver PubkeyChanged logs
	for _, pubkeyChanged := range logs["PubkeyChanged"] {
		// Get most recent state
		lastRecord, err := tr.ENSRepository.GetRecord(pubkeyChanged.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed pubkey variables and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.PubKeyX = pubkeyChanged.Values["x"]
		lastRecord.PubKeyY = pubkeyChanged.Values["y"]
		// Persist new record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
	}

	// Process resolver TextChanged logs
	for _, textChanged := range logs["TextChanged"] {
		// Get most recent state
		lastRecord, err := tr.ENSRepository.GetRecord(textChanged.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed pubkey variables and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.TextKey = textChanged.Values["key"]
		lastRecord.IndexedTextKey = textChanged.Values["indexedKey"]
		// Persist new record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
	}

	// Process resolver MultihashChanged logs
	for _, multihashChanged := range logs["MultihashChanged"] {
		// Get most recent state
		lastRecord, err := tr.ENSRepository.GetRecord(multihashChanged.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed pubkey variables and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.Multihash = multihashChanged.Values["hash"]
		// Persist new record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
	}

	// Process resolver ContenthashChanged logs
	for _, contenthashChanged := range logs["ContenthashChanged"] {
		// Get most recent state
		lastRecord, err := tr.ENSRepository.GetRecord(contenthashChanged.Values["node"], blockNumber)
		if err != nil {
			return err
		}
		// Update with changed pubkey variables and block height
		lastRecord.BlockNumber = blockNumber
		lastRecord.Contenthash = contenthashChanged.Values["hash"]
		// Persist new record
		err = tr.ENSRepository.CreateRecord(*lastRecord)
		if err != nil {
			return err
		}
	}

	return nil
}
