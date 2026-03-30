#!/bin/bash
# 将 TERM 配置添加到 .bashrc 文件末尾（如果尚未存在）
if ! grep -q 'export TERM=xterm-256color' ~/.bashrc; then
    echo 'export TERM=xterm-256color' >> ~/.bashrc
    echo "已将 'export TERM=xterm-256color' 添加到 ~/.bashrc"
else
    echo "配置 'export TERM=xterm-256color' 已存在于 ~/.bashrc"
fi