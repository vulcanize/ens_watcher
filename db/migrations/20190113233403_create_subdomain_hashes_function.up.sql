CREATE FUNCTION public.subdomain_hashes(node VARCHAR(66)) RETURNS SETOF VARCHAR(66) AS $$
  SELECT DISTINCT name_hash FROM public.domain_records
  WHERE parent_hash = node;
$$ LANGUAGE SQL STABLE;
