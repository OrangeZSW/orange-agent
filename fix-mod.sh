#!/bin/bash

# Go模块优化脚本
# 作用：清理冗余依赖、修复模块状态

# 设置错误退出机制
set -e

# 显示脚本名称和版本
echo "=== Go Mod Optimizer v1.0 ==="

# 检查go命令是否存在
if ! command -v go &> /dev/null; then
  echo "错误：未找到go命令，请确保已安装Go环境"
  exit 1
fi

# 执行模块清理
echo "正在执行 go mod tidy..."
go mod tidy

# 可选扩展功能（可根据需求启用）
# echo "正在生成vendor目录..."
# go mod vendor

# echo "正在下载依赖..."
# go mod download

echo "=== 模块优化完成 ==="