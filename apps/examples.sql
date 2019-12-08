CREATE TABLE configs (
  id SERIAL,
  data JSONB
);

-- Make sure Name is unique
CREATE UNIQUE INDEX configs_name_idx ON configs(((data->>'name')::name));

INSERT INTO configs (data) VALUES ('{"name": "foo","metadata": {"monitoring": {"enabled": "true"},"limits": {"cpu": {"enabled": "false","value": "300m"}}}}');
INSERT INTO configs (data) VALUES ('{"name": "bar","metadata": {"monitoring": {"enabled": "false"},"limits": {"cpu": {"enabled": "true","value": "200m"}}}}');

CREATE INDEX idxmon ON configs ((data->>'monitoring'));

-- SELECT count(*) FROM configs WHERE data->'monitoring'->'enabled' ? 'true' AND data->'cpu'->'enabled' ? 'true';

CREATE ROLE demo superuser createdb login password 'demo';
