-- CREATE USER rw_user WITH PASSWORD '';
-- GRANT ALL PRIVILEGES ON DATABASE postgres TO rw_user;

CREATE ROLE ro_user WITH REPLICATION LOGIN PASSWORD 'ro_password';

GRANT USAGE ON SCHEMA public TO ro_user;

GRANT CONNECT ON DATABASE postgres TO ro_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO ro_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO ro_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON sequences TO ro_user;

GRANT pg_read_all_data TO ro_user;

-- GRANT SELECT ON appuser TO ro_user;

-- docker exec -it <container_id> bash
-- root@<container_id>:/# psql -U rw_user

-- SELECT *
--   FROM information_schema.role_table_grants
--  WHERE grantee = 'ro_user';
