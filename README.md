# Orange Agent - 智能 Telegram 代理机器人

Orange Agent 是一个基于 Go 语言开发的 Telegram 智能代理机器人，集成了 LangChain 框架，支持多 AI 模型切换、工具调用和 Telegram 消息通知功能。

## 功能特性

- 🤖 **多 AI 代理支持**：支持配置多个 AI 代理（如 OpenAI、本地模型等）
- 🔄 **模型热切换**：支持运行时切换不同的 AI 模型
- 🛠️ **工具调用**：支持文件操作、时间查询、Git 操作、数据库操作等工具调用
- 📚 **技能系统**：内置多种技能模板，支持快速查询工具使用说明
- 🌐 **联网搜索**：支持搜索引擎查询和网页内容抓取
- 💾 **对话记忆**：保存用户对话历史，提供上下文感知
- 📊 **使用统计**：记录 AI 调用次数和 Token 使用情况
- 🔔 **Telegram 消息通知**：支持发送系统通知和任务状态更新
- 🔧 **配置管理**：通过配置文件灵活管理代理和模型
- 🚀 **代理支持**：支持 HTTP 代理连接 Telegram API
- 🎯 **Agent 单例模式**：优化 Agent 管理，提升性能和稳定性

## 项目结构

```
orange-agent/
├── agent/                        # Agent 核心模块
│   ├── agent.go                  # Agent 主逻辑
│   ├── client/                   # Agent 客户端
│   │   └── client.go
│   ├── interfaces/               # Agent 接口定义
│   │   └── interfaces.go
│   ├── manager/                  # Agent 管理器
│   │   └── manager.go
│   ├── task/                     # 任务处理模块
│   │   ├── analyzer/             # 任务分析器
│   │   ├── context/              # 任务上下文
│   │   ├── executor/             # 任务执行器
│   │   ├── orchestrator/         # 任务编排器
│   │   └── summarizer/           # 任务总结器
│   └── tools/                    # Agent 工具集
│       ├── agent/                # Agent 管理工具
│       ├── file/                 # 文件操作工具
│       ├── git/                  # Git 操作工具
│       ├── database/             # 数据库操作工具
│       └── system/               # 系统工具
├── common/                       # 通用工具和基础类型
│   ├── base_tool.go
│   └── file_node.go
├── config/                       # 配置管理
│   ├── config.go
│   └── config.yaml
├── domain/                       # 领域模型（纯数据，无依赖）
│   ├── agentConfig.go
│   ├── config.go
│   ├── memory.go
│   ├── task.go
│   └── user.go
├── log/                          # 日志目录
│   └── orange-agent.log
├── promet/                       # 系统提示词
│   └── telegram.md
├── repository/                   # 数据访问层
│   ├── db/                       # 数据库实现
│   │   └── mysql.go
│   ├── gorm/                     # GORM 模型
│   │   ├── SubTask.go
│   │   ├── Task.go
│   │   ├── agentCallRecord.go
│   │   ├── agentConfig.go
│   │   ├── memory.go
│   │   ├── task_result.go
│   │   └── user.go
│   ├── interface.go              # 接口定义
│   └── resource/                 # 资源管理
│       └── resource.go
├── telegram/                     # Telegram Bot 处理
│   ├── client/                   # Telegram 客户端
│   ├── interfaces/               # Telegram 接口
│   ├── manager/                  # Telegram 管理器
│   └── telegram.go
├── utils/                        # 通用工具函数
│   ├── file/                     # 文件工具
│   ├── http/                     # HTTP 工具
│   ├── logger/                   # 日志工具
│   ├── map.go
│   └── utils.go
├── main.go                       # 主入口文件
├── go.mod                        # Go 模块定义
├── build.sh                      # 构建脚本
├── start.sh                      # 启动脚本
├── fix-mod.sh                    # 模块修复脚本
└── docs/                         # 文档目录
    ├── task.md
    └── web_search_guide.md
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

编辑 `config.yaml`：

```yaml
telegram:
  bot_token: "YOUR_TELEGRAM_BOT_TOKEN"  # 从 @BotFather 获取
  proxy: "http://127.0.0.1:7897"       # 代理地址（可选）
  prompt: "promet/telegram.md"          # 系统提示词文件

