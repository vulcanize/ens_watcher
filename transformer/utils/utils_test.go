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

package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/ens_watcher/transformer/utils"
)

var _ = Describe("Utils", func() {
	Describe("CreateSubnode", func() {
		It("Creates a subnode hash from a given parent and label hash", func() {
			node := "0x583506c12610038ce46126390030389f0555a9525aa5be5a2dd2bf08e316bb00"
			label := "0xbe71a413dd3e859f6c2f69eebb2d3bfcdefc8884a5086d0c8c8a7715b3e328c1"
			subnode := utils.CreateSubnode(node, label)
			Expect(subnode).To(Equal("0xb4664b154f4dd9abf5bb27d6e3ff12181d6e37b0606b4ff61ff8796e6e29a2e4"))
		})
	})
})
