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

package transformer_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vulcanize/vulcanizedb/pkg/core"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres/repositories"
	"github.com/vulcanize/vulcanizedb/pkg/omni/shared/constants"

	"github.com/vulcanize/ens_watcher/transformer"
	"github.com/vulcanize/ens_watcher/transformer/test_helpers"
	"github.com/vulcanize/ens_watcher/transformer/test_helpers/mocks"
)

var mockLogs = []types.Log{
	{
		Address:     common.HexToAddress(constants.EnsContractAddress),
		BlockNumber: 7071367,
		BlockHash:   common.HexToHash("0xMockBlockHash01"),
		TxHash:      common.HexToHash("0xMockTxHash01"),
		TxIndex:     1,
		Index:       1,
		Removed:     false,
		Topics: []common.Hash{
			common.HexToHash("0xce0457fe73731f824cc272376169235128c118b49d344817417c6d108d155e82"),
			common.HexToHash("0xd1115c02622703bb9236a0e6609cb250a874e903494bd9071c25078f4033dac1"),
			common.HexToHash("0xadc756803e4eb4ccfb136b73d5f72e3dc0d452d30ae1f4bc82af394c73ce7115"),
		},
		Data: common.HexToHash("0x00000000000000000000000042032c22c510ad0698f16be9b99640efdeb02832").Bytes(),
	},
	{
		Address:     common.HexToAddress(constants.EnsContractAddress),
		BlockNumber: 7071427,
		BlockHash:   common.HexToHash("0xMockBlockHash02"),
		TxHash:      common.HexToHash("0xMockTxHash02"),
		TxIndex:     1,
		Index:       1,
		Removed:     false,
		Topics: []common.Hash{
			common.HexToHash("0x335721b01866dc23fbee8b6b2c7b1e14d6f05c28cd35a2c934239f94095602a0"),
			common.HexToHash("0x5954c882606735d75f2775ff380873d6d6b546f63cdf79424f12209b9e15bb91"),
		},
		Data: common.HexToHash("0x000000000000000000000000d3ddccdd3b25a8a7423b5bee360a42146eb4baf3").Bytes(),
	},
	{
		Address:     common.HexToAddress("0x000000000000000000000000d3ddccdd3b25a8a7423b5bee360a42146eb4baf3"),
		BlockNumber: 7071440,
		BlockHash:   common.HexToHash("0xMockBlockHash03"),
		TxHash:      common.HexToHash("0xMockTxHash03"),
		TxIndex:     1,
		Index:       1,
		Removed:     false,
		Topics: []common.Hash{
			common.HexToHash("0x52d7d861f09ab3d26239d492e8968629f95e9e318cf0b73bfddc441522a15fd2"),
			common.HexToHash("0x5954c882606735d75f2775ff380873d6d6b546f63cdf79424f12209b9e15bb91"),
		},
		Data: common.HexToHash("0x000000000000000000000000a54aef7fa503e75a03b262a4cd73037c1774735d").Bytes(),
	},
}