database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  username: "root"
  password: "your_password"
  database: "orange-agent"
  charset: "utf8mb4"
  parse_time: true
  loc: "Local"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: "300s"

logger:
  level: "debug"
  output: "both"                    # both/file/console
  file_path: "./log"
  file_name: "orange-agent.log"
  max_size: 10                      # MB
  max_age: 30                       # 天
  max_backups: 5
  compress: true
  show_caller: true
  module: "orange-agent"
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

- `/start` - 显示欢迎信息并启动服务
- `/help` - 显示帮助信息
- `/agents` - 显示所有可用的 AI 代理
- `/model` - 显示当前使用的模型

### 管理命令

- `/addAgent <agent_name> <base_url> <token>` - 添加新的 AI 代理

  - 示例：`/addAgent OpenAI https://api.openai.com/v1 sk-xxx`
- `/addModel <agent_id> <model_name>` - 为代理添加模型

  - 示例：`/addModel 1 gpt-4-turbo`
- `/switch <agent_id> <model_index>` - 切换到指定代理和模型

  - 示例：`/switch 1 2`

### 交互使用

直接发送消息即可与 AI 对话，系统会自动处理工具调用和上下文记忆。

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
- `randomReadFile` - 随机读取文件内容

### Git 工具

- `git_commit` - 提交代码
- `git_push` - 推送到远程仓库
- `git_diff` - 查看代码差异

### 数据库工具
- `database_query` - 执行数据库查询操作（支持SELECT语句）
- `database_execute` - 执行数据库写操作（支持INSERT、UPDATE、DELETE语句）

### 系统工具

- `curr_time` - 获取当前时间
- `build_tools` - 构建项目
- `project_reboot` - 重启项目
- `log_view` - 查看日志
- `env_manage` - 管理环境变量
- `test_run` - 运行测试
- `dependency_check` - 检查依赖
- `performance_monitor` - 性能监控
- `api_tester` - API 接口测试
- `skill` - 获取技能详细信息

### 🌐 联网搜索工具

- `web_search` - 联网搜索功能

**功能特性：**

- 🔍 **搜索引擎搜索**：支持 DuckDuckGo、Google、Bing 搜索
- 📄 **网页内容抓取**：获取网页正文内容
- 🚀 **无需 API Key**：DuckDuckGo 搜索无需配置 API 密钥
- 🌍 **多语言支持**：支持中英文搜索

**使用示例：**

```
# 搜索关键词
搜索 "Go 语言教程"

# 抓取网页内容
抓取 https://example.com 的内容

# 使用不同搜索引擎
用 Google 搜索 "人工智能最新进展"
```

**参数说明：**

- `query`: 搜索关键词或 URL
- `search_type`: 搜索类型
  - `search`: 搜索引擎搜索
  - `fetch`: 抓取网页内容
- `engine`: 搜索引擎（可选，默认 duckduckgo）
  - `duckduckgo`: DuckDuckGo（推荐，免费）
  - `google`: Google（需 API Key）
  - `bing`: Bing（需 API Key）
- `num_results`: 返回结果数量（默认 5 条，最多 10 条）

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

每个用户可以独立选择不同的 AI 模型，切换记录会保存在数据库中。

## 数据库表结构

项目使用以下核心表：

1. `users` - 用户信息
2. `agent_configs` - AI 代理配置
3. `memories` - 对话记忆
4. `agent_call_records` - AI 调用记录
5. `tasks` - 任务记录
6. `sub_tasks` - 子任务记录
7. `task_results` - 任务结果

## 开发指南

### 添加新工具

1. 在 `agent/tools/` 目录下创建新的工具包
2. 实现工具函数
3. 在对应工具的 `tools.go` 中注册工具
4. 重新编译运行

### 扩展 AI 代理

1. 在 `agent/` 中扩展代理类型
2. 实现相应的配置加载逻辑
3. 通过 `/addAgent` 命令添加新代理

### 任务处理流程

