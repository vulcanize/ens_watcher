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
  text_key              TEXT,
  indexed_text_key      TEXT,
  multihash             TEXT,
  UNIQUE (block_number, name_hash)
);