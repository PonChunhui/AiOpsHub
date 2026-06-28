#!/bin/bash

# AiOpsHub Backend启动脚本

set -e

# 检查配置文件
if [ ! -f "config/config.yaml" ]; then
    echo "Config file not found, copying example..."
    cp config/config.yaml.example config/config.yaml
fi

# 启动API Server
echo "Starting API Server..."
./bin/api-server