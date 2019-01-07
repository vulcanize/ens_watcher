# ENS Contract Watcher

## Description
A [VulcanizeDB](https://github.com/vulcanize/VulcanizeDB) transformer for watching events related to the [Ethereum Name Service](https://ens.domains) in order to generate domain records.

## Dependencies
 - Go 1.11+
 - Postgres 10
 - Ethereum Node
   - [Go Ethereum](https://ethereum.github.io/go-ethereum/downloads/) (1.8.18+)
   - [Parity 1.8.11+](https://github.com/paritytech/parity/releases)

## Installation
1. Setup Postgres and an Ethereum node - see [VulcanizeDB README](https://github.com/vulcanize/VulcanizeDB/blob/master/README.md).
2. `git clone git@github.com:vulcanize/ens_watcher.git`

3. Install dependencies:
    ```
    make installtools
    ```
4. Execute `createdb vulcanize_ens_watcher`
5. Execute `cd $GOPATH/src/github.com/vulcanize/ens_watcher`
5. Run the migrations: `make migrate HOST_NAME=localhost NAME=vulcanize_ens_watcher PORT=<postgres port, default 5432>`
6. Build:
    ```
    make build
    ```

## Configuration
- To use a local Ethereum node, copy `environments/public.toml.example` to
  `environments/public.toml` and update the `ipcPath` to the local node's IPC filepath:
  - when using geth:
    - The IPC file is called `geth.ipc`.
    - The geth IPC file path is printed to the console when you start geth.
    - The default location is:
      - Mac: `$HOME/Library/Ethereum`
      - Linux: `$HOME/.ethereum`

  - when using parity:
    - The IPC file is called `jsonrpc.ipc`.
    - The default location is:
      - Mac: `$HOME/Library/Application\ Support/io.parity.ethereum/`
      - Linux: `$HOME/.local/share/io.parity.ethereum/`

- See `environments/infura.toml` to configure commands to run against infura, if a local node is unavailable.

## Running the lightSync command
This command syncs VulcanizeDB with the configured Ethereum node.
1. Start node (or configure with Infura)
1. In a separate terminal window:
  `./ens_watcher lightSync --config <config.toml> --starting-block-number <block-number>`
  - where `block-number` is a block number less than or equal to the starting block of the ENS registry contract

## Running the generateRecords command
`generateRecords` starts up a process to watch for ENS-registry and ENS-resolver events and transform the data emitted into Postgres.    

It transforms data collected from these events:
```
// Registry events
event NewOwner(bytes32 indexed node, bytes32 indexed label, address owner);
event Transfer(bytes32 indexed node, address owner);
event NewResolver(bytes32 indexed node, address resolver);
event NewTTL(bytes32 indexed node, uint64 ttl);

// Resolver events
event AddrChanged(bytes32 indexed node, address a);
event ContentChanged(bytes32 indexed node, bytes32 hash);
event NameChanged(bytes32 indexed node, string name);
event ABIChanged(bytes32 indexed node, uint256 indexed contentType);
event PubkeyChanged(bytes32 indexed node, bytes32 x, bytes32 y);
event TextChanged(bytes32 indexed node, string indexedKey, string key);
event MultihashChanged(bytes32 indexed node, bytes hash);
```

Into ENS domain records in a Postgres table of this form:
```postgresql
CREATE TABLE public.domain_records (
  id                    SERIAL PRIMARY KEY,
  block_number          BIGINT NOT NULL,
  name_hash             VARCHAR(66) NOT NULL,
  label_hash            VARCHAR(66) NOT NULL,
  parent_hash           VARCHAR(66) NOT NULL,
  owner_addr            VARCHAR(66) NOT NULL,
  resolver_addr         VARCHAR(66),
  points_to_addr        VARCHAR(66),
  resolved_name         VARCHAR(66),
  content_hash          VARCHAR(66),
  content_type          BIGINT,
  pub_key_x             VARCHAR(66),
  pub_key_y             VARCHAR(66),
  ttl                   BIGINT,
  text_key              TEXT,
  indexed_text_key      TEXT,
  multihash             TEXT,
  UNIQUE (block_number, name_hash)
);
```

This command will need to be run against a lightSynced vDB and a full archival eth node. If a local full archive node is unavailable, see the previous point about running
this command against infura.

`./ens_watcher generateRecords --config <config.toml>`

To watch ENS on ropsten:

`./ens_watcher generateRecords --config <config.toml> --network ropsten`

Note that the database inserts an updated record only every time the record changes state (a new event occurs for that namehash)
This means the sequence of records for a given name_hash will have large block_number gaps where the state of the domain in those gaps has not changed since the previous record. 
This removes a lot of redundancy that would otherwise exist in the database, reducing the storage used and greatly reducing the number of database writes performed during sync.
But, this also affects how queries against the database must be structured to extract certain information, as shown below.

The below returns the most recent record for a given domain namehash
```postgresql
SELECT * FROM public.domain_records
WHERE name_hash = <domain_name_hash>
AND block_number <= (SELECT MAX(block_number)
                    FROM public.domain_records)
ORDER BY block_number
DESC LIMIT 1;             
```

The below returns the record for a given domain namehash at the given block height
```postgresql
SELECT * FROM public.domain_records
WHERE name_hash = <domain_name_hash>
AND block_number <= <block_height>
ORDER BY block_number
DESC LIMIT 1;               
```

The below returns the most recent record for all domains 
```postgresql
SELECT * FROM public.domain_records AS records
LEFT OUTER JOIN public.domain_records AS newer_records ON newer_records.name_hash = records.name_hash
AND newer_records.block_number > records.block_number
WHERE newer_records.block_number IS NULL;
```

The below returns the records for all domains at a given blockheight
```postgresql
WITH records AS (SELECT * FROM public.domain_records
                WHERE records.block_number <= <block_height>)
SELECT * FROM records
LEFT OUTER JOIN public.domain_records AS newer_records ON newer_records.name_hash = records.name_hash
AND newer_records.block_number > records.block_number
WHERE newer_records.block_number IS NULL;
```

### Queries against our domain_records table can be used to answer many different questions:

1. What is the label hash of this domain?

The label_hash for a domain never changes, so we don't care which record we get it from
```postgresql
SELECT label_hash FROM public.domain_records
WHERE name_hash = <domain_name_hash> LIMIT 1;
``` 

2. What is the parent domain of this domain?

To get the parent domain's namehash:
```postgresql
SELECT parent_hash FROM public.domain_records
WHERE name_hash = <domain_name_hash> LIMIT 1;
```

Or if we want the entire, most recent, domain record:
```postgresql
SELECT * FROM public.domain_records
WHERE parent_hash = (SELECT parent_hash 
                    FROM public.domain_records
                    WHERE name_hash = <domain_name_hash> 
                    LIMIT 1)
ORDER BY block_number DESC LIMIT 1;
```

3. What are the subdomains of this domain?

If we want the subdomain namehashes at a given blockheight:
```postgresql
SELECT DISTINCT name_hash
FROM public.domain_records
WHERE parent_hash = <domain_name_hash>
AND block_number <= <block_height>;
```

If we want all the subdomain records that have ever existed up until the given block height:
```postgresql
SELECT * 
FROM public.domain_records
WHERE parent_hash = <domain_name_hash>
AND block_number <= <block_height>;
```

If we want only the most recent record for each subdomain:
```postgresql
WITH recent_records AS (SELECT * FROM public.domain_records AS records
                        LEFT OUTER JOIN public.domain_records AS newer_records ON newer_records.name_hash = records.name_hash
                        AND newer_records.block_number > records.block_number
                        WHERE newer_records.block_number IS NULL)
SELECT *
FROM recent_records
WHERE parent_hash = <domain_name_hash>;
```

If we want the record for each subdomain at a given blockheight:
```postgresql
WITH records_at AS (WITH records AS (SELECT * FROM public.domain_records
                                    WHERE records.block_number <= <block_height>)
                    SELECT * FROM records
                    LEFT OUTER JOIN public.domain_records AS newer_records ON newer_records.name_hash = records.name_hash
                    AND newer_records.block_number > records.block_number
                    WHERE newer_records.block_number IS NULL)
SELECT *
FROM records_at
WHERE parent_hash = <domain_name_hash>;
```

4. What domains does this address own?

To get the domains the address currently owns:
```postgresql
WITH recent_records AS (SELECT * FROM public.domain_records AS records
                        LEFT OUTER JOIN public.domain_records AS newer_records ON newer_records.name_hash = records.name_hash
                        AND newer_records.block_number > records.block_number
                        WHERE newer_records.block_number IS NULL)
SELECT name_hash
FROM recent_records
WHERE owner_addr = <address>;
```

To check what domains the address owned at a given block height:
```postgresql
WITH records_at AS (WITH records AS (SELECT * FROM public.domain_records
                                    WHERE records.block_number <= <block_height>)
                    SELECT * FROM records
                    LEFT OUTER JOIN public.domain_records AS newer_records ON newer_records.name_hash = records.name_hash
                    AND newer_records.block_number > records.block_number
                    WHERE newer_records.block_number IS NULL)
SELECT name_hash
FROM records_at
WHERE owner_addr = <address>;
```

5. What names/domains currently point to this address?

To get the namehashes that resolve to a given address:
```postgresql
WITH recent_records AS (SELECT * FROM public.domain_records AS records
                        LEFT OUTER JOIN public.domain_records AS newer_records ON newer_records.name_hash = records.name_hash
                        AND newer_records.block_number > records.block_number
                        WHERE newer_records.block_number IS NULL)
SELECT name_hash
FROM recent_records
WHERE points_to_addr = <address>;
```

Can also get the resolved names (if it exists) of the domains which resolve to a given address:
```postgresql
WITH recent_records AS (SELECT * FROM public.domain_records AS records
                        LEFT OUTER JOIN public.domain_records AS newer_records ON newer_records.name_hash = records.name_hash
                        AND newer_records.block_number > records.block_number
                        WHERE newer_records.block_number IS NULL)
SELECT resolved_name
FROM recent_records
WHERE points_to_addr = <address>;
```

6. What domains are using this resolver?

To get which domains are currently using a given resolver:
```postgresql
WITH recent_records AS (SELECT * FROM public.domain_records AS records
                        LEFT OUTER JOIN public.domain_records AS newer_records ON newer_records.name_hash = records.name_hash
                        AND newer_records.block_number > records.block_number
                        WHERE newer_records.block_number IS NULL)
SELECT *
FROM recent_records
WHERE resolver_addr = <address>;
```

To get which domains were using a given resolver at a given block height:
```postgresql
WITH records_at AS (WITH records AS (SELECT * FROM public.domain_records
                                    WHERE records.block_number <= <block_height>)
                    SELECT * FROM records
                    LEFT OUTER JOIN public.domain_records AS newer_records ON newer_records.name_hash = records.name_hash
                    AND newer_records.block_number > records.block_number
                    WHERE newer_records.block_number IS NULL)
SELECT *
FROM records_at
WHERE resolver_addr = <address>;
```

## Running the tests
1. Install dependencies `make installtools`
2. Execute `createdb vulcanize_ens_watcher_private`
3. Execute `cd $GOPATH/src/github.com/vulcanize/ens_watcher`
4. Run the migrations `make migrate HOST_NAME=localhost NAME=vulcanize_ens_watcher_private PORT=<postgres port, default 5432>`
5. Run the tests `make test`