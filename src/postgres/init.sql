-- CREATE USER rw_user WITH PASSWORD '';
CREATE USER ro_user WITH PASSWORD 'ro_password';

-- GRANT ALL PRIVILEGES ON DATABASE postgres TO rw_user;
GRANT CONNECT ON DATABASE postgres TO ro_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO ro_user;

-- GRANT SELECT ON appuser TO ro_user;
