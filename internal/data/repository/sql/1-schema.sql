DROP DATABASE IF EXISTS rotator;
CREATE DATABASE rotator;
CREATE USER igor WITH encrypted password 'igor';
GRANT ALL PRIVILEGES ON DATABASE rotator to igor;