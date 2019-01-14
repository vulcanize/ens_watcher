CREATE TABLE public.eth_nodes (
    id integer NOT NULL,
    genesis_block character varying(66),
    network_id numeric,
    eth_node_id character varying(128),
    client_name character varying,
    CONSTRAINT eth_node_uc UNIQUE (genesis_block, network_id, eth_node_id),
    CONSTRAINT nodes_pkey PRIMARY KEY (id)
);

CREATE SEQUENCE public.nodes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.nodes_id_seq OWNED BY public.eth_nodes.id;

ALTER TABLE ONLY public.eth_nodes ALTER COLUMN id SET DEFAULT nextval('public.nodes_id_seq'::regclass);

