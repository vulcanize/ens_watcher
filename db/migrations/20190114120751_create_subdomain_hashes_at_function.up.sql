CREATE FUNCTION public.subdomain_hashes_at(node VARCHAR(66), block BIGINT) RETURNS SETOF VARCHAR(66) AS $$
  SELECT DISTINCT name_hash FROM public.domain_records
  WHERE parent_hash = node
  AND block_number <= block;
$$ LANGUAGE SQL STABLE;
