#!/bin/bash

# 设置错误时退出
set -e

echo "========== 开始编译 =========="

# 编译
go build -o orange-agent

# 判断编译是否成功
if [ $? -eq 0 ]; then
    echo "✅ 编译成功"
else
    echo "❌ 编译失败"
    exit 1
fi

# 修改权限
chmod +x orange-agent

echo "========== 停止旧进程 =========="

# 查询运行中的PID
PID=$(ps -ef | grep orange-agent | grep -v grep | grep -v "bash" | awk '{print $2}')

# 停止旧进程
if [ -n "$PID" ]; then
    echo "发现运行中的进程 PID: $PID"
    
    # 先尝试优雅关闭
    kill $PID
    echo "发送停止信号，等待进程退出..."
    
    # 等待最多10秒
    for i in {1..10}; do
        if ! kill -0 $PID 2>/dev/null; then
            echo "进程已优雅退出"
            break
        fi
        sleep 1
    done
    
    # 如果还在运行，强制杀死
    if kill -0 $PID 2>/dev/null; then
        echo "进程未响应，强制停止"
        kill -9 $PID
        echo "进程已强制停止"
    fi
else
    echo "未发现运行中的程序"
fi

echo "========== 启动新进程 =========="

# 后台运行（可选）
# nohup ./orange-agent > app.log 2>&1 &

# 前台运行
./orange-agent

# 判断运行是否成功
if [ $? -eq 0 ]; then
    echo "✅ 运行成功"
else
    echo "❌ 运行失败"
    exit 1
fi