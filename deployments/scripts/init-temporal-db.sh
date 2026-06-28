-- Initialize Temporal databases
CREATE DATABASE IF NOT EXISTS temporal;
CREATE DATABASE IF NOT EXISTS temporal_visibility;

-- Initialize AiOpsHub database
CREATE DATABASE IF NOT EXISTS aiopsdb;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE temporal TO aiops;
GRANT ALL PRIVILEGES ON DATABASE temporal_visibility TO aiops;
GRANT ALL PRIVILEGES ON DATABASE aiopsdb TO aiops;