#!/bin/bash

# 编译脚本
echo "开始编译 orange-agent..."
go build -o orange-agent

if [ $? -eq 0 ]; then
    echo "✅ 编译成功"
    chmod +x orange-agent
else
    echo "❌ 编译失败"
    exit 1
fi