var _ = Describe("Transformer", func() {
	var db *postgres.DB
	var blockChain core.BlockChain
	var headerRepository repositories.HeaderRepository

	BeforeEach(func() {
		db, blockChain = test_helpers.SetupDBandBC()
		headerRepository = repositories.NewHeaderRepository(db)
	})

	AfterEach(func() {
		test_helpers.TearDown(db)
	})

	Describe("Init", func() {
		It("Initializes transformer's registry contract", func() {
			t, err := transformer.NewTransformer("", blockChain, db)
			Expect(err).ToNot(HaveOccurred())

			err = t.Init()
			Expect(err).ToNot(HaveOccurred())

			registryContract := t.Registry
			Expect(registryContract.Address).To(Equal(constants.EnsContractAddress))
			Expect(registryContract.StartingBlock).To(Equal(int64(3327417)))
			Expect(registryContract.LastBlock).To(Equal(int64(-1)))
			Expect(registryContract.Abi).To(Equal(constants.ENSAbiString))
			Expect(registryContract.Name).To(Equal("ENS-Registry"))
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			header1, err := blockChain.GetHeaderByNumber(7071367)
			Expect(err).ToNot(HaveOccurred())
			header2, err := blockChain.GetHeaderByNumber(7071427)
			Expect(err).ToNot(HaveOccurred())
			header3, err := blockChain.GetHeaderByNumber(7071440)
			Expect(err).ToNot(HaveOccurred())
			_, err = headerRepository.CreateOrUpdateHeader(header1)
			Expect(err).ToNot(HaveOccurred())
			_, err = headerRepository.CreateOrUpdateHeader(header2)
			Expect(err).ToNot(HaveOccurred())
			_, err = headerRepository.CreateOrUpdateHeader(header3)
			Expect(err).ToNot(HaveOccurred())
		})

		It("With Mock Fetcher: transforms registry event data into domain records", func() {
			t, err := transformer.NewTransformer("", blockChain, db)
			Expect(err).ToNot(HaveOccurred())
			f := mocks.NewMockFetcher(blockChain)
			f.Logs = mockLogs
			t.Fetcher = f
			Expect(err).ToNot(HaveOccurred())
			err = t.Init()
			err = t.Execute()
			Expect(err).ToNot(HaveOccurred())
			record, err := t.ENSRepository.GetRecord("0x5954c882606735d75f2775ff380873d6d6b546f63cdf79424f12209b9e15bb91", 7071440)
			Expect(err).ToNot(HaveOccurred())
			Expect(record.BlockNumber).To(Equal(int64(7071440)))
			Expect(record.NameHash).To(Equal("0x5954c882606735d75f2775ff380873d6d6b546f63cdf79424f12209b9e15bb91"))
			Expect(record.LabelHash).To(Equal("0xadc756803e4eb4ccfb136b73d5f72e3dc0d452d30ae1f4bc82af394c73ce7115"))
			Expect(record.ParentHash).To(Equal("0xd1115c02622703bb9236a0e6609cb250a874e903494bd9071c25078f4033dac1"))
			Expect(record.Owner).To(Equal("0x42032C22C510AD0698f16bE9b99640eFDEB02832"))
			Expect(record.ResolverAddr).To(Equal("0xD3ddcCDD3b25A8a7423B5bEe360a42146eb4Baf3"))
			Expect(record.PointsToAddr).To(Equal("0xa54AEF7fA503E75a03b262A4Cd73037C1774735D"))
		})

		It("With real fetcher: Transforms registry event data into domain records", func() {
			t, err := transformer.NewTransformer("", blockChain, db)
			Expect(err).ToNot(HaveOccurred())
			err = t.Init()
			Expect(err).ToNot(HaveOccurred())
			err = t.Execute()
			Expect(err).ToNot(HaveOccurred())
			record, err := t.ENSRepository.GetRecord("0x5954c882606735d75f2775ff380873d6d6b546f63cdf79424f12209b9e15bb91", 7071440)
			Expect(err).ToNot(HaveOccurred())
			Expect(record.BlockNumber).To(Equal(int64(7071440)))
			Expect(record.NameHash).To(Equal("0x5954c882606735d75f2775ff380873d6d6b546f63cdf79424f12209b9e15bb91"))
			Expect(record.LabelHash).To(Equal("0xadc756803e4eb4ccfb136b73d5f72e3dc0d452d30ae1f4bc82af394c73ce7115"))
			Expect(record.ParentHash).To(Equal("0xd1115c02622703bb9236a0e6609cb250a874e903494bd9071c25078f4033dac1"))
			Expect(record.Owner).To(Equal("0x42032C22C510AD0698f16bE9b99640eFDEB02832"))
			Expect(record.ResolverAddr).To(Equal("0xD3ddcCDD3b25A8a7423B5bEe360a42146eb4Baf3"))
			Expect(record.PointsToAddr).To(Equal("0xa54AEF7fA503E75a03b262A4Cd73037C1774735D"))
		})

	})
})
