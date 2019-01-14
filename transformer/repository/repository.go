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

package repository

import (
	"github.com/hashicorp/golang-lru"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"

	"github.com/vulcanize/ens_watcher/transformer/models"
)

type ENSRepository interface {
	RecordExists(node string) (bool, error)
	CreateRecord(record models.DomainModel) error
	GetRecord(node string, blockNumber int64) (*models.DomainModel, error)
}

type ensRepository struct {
	db          *postgres.DB
	cachedNodes *lru.Cache
}

func NewENSRepository(db *postgres.DB) *ensRepository {
	cache, _ := lru.New(1000)
	return &ensRepository{
		db:          db,
		cachedNodes: cache,
	}
}

func (r *ensRepository) RecordExists(node string) (bool, error) {
	_, ok := r.cachedNodes.Get(node)
	if ok {
		return true, nil
	}

	var exists bool
	err := r.db.Get(&exists,
		`SELECT EXISTS(SELECT 1 
				FROM public.domain_records
				WHERE name_hash = $1
				LIMIT 1)`,
		node)

	return exists, err
}

func (r *ensRepository) CreateRecord(record models.DomainModel) error {
	_, err := r.db.Exec(
		`INSERT INTO public.domain_records
			    (block_number, 
			    name_hash, 
				label_hash, 
				parent_hash, 
				owner_addr, 
				resolver_addr, 
				points_to_addr, 
				resolved_name, 
				content_hash,
				content_type,
				pub_key_x,
				pub_key_y,
				ttl)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			    ON CONFLICT (block_number, name_hash) DO UPDATE SET
				(block_number, 
			    name_hash, 
				label_hash, 
				parent_hash, 
				owner_addr, 
				resolver_addr, 
				points_to_addr, 
				resolved_name, 
				content_hash,
				content_type,
				pub_key_x,
				pub_key_y,
				ttl) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		record.BlockNumber,
		record.NameHash,
		record.LabelHash,
		record.ParentHash,
		record.Owner,
		record.ResolverAddr,
		record.PointsToAddr,
		record.Name,
		record.ContentHash,
		record.ContentType,
		record.PubKeyX,
		record.PubKeyY,
		record.TTL,
	)

	if err != nil {
		return err
	}

	r.cachedNodes.Add(record.NameHash, true)

	return nil
}

// Gets the record for the give node at the given blockheight
// We only store a new records when something has changed, so if a record does not exist for that precise blockheight
// The most recent record previous to that blockheight is the state of the record at that blockheight
func (r *ensRepository) GetRecord(node string, blockNumber int64) (*models.DomainModel, error) {
	exists, err := r.RecordExists(node)
	if err != nil {
		return nil, err
	}
	if !exists {
		return &models.DomainModel{}, nil
	}
	var result models.DomainModel
	err = r.db.Get(&result,
		`SELECT block_number, 
			    name_hash, 
				label_hash, 
				parent_hash, 
				owner_addr, 
				resolver_addr, 
				points_to_addr, 
				resolved_name, 
				content_hash,
				content_type,
				pub_key_x,
				pub_key_y,
				ttl
		 FROM public.domain_records
		 WHERE name_hash = $1
		 AND block_number <= $2 
		 ORDER BY block_number DESC LIMIT 1`,
		node, blockNumber,
	)

	return &result, err
}
