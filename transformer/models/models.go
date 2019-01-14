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

package models

type DomainModel struct {
	Name         string `db:"resolved_name"`
	NameHash     string `db:"name_hash"`
	BlockNumber  int64  `db:"block_number"`
	LabelHash    string `db:"label_hash"`
	ParentHash   string `db:"parent_hash"`
	Owner        string `db:"owner_addr"`
	ResolverAddr string `db:"resolver_addr"`
	PointsToAddr string `db:"points_to_addr"`
	ContentHash  string `db:"content_hash"`
	ContentType  string `db:"content_type"`
	PubKeyX      string `db:"pub_key_x"`
	PubKeyY      string `db:"pub_key_y"`
	TTL          string `db:"ttl"`
}
