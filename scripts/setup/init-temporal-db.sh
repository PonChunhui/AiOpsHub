#!/bin/bash
set -e

# 创建Temporal Server需要的数据库
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE temporal;
    CREATE DATABASE temporal_visibility;
    GRANT ALL PRIVILEGES ON DATABASE temporal TO aiops;
    GRANT ALL PRIVILEGES ON DATABASE temporal_visibility TO aiops;
EOSQL

echo "Temporal databases created successfully"