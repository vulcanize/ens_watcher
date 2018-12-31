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

package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/vulcanize/vulcanizedb/libraries/shared"
	"github.com/vulcanize/vulcanizedb/utils"

	"github.com/vulcanize/ens_watcher/transformer"
)

// generateRecordsCmd represents the generateRecords command
var generateRecordsCmd = &cobra.Command{
	Use:   "generateRecords",
	Short: "Generate ENS domain records",
	Long: `Generates ENS domain records. Must be run with a lightSynced vDB.
Watches events at the ENS-registry contract and ENS-resolver contracts to 
produce domain records:
	CREATE TABLE ens.domain_records (
 	 id                    SERIAL PRIMARY KEY,
 	 block_number          BIGINT NOT NULL,
 	 name_hash             VARCHAR(66) NOT NULL,
  	 label_hash            VARCHAR(66) NOT NULL,
  	 parent_hash           VARCHAR(66) NOT NULL,
  	 owner_addr            VARCHAR(66) NOT NULL,
  	 resolver_addr         VARCHAR(66),
  	 points_to_addr        VARCHAR(66),
  	 resolved_name         VARCHAR(66),
  	 UNIQUE (block_number, name_hash)
	);

Usage:
Expects ethereum node to be running and requires a .toml config:
  [database]
  name = "vulcanize_public"
  hostname = "localhost"
  port = 5432
  [client]
  ipcPath = "/Users/user/Library/Ethereum/geth.ipc"

Expects that lightSync has run or is running and was started below the ENS-registry contracts first block
./ens_watcher lightSync --starting-block-number 3327417 --config public.toml

Run the generateRecords command
./ens_watcher generateRecords --config public.toml

The default network is mainnet which uses ENS-registry 0x314159265dD8dbb310642f98f50C066173C1259b
To run on ropsten using ENS-registry 0x112234455C3a32FD11230C42E7Bccd4A84e02010, run
./ens_watcher generateRecords --network ropsten --config public.toml
`,
	Run: func(cmd *cobra.Command, args []string) {
		generateRecords()
	},
}

func generateRecords() {
	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()
	blockChain := getBlockChain()
	db := utils.LoadPostgres(databaseConfig, blockChain.Node())

	t, err := transformer.NewTransformer(network, blockChain, &db)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Init()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to initialized transformer\r\nerr: %v\r\n", err))
	}

	w := shared.Watcher{}
	w.AddTransformer(t)

	for range ticker.C {
		w.Execute()
	}
}

func init() {
	rootCmd.AddCommand(generateRecordsCmd)
	generateRecordsCmd.Flags().StringVarP(&network, "network", "n", "", "defualt is mainnet; set to ropsten to run on testnet")
}
