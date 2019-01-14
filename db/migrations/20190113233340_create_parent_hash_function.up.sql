CREATE FUNCTION public.parent_hash(node VARCHAR(66)) RETURNS VARCHAR(66) AS $$
  SELECT parent_hash FROM public.domain_records
  WHERE name_hash = node LIMIT 1;
$$ LANGUAGE SQL STABLE;