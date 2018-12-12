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

import "github.com/ethereum/go-ethereum/common"

// Derive and persist resolver and registry records from event logs and create Domain views using these records
type Domain struct {
	Name         string
	Namehash     common.Hash
	LabelHash    common.Hash
	ParentDomain *Domain
	SubDomains   []*Domain
	Owner        common.Address
	ResolverAddr common.Address
	PointsToAddr common.Address
}

type DomainModel struct {
	Name         string   `db:"name"`
	Namehash     []byte   `db:"name_hash"`
	LabelHash    []byte   `db:"label_hash"`
	ParentDomain []byte   `db:"parent_domain"`
	SubDomains   [][]byte `db:"subdomains"`
	Owner        []byte   `db:"owner"`
	ResolverAddr []byte   `db:"resolver"`
	PointsToAddr []byte   `db:"points_to"`
}

type ResolverRecord struct {
	Node      []byte
	Address   string
	Content   []byte
	Name      string
	PublicKey struct {
		X []byte
		Y []byte
	}
	Abis map[uint][]byte
}

type RegistryRecord struct {
	Node     []byte
	Owner    string
	Resolver string
	Ttl      uint
}
