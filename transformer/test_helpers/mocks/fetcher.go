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

package mocks

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/vulcanize/vulcanizedb/pkg/core"
)

type mockFetcher struct {
	blockChain core.BlockChain
	Logs       []types.Log
}

func NewMockFetcher(blockchain core.BlockChain) *mockFetcher {
	return &mockFetcher{
		blockChain: blockchain,
		Logs:       []types.Log{},
	}
}

// Checks all topic0s, on all addresses, fetching matching logs for the given header
func (fetcher *mockFetcher) FetchLogs(contractAddresses []string, topic0s []common.Hash, header core.Header) ([]types.Log, error) {
	addresses := hexStringsToAddresses(contractAddresses)
	returnLogs := make([]types.Log, 0, len(fetcher.Logs))
	for _, log := range fetcher.Logs {
		if checkAllAddr(log, addresses) && checkAllSigs(log, topic0s) {
			returnLogs = append(returnLogs, log)
		}
	}

	return returnLogs, nil
}

func hexStringsToAddresses(hexStrings []string) []common.Address {
	var addresses []common.Address
	for _, hexString := range hexStrings {
		address := common.HexToAddress(hexString)
		addresses = append(addresses, address)
	}

	return addresses
}

func checkAllAddr(log types.Log, addrs []common.Address) bool {
	for _, addr := range addrs {
		if log.Address == addr {
			return true
		}
	}

	return false
}

func checkAllSigs(log types.Log, topic0s []common.Hash) bool {
	for _, topic0 := range topic0s {
		if log.Topics[0] == topic0 {
			return true
		}
	}

	return false
}
