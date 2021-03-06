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
