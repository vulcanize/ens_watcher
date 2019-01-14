CREATE FUNCTION public.label_hash(node VARCHAR(66)) RETURNS VARCHAR(66) AS $$
  SELECT label_hash FROM public.domain_records
  WHERE name_hash = node LIMIT 1;
$$ LANGUAGE SQL STABLE;
