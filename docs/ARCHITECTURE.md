# Orange Agent 系统架构文档

## 📋 系统概述

Orange Agent 是一个基于 Go 语言开发的智能代理系统，集成了 LangChain 框架，支持多 AI 模型、工具调用和 Telegram 消息通知功能。系统采用模块化设计，具有高度可扩展性。

### 主要模块

```
Orange Agent
├── 应用层 (main.go)
├── 代理层 (agent/)
│   ├── 客户端管理 (client/)
│   ├── 代理管理 (manager/)
│   ├── 任务编排 (task/)
│   └── 工具系统 (tools/)
├── 配置层 (config/)
├── 领域层 (domain/)
├── 数据访问层 (repository/)
├── Telegram 层 (telegram/)
└── 工具层 (utils/)
```

### 1. Agent 模块 (agent/)

#### 1.1 Agent 主逻辑 (agent.go)

- 实现单例模式，确保全局唯一 Agent 实例
- 支持两种聊天模式：
  - **普通模式** (NORMAL)：直接调用 AI 进行对话
  - **任务模式** (TASK)：使用任务编排器处理复杂任务

#### 1.2 客户端管理 (client/)

- 负责与 AI 服务端的通信
- 支持工具调用和函数调用
- 处理 AI 响应解析

#### 1.3 代理管理 (manager/)

- 管理多个 AI 代理实例
- 处理模型切换和配置加载
- 提供代理状态监控

#### 1.4 任务编排系统 (task/)

```
task/
├── context_manager.go     # 任务上下文管理
├── dag_engine.go         # DAG 执行引擎
├── errors.go             # 错误定义
├── interfaces.go         # 接口定义
├── orchestrator.go       # 任务编排器
├── result_aggregator.go  # 结果聚合器
├── task_analyzer.go      # 任务分析器
├── task_queue.go         # 任务队列
├── task_splitter.go      # 任务分割器
├── task_summarizer.go    # 任务总结器
└── worker_pool.go        # 工作池管理
```

**任务处理流程**：

1. **任务分析**：分析用户请求，确定任务类型
2. **任务分割**：将复杂任务拆分为子任务
3. **依赖分析**：构建子任务依赖关系图 (DAG)
4. **任务调度**：按照依赖关系调度执行
5. **结果聚合**：汇总所有子任务结果
6. **任务总结**：生成最终输出

#### 1.5 工具系统 (tools/)

```
tools/
├── agent/        # Agent 管理工具
├── database/     # 数据库操作工具
├── file/         # 文件操作工具
├── git/          # Git 操作工具
├── system/       # 系统工具
├── time.go       # 时间工具
└── tools.go      # 工具注册和初始化
```

**工具分类**：

1. **文件工具**：文件读写、搜索、列表、删除等
2. **Git 工具**：提交、推送、差异查看
3. **数据库工具**：查询、执行 SQL 语句
4. **系统工具**：构建、重启、测试、性能监控
5. **Agent 工具**：Agent 配置管理
6. **联网工具**：网络搜索和内容抓取

### 2. 配置系统 (config/)

- 基于 Viper 的配置管理
- 支持 YAML 配置文件
- 支持环境变量覆盖
- 配置热加载支持

### 3. 领域模型 (domain/)

```
domain/
├── agentConfig.go   # Agent 配置模型
├── config.go        # 应用配置模型
├── memory.go        # 对话记忆模型
├── task.go          # 任务相关模型
└── user.go          # 用户模型
```

**核心模型**：

- **Task**：主任务模型，包含子任务和状态
- **SubTask**：子任务模型，支持依赖关系
- **DependencyGraph**：依赖图结构
- **TaskResult**：任务执行结果

### 4. 数据访问层 (repository/)

```
repository/
├── db/              # 数据库连接
├── gorm/           # GORM 模型定义
├── interface.go    # 仓库接口
└── resource/       # 资源管理
```

**数据库表**：

1. `users` - 用户信息
2. `agent_configs` - AI 代理配置
3. `memories` - 对话记忆
4. `agent_call_records` - AI 调用记录
5. `tasks` - 任务记录
6. `sub_tasks` - 子任务记录
7. `task_results` - 任务结果

### 5. Telegram 集成 (telegram/)

```
telegram/
├── client/          # Telegram 客户端
├── command/         # 命令处理
├── interfaces/      # 接口定义
├── manager/         # 管理器
├── telegram.go      # 主逻辑
└── ui/              # 用户界面组件
```

**Orange Agent 系统文档** - 版本 1.4.1
最后更新：2026-03-29
文档状态：完整
