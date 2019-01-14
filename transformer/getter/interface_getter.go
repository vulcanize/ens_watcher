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

package getter

import (
	"github.com/vulcanize/ens_watcher/transformer/constants"
	"github.com/vulcanize/vulcanizedb/examples/generic"
	"github.com/vulcanize/vulcanizedb/pkg/core"
)

type InterfaceGetter interface {
	GetSupportsResolverInterface(resolverAddr string, blockNumber int64) (bool, error)
	GetBlockChain() core.BlockChain
}

type interfaceGetter struct {
	generic.Fetcher
}

func NewInterfaceGetter(blockChain core.BlockChain) *interfaceGetter {
	return &interfaceGetter{
		Fetcher: generic.Fetcher{
			BlockChain: blockChain,
		},
	}
}

// Checks that resolver supports the standard interfaces we need it to
func (g *interfaceGetter) GetSupportsResolverInterface(resolverAddr string, blockNumber int64) (bool, error) {
	abi := constants.PublicResolverABI
	args := make([]interface{}, 1)
	args[0] = constants.AddrChangeSig.Bytes()
	supports, err := g.getSupportsInterface(abi, resolverAddr, blockNumber, args)
	if err != nil {
		return false, err
	}
	if !supports {
		return false, nil
	}
	args[0] = constants.NameChangeSig.Bytes()
	supports, err = g.getSupportsInterface(abi, resolverAddr, blockNumber, args)
	if err != nil {
		return false, err
	}
	if !supports {
		return false, nil
	}
	args[0] = constants.ContentChangeSig.Bytes()
	supports, err = g.getSupportsInterface(abi, resolverAddr, blockNumber, args)
	if err != nil {
		return false, err
	}
	if !supports {
		return false, nil
	}
	args[0] = constants.AbiChangeSig.Bytes()
	supports, err = g.getSupportsInterface(abi, resolverAddr, blockNumber, args)
	if err != nil {
		return false, err
	}
	if !supports {
		return false, nil
	}
	args[0] = constants.PubkeyChangeSig.Bytes()
	supports, err = g.getSupportsInterface(abi, resolverAddr, blockNumber, args)
	if err != nil {
		return false, err
	}
	if !supports {
		return false, nil
	}

	return true, nil
}

// Use this method to check whether or not a contract supports a given method/event interface
func (g *interfaceGetter) getSupportsInterface(contractAbi, contractAddress string, blockNumber int64, methodArgs []interface{}) (bool, error) {
	return g.Fetcher.FetchBool("supportsInterface", contractAbi, contractAddress, blockNumber, methodArgs)
}

// Method to retrieve the Getter's blockchain
func (g *interfaceGetter) GetBlockChain() core.BlockChain {
	return g.Fetcher.BlockChain
}
