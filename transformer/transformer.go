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
	"strings"

	"github.com/ethereum/go-ethereum/common"
	s1 "github.com/vulcanize/vulcanizedb/libraries/shared"
	"github.com/vulcanize/vulcanizedb/pkg/core"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"
	"github.com/vulcanize/vulcanizedb/pkg/omni/light/fetcher"
	hr "github.com/vulcanize/vulcanizedb/pkg/omni/light/repository"

	"github.com/vulcanize/ens_watcher/transformer/shared"
)

type ENSTransformer struct {
	Config           s1.ContractConfig
	Converter        shared.Converter
	Repository       shared.Repository
	HeaderRepository hr.HeaderRepository
	Fetcher          fetcher.Fetcher
}

func NewTransformer(db *postgres.DB, bc core.BlockChain) s1.Transformer {
	return &ENSTransformer{}
}

// Execute across all converters to collect all events
func (tr *ENSTransformer) Execute() error {
	// Iterate through events
	for _, filter := range tr.Config.Filters {
		// Filter using the event signature
		topics := [][]common.Hash{{common.HexToHash(filter.Topics[0])}}

		// Generate eventID and use it to create a checked_header column if one does not already exist
		eventId := strings.ToLower(filter.Name + "_" + filter.Address)
		if err := tr.HeaderRepository.AddCheckColumn(eventId); err != nil {
			return err
		}

		missingHeaders, err := tr.HeaderRepository.MissingHeaders(filter.FromBlock, filter.ToBlock, eventId)
		if err != nil {
			return err
		}
		for _, header := range missingHeaders {
			logs, err := tr.Fetcher.FetchLogs([]string{tr.Config.Address}, topics, header)
			if err != nil {
				return err
			}

			if len(logs) < 1 {
				err = tr.HeaderRepository.MarkHeaderChecked(header.Id, eventId)
				if err != nil {
					return err
				}

				continue
			}

			entities, err := tr.Converter.ToEntities(tr.Config.Abi, logs)
			if err != nil {
				return err
			}

			models, err := tr.Converter.ToModels(entities)
			if err != nil {
				return err
			}

			err = tr.Repository.Create(header.Id, models)
			if err != nil {
				return err
			}

		}
	}

	return nil
}