1. **任务分析** (`analyzer`) - 分析用户请求，确定任务类型
2. **任务编排** (`orchestrator`) - 规划任务执行步骤
3. **任务执行** (`executor`) - 执行具体操作
4. **任务总结** (`summarizer`) - 汇总执行结果

## 日志系统

日志系统支持多级别输出：

- `debug` - 调试信息
- `info` - 常规信息
- `warn` - 警告信息
- `error` - 错误信息

日志文件保存在 `./log/orange-agent.log`，支持轮转和压缩。

## 故障排除

### 常见问题

1. **无法连接到 Telegram**

   - 检查网络连接和代理设置
   - 确认 Bot Token 是否正确
2. **数据库连接失败**

   - 检查数据库配置
   - 确认 MySQL 服务正在运行
3. **AI 调用失败**

   - 检查代理配置和 API 密钥
   - 确认网络可以访问 AI 服务
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
5. 创建 Pull Request

## 相关技术

- [Go](https://golang.org/) - 编程语言 (v1.25.6)
- [LangChain Go](https://github.com/tmc/langchaingo) - AI 框架 (v0.1.14)
- [TeleBot](https://github.com/tucnak/telebot) - Telegram Bot 框架 (v3.3.8)
- [GORM](https://gorm.io/) - ORM 框架 (v1.31.1)
- [Viper](https://github.com/spf13/viper) - 配置管理 (v1.21.0)

## 更新日志

### v1.4.2 (2026-04-02)
- ✨ **新增技能管理工具**：支持查询技能详细信息，快速了解工具使用方法
- ⚡ **性能优化**：优化工具调用逻辑，提升响应速度
- 📚 **完善文档**：更新工具列表说明，补充技能系统功能介绍
- 🐛 **问题修复**：修复已知的小bug，提升系统稳定性

### v1.4.1 (2026-03-29)
- ✨ **新增数据库操作工具**：支持数据库查询和写入操作
- 🐛 **修复拼写错误**：修正 agent/client 目录下文件名错误
- 📝 **修正配置字段**：将配置中的 `promete` 字段修正为 `prompt`
- 📚 **完善文档**：补充数据库工具说明，更新项目结构描述

### v1.4.0 (2026-03-29)

- ✨ **新增 Telegram 消息通知功能**：支持发送系统通知和任务状态更新
- 🎯 **优化 Agent 单例模式**：改进 Agent 管理架构，提升性能和稳定性
- 📦 **重构项目结构**：采用更清晰的模块化架构
  - 新增 `agent/` 目录作为核心模块
  - 完善任务处理流程（analyzer → orchestrator → executor → summarizer）
  - 优化 `repository/` 层次结构，新增任务相关表
  - 改进 `telegram/` 模块，支持消息通知
- 📝 **更新文档**：修正 README 中的路径错误和结构描述
  - 修正 `lanchain/` → `langchain/`
  - 修正 `promete/telegram.text` → `promet/telegram.md`
  - 更新项目结构树，反映最新目录布局
- 🐛 **持续改进**：修复已知问题，提升稳定性

### v1.3.0 (2026-03-28)

- ✨ **重构项目结构**：采用更清晰的模块化架构
  - 新增 `common/` 目录存放通用工具
  - 优化 `repository/` 层次结构
  - 改进 `tools/` 工具分类
  - 完善 `langchain/` 目录结构
- 📝 **更新文档**：修正 README 中的路径错误
  - 修正 `lanchain/` → `langchain/`
  - 修正 `promete/telegram.text` → `promet/telegram.md`
- 🐛 **持续改进**：修复已知问题，提升稳定性

### v1.2.0

- ✨ 新增联网搜索功能
  - 支持 DuckDuckGo、Google、Bing 搜索引擎
  - 支持网页内容抓取
  - 无需 API Key 即可使用 DuckDuckGo 搜索
- 🐛 修复若干已知问题
- 📝 更新文档

### v1.1.0

- 🎉 初始版本发布
- 🤖 多 AI 代理支持
- 🛠️ 工具调用系统
- 💾 对话记忆功能

---

如有问题或建议，请通过 Issue 或 Pull Request 提交。