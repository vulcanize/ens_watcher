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

//namehash('') = 0x0000000000000000000000000000000000000000000000000000000000000000
//namehash('eth') = 0x93cdeb708b7545dc668eb9280176169d1c33cfd8ed6f04690a0bcc88a93fc4ae
//namehash('foo.eth') = 0xde9b09fd7c5f901e23a3f19fecc54828e9c848539801e86591bd9801b019f84f

package utils_test

import (
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/ens_watcher/utils"
)

var _ = Describe("Utils", func() {
	Describe("NameHash", func() {
		It("Returns the namehash for the input string", func() {
			hash := utils.NameHash("")
			Expect(hash).To(Equal(common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")))

			hash = utils.NameHash("eth")
			Expect(hash).To(Equal(common.HexToHash("0x93cdeb708b7545dc668eb9280176169d1c33cfd8ed6f04690a0bcc88a93fc4ae")))

			hash = utils.NameHash("foo.eth")
			Expect(hash).To(Equal(common.HexToHash("0xde9b09fd7c5f901e23a3f19fecc54828e9c848539801e86591bd9801b019f84f")))
		})
	})
})
