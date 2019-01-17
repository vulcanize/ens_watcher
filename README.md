# ENS Contract Watcher
[![Build Status](https://travis-ci.org/vulcanize/ens_watcher.svg?branch=master)](https://travis-ci.org/vulcanize/ens_watcher)

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
event ContenthashChanged(bytes32 indexed node, bytes hash);
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
  content_              VARCHAR(66),
  content_type          TEXT,
  pub_key_x             VARCHAR(66),
  pub_key_y             VARCHAR(66),
  ttl                   TEXT,
  text_key              TEXT,
  indexed_text_key      TEXT,
  multihash             TEXT,
  contenthash           TEXT,
  UNIQUE (block_number, name_hash)
);
```

This command will need to be run against a lightSynced vDB and a fast/full eth node. If a local full/fst node is unavailable, see the previous point about running
this command against infura.

`./ens_watcher generateRecords --config <config.toml>`

To watch ENS on ropsten:

`./ens_watcher generateRecords --config <config.toml> --network ropsten`

Note that the database inserts an updated record only every time the record changes state (a new event occurs for that namehash)
This means the sequence of records for a given name_hash will have large block_number gaps where the state of the domain in those gaps has not changed since the previous record. 
This removes a lot of redundancy that would otherwise exist in the database, reducing the storage used and greatly reducing the number of database writes performed during sync.
But, this also affects how queries against the database must be structured to extract certain information, to help alleviate this issue we have created some stored Postgres functions that 
can be easily accessed over a GraphQL api to answer certain questions:


### Queries against our domain_records table can be used to answer many different questions:

1. What is the label hash of this domain?

The label_hash for a domain never changes, so we don't care which record we get it from
```postgresql
CREATE FUNCTION public.label_hash(node VARCHAR(66)) RETURNS VARCHAR(66) AS $$
  SELECT label_hash FROM public.domain_records
  WHERE name_hash = node LIMIT 1;
$$ LANGUAGE SQL STABLE;
``` 

2. What is the parent domain of this domain?

To get the parent domain's namehash:
```postgresql
CREATE FUNCTION public.parent_hash(node VARCHAR(66)) RETURNS VARCHAR(66) AS $$
  SELECT parent_hash FROM public.domain_records
  WHERE name_hash = node LIMIT 1;
$$ LANGUAGE SQL STABLE;
```

Or if we want the entire, most recent, domain record:
```postgresql
CREATE FUNCTION public.parent_record(node VARCHAR(66)) RETURNS public.domain_records AS $$
  SELECT * FROM public.domain_records
  WHERE parent_hash = (SELECT parent_hash
                    FROM public.domain_records
                    WHERE name_hash = node
                    LIMIT 1)
  ORDER BY block_number DESC LIMIT 1;
$$ LANGUAGE SQL STABLE;
```

3. What are the subdomains of this domain?

If we want the current subdomain namehashes of a node
```postgresql
CREATE FUNCTION public.subdomain_hashes(node VARCHAR(66)) RETURNS SETOF VARCHAR(66) AS $$
  SELECT DISTINCT name_hash FROM public.domain_records
  WHERE parent_hash = node;
$$ LANGUAGE SQL STABLE;
```

If we want the subdomain namehashes of a node at a given blockheight
```postgresql
CREATE FUNCTION public.subdomain_hashes_at(node VARCHAR(66), block BIGINT) RETURNS SETOF VARCHAR(66) AS $$
  SELECT DISTINCT name_hash FROM public.domain_records
  WHERE parent_hash = node
  AND block_number <= block;
$$ LANGUAGE SQL STABLE;
```

If we want a set of the current records for each subdomain of the given node:
```postgresql
CREATE FUNCTION public.subdomain_records(node VARCHAR(66)) RETURNS SETOF public.domain_records AS $$
  SELECT
    domain_records.*
  FROM
    (SELECT
      name_hash, MAX(block_number) AS block_number
     FROM
      domain_records
     GROUP BY
      name_hash) AS latest_records
  INNER JOIN
    domain_records
  ON
    domain_records.name_hash = latest_records.name_hash AND
    domain_records.block_number = latest_records.block_number
  WHERE
    domain_records.parent_hash = node;
$$ LANGUAGE SQL STABLE;
```

4. What domains does this address currently own?

If we want a set of domain records currently owned by a given address
```postgresql
CREATE FUNCTION public.domains_owned_by(addr VARCHAR(66)) RETURNS SETOF public.domain_records AS $$
  SELECT
    domain_records.*
  FROM
    (SELECT
      name_hash, MAX(block_number) AS block_number
     FROM
      domain_records
     GROUP BY
      name_hash) AS latest_records
  INNER JOIN
    domain_records
  ON
    domain_records.name_hash = latest_records.name_hash AND
    domain_records.block_number = latest_records.block_number
  WHERE
    domain_records.owner_addr = addr;
$$ LANGUAGE SQL STABLE;
```

5. What domains resolve to this address?

To get a set of domain records that currently resolve to a given address:
```postgresql
CREATE FUNCTION public.domains_resolving_to_addr(addr VARCHAR(66)) RETURNS SETOF public.domain_records AS $$
  SELECT
    domain_records.*
  FROM
    (SELECT
      name_hash, MAX(block_number) AS block_number
     FROM
      domain_records
     GROUP BY
      name_hash) AS latest_records
  INNER JOIN
    domain_records
  ON
    domain_records.name_hash = latest_records.name_hash AND
    domain_records.block_number = latest_records.block_number
  WHERE
    domain_records.points_to_addr = addr;
$$ LANGUAGE SQL STABLE;
```

6. What domains resolve to this name?

To get a set of domain records that currently resolve to a given name:
```postgresql
CREATE FUNCTION public.domains_resolving_to_name(name_str TEXT) RETURNS SETOF public.domain_records AS $$
  SELECT
    domain_records.*
  FROM
    (SELECT
      name_hash, MAX(block_number) AS block_number
     FROM
      domain_records
     GROUP BY
      name_hash) AS latest_records
  INNER JOIN
    domain_records
  ON
    domain_records.name_hash = latest_records.name_hash AND
    domain_records.block_number = latest_records.block_number
  WHERE
    domain_records.resolved_name = name_str;
$$ LANGUAGE SQL STABLE;
```

7. What domains are using this resolver?

To get a set of domain records that are currently using a given resolver:
```postgresql
CREATE FUNCTION public.domains_resolved_by(resolver VARCHAR(66)) RETURNS SETOF public.domain_records AS $$
  SELECT
    domain_records.*
  FROM
    (SELECT
      name_hash, MAX(block_number) AS block_number
     FROM
      domain_records
     GROUP BY
      name_hash) AS latest_records
  INNER JOIN
    domain_records
  ON
    domain_records.name_hash = latest_records.name_hash AND
    domain_records.block_number = latest_records.block_number
  WHERE
    domain_records.resolver_addr = resolver;
$$ LANGUAGE SQL STABLE;
```

## Setting up GraphQL API using Postgraphile
Install postgraphile and run the following command to access the GraphQL API:
`npx postgraphile -c postgres:///ens_watcher --schema public`

## Running the tests
1. Install dependencies `make installtools`
2. Execute `createdb vulcanize_ens_watcher_private`
3. Execute `cd $GOPATH/src/github.com/vulcanize/ens_watcher`
4. Run the migrations `make migrate HOST_NAME=localhost NAME=vulcanize_ens_watcher_private PORT=<postgres port, default 5432>`
5. Run the tests `make test`