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

package constants

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/vulcanize/vulcanizedb/pkg/core"
	"github.com/vulcanize/vulcanizedb/pkg/filters"
	"github.com/vulcanize/vulcanizedb/pkg/omni/shared/helpers"
)

var PublicResolverAddress = "0x1da022710dF5002339274AaDEe8D58218e9D6AB5"

var PublicResolverABI = `[{"constant":true,"inputs":[{"name":"interfaceID","type":"bytes4"}],"name":"supportsInterface","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"},{"name":"contentTypes","type":"uint256"}],"name":"ABI","outputs":[{"name":"contentType","type":"uint256"},{"name":"data","type":"bytes"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"x","type":"bytes32"},{"name":"y","type":"bytes32"}],"name":"setPubkey","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"content","outputs":[{"name":"ret","type":"bytes32"}],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"addr","outputs":[{"name":"ret","type":"address"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"contentType","type":"uint256"},{"name":"data","type":"bytes"}],"name":"setABI","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"name","outputs":[{"name":"ret","type":"string"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"name","type":"string"}],"name":"setName","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"hash","type":"bytes32"}],"name":"setContent","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"pubkey","outputs":[{"name":"x","type":"bytes32"},{"name":"y","type":"bytes32"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"addr","type":"address"}],"name":"setAddr","outputs":[],"payable":false,"type":"function"},{"inputs":[{"name":"ensAddr","type":"address"}],"payable":false,"type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"a","type":"address"}],"name":"AddrChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"hash","type":"bytes32"}],"name":"ContentChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"name","type":"string"}],"name":"NameChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":true,"name":"contentType","type":"uint256"}],"name":"ABIChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"x","type":"bytes32"},{"indexed":false,"name":"y","type":"bytes32"}],"name":"PubkeyChanged","type":"event"}]`

var StartingBlock = int64(3648359)

var Filters = []filters.LogFilter{
	{
		Name:      "AddrChanged",
		FromBlock: StartingBlock,
		ToBlock:   -1,
		Address:   PublicResolverAddress,
		Topics:    core.Topics{helpers.GenerateSignature("AddrChanged(bytes32,address)")},
	},
}

// Resolver interface signatures
type Interface int

const (
	MetaSig Interface = iota
	AddrChangeSig
	ContentChangeSig
	NameChangeSig
	AbiChangeSig
	PubkeyChangeSig
	TextChangeSig
	MultihashChangeSig
)

func (e Interface) Hex() string {
	strings := [...]string{
		"0x01ffc9a7",
		"0x3b3b57de",
		"0xd8389dc5",
		"0x691f3431",
		"0x2203ab56",
		"0xc8690233",
		"0x59d1d43c",
		"0xe89401a1",
	}

	if e < MetaSig || e > MultihashChangeSig {
		return "Unknown"
	}

	return strings[e]
}

func (e Interface) Bytes() [4]uint8 {
	if e < MetaSig || e > MultihashChangeSig {
		return [4]byte{}
	}

	str := e.Hex()
	by, _ := hexutil.Decode(str)
	var byArray [4]uint8
	for i := 0; i < 4; i++ {
		byArray[i] = by[i]
	}

	return byArray
}

func (e Interface) EventSig() string {
	strings := [...]string{
		"",
		"AddrChanged(bytes32,address)",
		"ContentChanged(bytes32,bytes32)",
		"NameChanged(bytes32,string)",
		"ABIChanged(bytes32,uint256)",
		"PubkeyChanged(bytes32,bytes32,bytes32)",
		"TextChanged(bytes32,string,string)",
		"MultihashChanged(bytes32,bytes)",
	}

	if e < MetaSig || e > MultihashChangeSig {
		return "Unknown"
	}

	return strings[e]
}

func (e Interface) MethodSig() string {
	strings := [...]string{
		"supportsInterface(bytes4)",
		"addr(bytes32)",
		"content(bytes32)",
		"name(bytes32)",
		"ABI(bytes32,uint256)",
		"pubkey(bytes32)",
		"text(bytes32,string)",
		"multihash(bytes32)",
	}

	if e < MetaSig || e > MultihashChangeSig {
		return "Unknown"
	}

	return strings[e]
}
