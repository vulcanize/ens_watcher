CREATE TABLE public.eth_nodes (
    id integer NOT NULL,
    genesis_block character varying(66),
    network_id numeric,
    eth_node_id character varying(128),
    client_name character varying,
    CONSTRAINT eth_node_uc UNIQUE (genesis_block, network_id, eth_node_id),
    CONSTRAINT nodes_pkey PRIMARY KEY (id)
);