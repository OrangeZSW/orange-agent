# Orange Agent - 智能Telegram代理机器人

Orange Agent 是一个基于 Go 语言开发的 Telegram 智能代理机器人，集成了 LangChain 框架，支持多 AI 模型切换和工具调用功能。

## 功能特性

- 🤖 **多AI代理支持**：支持配置多个AI代理（如OpenAI、本地模型等）
- 🔄 **模型热切换**：支持运行时切换不同的AI模型
- 🛠️ **工具调用**：支持文件操作、时间查询等工具调用
- 💾 **对话记忆**：保存用户对话历史，提供上下文感知
- 📊 **使用统计**：记录AI调用次数和Token使用情况
- 🔧 **配置管理**：通过配置文件灵活管理代理和模型
- 🚀 **代理支持**：支持HTTP代理连接Telegram API

## 项目结构

```
orange-agent/
├── common/                  # 通用工具和基础结构
├── config/                  # 配置管理
│   ├── config.go
│   └── config.yaml          # 配置文件模板
├── domain/                  # 领域模型定义
├── lanchain/                # LangChain集成
│   ├── answer.go            # AI回答处理
│   └── base.go              # LangChain基础配置
├── mysql/                   # 数据库操作
├── telegram/                # Telegram Bot集成
│   ├── bot.go               # Bot主逻辑
│   ├── command.go           # 命令处理器
│   └── text.go              # 文本消息处理器
├── tools/                   # 工具定义
├── utils/                   # 工具函数
├── main.go                  # 程序入口
├── go.mod                   # Go模块定义
└── start.sh                 # 启动脚本
```

## 快速开始

### 1. 环境要求

- Go 1.25.6 或更高版本
- MySQL 5.7+ 数据库
- Telegram Bot Token

### 2. 配置数据库

创建数据库：
```sql
CREATE DATABASE `orange-agent` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. 配置文件

复制并修改配置文件：
```bash
cp config.yaml.example config.yaml
```

编辑 `config.yaml`：
```yaml
telegram:
  bot_token: "YOUR_TELEGRAM_BOT_TOKEN"  # 从 @BotFather 获取
  proxy: "http://127.0.0.1:7897"       # 代理地址（可选）
  promete: "promete/telegram.text"     # 系统提示词文件

database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  username: "root"
  password: "your_password"
  database: "orange-agent"

logger:
  level: "info"
  file_path: "./log"
  file_name: "orange-agent.log"
```

### 4. 安装依赖

```bash
go mod download
```

### 5. 编译运行

```bash
# 编译
go build -o orange-agent

# 运行
./orange-agent
```

或者使用启动脚本：
```bash
chmod +x start.sh
./start.sh
```

## Telegram 命令使用

### 基础命令

- `/start` - 显示欢迎信息
- `/help` - 显示帮助信息
- `/agents` - 显示所有可用的AI代理
- `/model` - 显示当前使用的模型

### 管理命令

- `/addAgent <agent_name> <base_url> <token>` - 添加新的AI代理
  - 示例：`/addAgent OpenAI https://api.openai.com/v1 sk-xxx`
  
- `/addModel <agent_id> <model_name>` - 为代理添加模型
  - 示例：`/addModel 1 gpt-4-turbo`
  
- `/switch <agent_id> <model_index>` - 切换到指定代理和模型
  - 示例：`/switch 1 2`

### 交互使用

直接发送消息即可与AI对话，系统会自动处理工具调用和上下文记忆。

## 工具系统

Orange Agent 集成了以下工具：

### 文件工具
- `file_read` - 读取文件内容
- `file_list` - 列出目录中的文件
- `file_write` - 写入文件内容

### 其他工具
- `curr_time` - 获取当前时间

## 配置说明

### AI 代理配置

通过 Telegram 命令或数据库直接配置 AI 代理：

```sql
-- 示例：添加 OpenAI 代理
INSERT INTO agent_configs (name, base_url, token, models) 
VALUES ('OpenAI', 'https://api.openai.com/v1', 'sk-xxx', '["gpt-3.5-turbo", "gpt-4"]');
```

### 模型切换

每个用户可以独立选择不同的AI模型，切换记录会保存在数据库中。

## 数据库表结构

项目使用以下核心表：

1. `users` - 用户信息
2. `agent_configs` - AI代理配置
3. `memories` - 对话记忆
4. `agent_call_records` - AI调用记录
5. `agent_configs` - 代理配置

## 开发指南

### 添加新工具

1. 在 `tools/` 目录下创建新的工具包
2. 实现工具函数
3. 在 `tools/tools.go` 中注册工具
4. 重新编译运行

### 扩展AI代理

1. 在 `lanchain/base.go` 中扩展代理类型
2. 实现相应的配置加载逻辑
3. 通过 `/addAgent` 命令添加新代理

## 日志系统

日志系统支持多级别输出：
- `debug` - 调试信息
- `info` - 常规信息
- `warn` - 警告信息
- `error` - 错误信息

日志文件保存在 `./log/orange-agent.log`，支持轮转和压缩。

## 故障排除

### 常见问题

1. **无法连接到Telegram**
   - 检查网络连接和代理设置
   - 确认Bot Token是否正确

2. **数据库连接失败**
   - 检查数据库配置
   - 确认MySQL服务正在运行

3. **AI调用失败**
   - 检查代理配置和API密钥
   - 确认网络可以访问AI服务

### 查看日志

```bash
tail -f log/orange-agent.log
```

## 许可证

本项目基于 MIT 许可证开源。

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送分支
5. 创建Pull Request

## 相关技术

- [Go](https://golang.org/) - 编程语言
- [LangChain Go](https://github.com/tmc/langchaingo) - AI框架
- [TeleBot](https://github.com/tucnak/telebot) - Telegram Bot框架
- [GORM](https://gorm.io/) - ORM框架
- [Viper](https://github.com/spf13/viper) - 配置管理

---

如有问题或建议，请通过 Issue 或 Pull Request 提交。