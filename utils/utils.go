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

package utils

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func NameHash(name string) common.Hash {
	if name == "" {
		return common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")
	}
	labels := strings.Split(name, ".")
	labelHash := crypto.Keccak256([]byte(labels[len(labels)-1]))
	remainderHash := NameHash(strings.Join(labels[:len(labels)-1], ".")).Bytes()
	return crypto.Keccak256Hash(append(remainderHash, labelHash...))
}

func CreateSubnode(node, label []byte) common.Hash {
	return crypto.Keccak256Hash(append(node, label...))
}
