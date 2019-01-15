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

package repository_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vulcanize/ens_watcher/transformer/models"
	"github.com/vulcanize/ens_watcher/transformer/test_helpers"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"

	"github.com/vulcanize/ens_watcher/transformer/repository"
)

var _ = Describe("Repository", func() {
	var repo repository.ENSRepository
	var db *postgres.DB
	mockRecord := models.DomainModel{
		Name:        "fakeName",
		NameHash:    "fakeNameHash",
		BlockNumber: 3327420,
		LabelHash:   "fakeLabelHash",
		ParentHash:  "fakeParentHash",
		Owner:       "fakeOwnerAddress",
	}

	BeforeEach(func() {
		db, _ = test_helpers.SetupENSRepo(3327417, -1)
		repo = repository.NewENSRepository(db)
	})

	AfterEach(func() {
		test_helpers.TearDown(db)
	})

	Describe("RecordExists", func() {
		It("Returns false if no domain record exists for the given node", func() {
			exists, err := repo.RecordExists("fakeNameHash")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(Equal(false))
		})

		It("Returns true if a domain record exists for the given node", func() {
			err := repo.CreateRecord(mockRecord)
			Expect(err).ToNot(HaveOccurred())

			exists, err := repo.RecordExists("fakeNameHash")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(Equal(true))
		})
	})

	Describe("GetRecord", func() {
		It("Returns empty record if none yet exists for the node", func() {
			exists, err := repo.RecordExists("fakeNameHash")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(Equal(false))

			record, err := repo.GetRecord("fakeNameHash", 3327420)
			Expect(err).ToNot(HaveOccurred())
			Expect(*record).To(Equal(models.DomainModel{
				NameHash: "fakeNameHash",
			}))
		})

		It("Fecthes the most up-to-date record for a given node, if it exists", func() {
			exists, err := repo.RecordExists("fakeNameHash")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(Equal(false))

			err = repo.CreateRecord(mockRecord)
			Expect(err).ToNot(HaveOccurred())

			mockRecord2 := mockRecord
			mockRecord2.BlockNumber = 3327421
			err = repo.CreateRecord(mockRecord2)
			Expect(err).ToNot(HaveOccurred())

			mockRecord3 := mockRecord
			mockRecord3.BlockNumber = 3327422
			err = repo.CreateRecord(mockRecord3)
			Expect(err).ToNot(HaveOccurred())

			record, err := repo.GetRecord("fakeNameHash", 3327425)
			Expect(err).ToNot(HaveOccurred())
			Expect(*record).To(Equal(mockRecord3))
		})
	})

	Describe("CreateRecord", func() {
		It("Creates a new domain record if one does not yet exist for the node and blockheight", func() {
			exists, err := repo.RecordExists("fakeNameHash")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(Equal(false))

			err = repo.CreateRecord(mockRecord)
			Expect(err).ToNot(HaveOccurred())

			exists, err = repo.RecordExists("fakeNameHash")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(Equal(true))
		})

		It("Updates the domain record if one already exists for the node at the blockheight", func() {
			exists, err := repo.RecordExists("fakeNameHash")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(Equal(false))

			err = repo.CreateRecord(mockRecord)
			Expect(err).ToNot(HaveOccurred())

			record, err := repo.GetRecord("fakeNameHash", 3327420)
			Expect(err).ToNot(HaveOccurred())
			Expect(*record).To(Equal(mockRecord))

			record.PointsToAddr = "fakePointsToAddress"
			err = repo.CreateRecord(*record)
			Expect(err).ToNot(HaveOccurred())

			record, err = repo.GetRecord("fakeNameHash", 3327420)
			Expect(err).ToNot(HaveOccurred())
			Expect(*record).ToNot(Equal(mockRecord))
			Expect(record.Name).To(Equal("fakeName"))
			Expect(record.NameHash).To(Equal("fakeNameHash"))
			Expect(record.BlockNumber).To(Equal(int64(3327420)))
			Expect(record.LabelHash).To(Equal("fakeLabelHash"))
			Expect(record.ParentHash).To(Equal("fakeParentHash"))
			Expect(record.Owner).To(Equal("fakeOwnerAddress"))
			Expect(record.PointsToAddr).To(Equal("fakePointsToAddress"))

			record.ResolverAddr = "fakeResolverAddress"
			err = repo.CreateRecord(*record)
			Expect(err).ToNot(HaveOccurred())

			record, err = repo.GetRecord("fakeNameHash", 3327420)
			Expect(err).ToNot(HaveOccurred())
			Expect(*record).ToNot(Equal(mockRecord))
			Expect(record.Name).To(Equal("fakeName"))
			Expect(record.NameHash).To(Equal("fakeNameHash"))
			Expect(record.BlockNumber).To(Equal(int64(3327420)))
			Expect(record.LabelHash).To(Equal("fakeLabelHash"))
			Expect(record.ParentHash).To(Equal("fakeParentHash"))
			Expect(record.Owner).To(Equal("fakeOwnerAddress"))
			Expect(record.PointsToAddr).To(Equal("fakePointsToAddress"))
			Expect(record.ResolverAddr).To(Equal("fakeResolverAddress"))
		})
	})
})
