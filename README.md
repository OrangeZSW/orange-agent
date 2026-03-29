# Orange Agent - 智能Telegram代理机器人

Orange Agent 是一个基于 Go 语言开发的 Telegram 智能代理机器人，集成了 LangChain 框架，支持多 AI 模型切换和工具调用功能。

## 功能特性

- 🤖 **多AI代理支持**：支持配置多个AI代理（如OpenAI、本地模型等）
- 🔄 **模型热切换**：支持运行时切换不同的AI模型
- 🛠️ **工具调用**：支持文件操作、时间查询等工具调用
- 🌐 **联网搜索**：支持搜索引擎查询和网页内容抓取
- 💾 **对话记忆**：保存用户对话历史，提供上下文感知
- 📊 **使用统计**：记录AI调用次数和Token使用情况
- 🔧 **配置管理**：通过配置文件灵活管理代理和模型
- 🚀 **代理支持**：支持HTTP代理连接Telegram API

## 项目结构

```
orange-agent/
├── cmd/
│   └── agent/
│       └── main.go              # 主入口
│
├── internal/                     # 内部包（不对外暴露）
│   ├── domain/                   # 领域模型（纯数据，无依赖）
│   │   ├── agent_config.go
│   │   ├── memory.go
│   │   └── user.go
│   │
│   ├── repository/               # 数据访问层（依赖 domain）
│   │   ├── interface.go         # 定义接口
│   │   ├── mysql/
│   │   │   ├── agent_config.go
│   │   │   ├── memory.go
│   │   │   └── user.go
│   │   └── factory.go
│   │
│   ├── service/                  # 业务逻辑层（依赖 domain, repository）
│   │   ├── agent/
│   │   │   ├── service.go
│   │   │   └── tools.go
│   │   ├── file/
│   │   │   └── service.go
│   │   └── system/
│   │       └── service.go
│   │
│   ├── handler/                  # 处理器层（依赖 service）
│   │   ├── telegram/
│   │   │   ├── bot.go
│   │   │   ├── command.go
│   │   │   └── text.go
│   │   └── lanchain/
│   │       ├── handler.go
│   │       └── execute_tool.go
│   │
│   └── pkg/                      # 内部共享包
│       ├── logger/               # 日志工具
│       │   └── logger.go
│       ├── utils/                # 通用工具函数
│       │   ├── file.go
│       │   └── map.go
│       └── config/               # 配置管理
│           └── config.go
│
├── pkg/                          # 可公开的包（如果有需要）
│   └── ...
│
├── config.yaml
├── go.mod
└── go.sum
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
- `file_delete` - 删除文件
- `file_rename` - 重命名文件
- `file_copy` - 复制文件
- `file_search` - 搜索文件内容

### Git 工具

- `git_commit` - 提交代码
- `git_push` - 推送到远程仓库
- `git_diff` - 查看代码差异

### 系统工具

- `curr_time` - 获取当前时间
- `build_tools` - 构建项目
- `project_reboot` - 重启项目
- `log_view` - 查看日志
- `env_manage` - 管理环境变量
- `test_run` - 运行测试
- `dependency_check` - 检查依赖
- `performance_monitor` - 性能监控
- `api_tester` - API接口测试

### 🌐 联网搜索工具（新增）

- `web_search` - 联网搜索功能

**功能特性：**

- 🔍 **搜索引擎搜索**：支持 DuckDuckGo、Google、Bing 搜索
- 📄 **网页内容抓取**：获取网页正文内容
- 🚀 **无需 API Key**：DuckDuckGo 搜索无需配置 API 密钥
- 🌍 **多语言支持**：支持中英文搜索

**使用示例：**

```
# 搜索关键词
搜索 "Go语言教程"

# 抓取网页内容
抓取 https://example.com 的内容

# 使用不同搜索引擎
用 Google 搜索 "人工智能最新进展"
```

**参数说明：**

- `query`: 搜索关键词或URL
- `search_type`: 搜索类型
  - `search`: 搜索引擎搜索
  - `fetch`: 抓取网页内容
- `engine`: 搜索引擎（可选，默认 duckduckgo）
  - `duckduckgo`: DuckDuckGo（推荐，免费）
  - `google`: Google（需 API Key）
  - `bing`: Bing（需 API Key）
- `num_results`: 返回结果数量（默认5条，最多10条）

### Agent 管理工具

- `agent_add` - 添加 Agent 配置
- `agent_remove` - 删除 Agent
- `agent_list` - 列出所有 Agent
- `agent_update` - 更新 Agent 配置
- `agent_test` - 测试 Agent 连接

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
4. **联网搜索失败**

   - 检查网络连接
   - 确认可以访问 DuckDuckGo API
   - 如需使用 Google/Bing，需配置相应 API Key

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

## 更新日志

### v1.1.0 (最新)

- ✨ 新增联网搜索功能
  - 支持 DuckDuckGo、Google、Bing 搜索引擎
  - 支持网页内容抓取
  - 无需 API Key 即可使用 DuckDuckGo 搜索
- 🐛 修复若干已知问题
- 📝 更新文档

### v1.0.0

- 🎉 初始版本发布
- 🤖 多 AI 代理支持
- 🛠️ 工具调用系统
- 💾 对话记忆功能

---

如有问题或建议，请通过 Issue 或 Pull Request 提交。
