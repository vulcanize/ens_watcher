CREATE FUNCTION public.parent_record(node VARCHAR(66)) RETURNS public.domain_records AS $$
  SELECT * FROM public.domain_records
  WHERE parent_hash = (SELECT parent_hash
                    FROM public.domain_records
                    WHERE name_hash = node
                    LIMIT 1)
  ORDER BY block_number DESC LIMIT 1;
$$ LANGUAGE SQL STABLE;
