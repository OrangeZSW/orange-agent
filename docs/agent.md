# Orange Agent 系统架构文档

## 概述

Orange Agent 是一个基于 Go 语言开发的 Telegram 智能代理机器人，集成了 LangChain 框架，支持多 AI 模型切换、工具调用和任务编排功能。系统采用模块化设计，具备强大的任务处理能力和灵活的扩展性。

## 核心架构

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    Telegram 用户界面层                        │
├─────────────────────────────────────────────────────────────┤
│                    Agent 核心层                              │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Agent     │  │   Manager    │  │     Client       │  │
│  │  (单例模式)  │  │ (用户管理)   │  │ (AI 模型调用)    │  │
│  └─────────────┘  └──────────────┘  └──────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    任务处理层                                │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  分析器     │  │  拆分器      │  │  编排器          │  │
│  │ TaskAnalyzer│  │ TaskSplitter │  │ TaskOrchestrator│  │
│  └─────────────┘  └──────────────┘  └──────────────────┘  │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  工作池     │  │  DAG引擎     │  │  总结器          │  │
│  │ WorkerPool  │  │ DAGEngine    │  │ TaskSummarizer   │  │
│  └─────────────┘  └──────────────┘  └──────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    工具系统层                                │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  文件工具   │  │  Git工具     │  │  系统工具        │  │
│  └─────────────┘  └──────────────┘  └──────────────────┘  │
│  ┌─────────────┐  ┌──────────────┐                        │
│  │  Agent工具  │  │  数据库工具  │                        │
│  └─────────────┘  └──────────────┘                        │
├─────────────────────────────────────────────────────────────┤
│                    数据访问层                                │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  用户表     │  │  Agent配置   │  │  任务记录        │  │
│  └─────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 核心模块详解

### 1. Agent 模块 (`agent/agent.go`)

Agent 是系统的核心，采用单例模式设计，负责协调整个系统的运行。

**主要功能：**
- 提供统一的聊天接口 `Chat()` 和 `TeleGramChat()`
- 支持多种聊天模式：NORMAL（普通对话）和 TASK（任务模式）
- 任务模式下的智能任务处理

**代码示例：**
```go
// Agent 单例模式
var (
    Agent interfaces.Agent
    once  sync.Once
)

type agent struct {
    repo     *repository.Repositories
    Telegram interfaces.Telegram
    log      *logger.Logger
}

func NewAgent() interfaces.Agent {
    once.Do(func() {
        Agent = &agent{
            repo: resource.GetRepositories(),
            log:  logger.GetLogger(),
        }
    })
    return Agent
}
```

### 2. Manager 模块 (`agent/manager/manager.go`)

Manager 负责用户管理和技能管理。

**主要功能：**
- 用户信息管理
- AI 调用记录保存
- Telegram 消息发送
- 技能加载和管理

**技能系统：**
系统支持从 `SKILL.md` 文件加载技能，技能格式：
```yaml
---
name: 技能名称
description: 技能描述
---
技能详细内容
```

### 3. 任务处理系统 (`agent/task/`)

任务处理系统是 Orange Agent 的核心功能，支持复杂的任务分解和执行。

#### 3.1 任务分析器 (`task_analyzer.go`)

**功能：**
- 分析用户输入的任务复杂度
- 评估子任务数量
- 推荐执行引擎（顺序/DAG/并行）

**分析结果结构：**
```go
type AnalysisResult struct {
    TaskType          string   `json:"task_type"`
    Complexity        string   `json:"complexity"` // low, medium, high
    EstimatedSubtasks int      `json:"estimated_subtasks"`
    KeyObjectives     []string `json:"key_objectives"`
    Constraints       []string `json:"constraints"`
    RecommendEngine   string   `json:"recommend_engine"` // sequential, dag, parallel
    EstimatedTime     int      `json:"estimated_time"`   // 预估执行时间（分钟）
}
```

#### 3.2 任务拆分器 (`task_splitter.go`)

**功能：**
- 将总任务拆分为多个子任务
- 建立任务依赖关系
- 设置执行顺序和并行标志

**子任务结构：**
```go
type SubTask struct {
    Description    string                 `json:"description"`
    Input          map[string]interface{} `json:"input"`
    Dependencies   []string               `json:"dependencies"`
    ExecutionOrder int                    `json:"execution_order"`
    CanParallel    bool                   `json:"can_parallel"`
    IsDAGNode      bool                   `json:"is_dag_node"`
}
```

#### 3.3 任务编排器 (`orchestrator.go`)

**功能：**
- 协调整个任务执行流程
- 根据任务复杂度选择执行引擎
- 管理任务上下文和状态

**编排器配置：**
```go
type OrchestratorConfig struct {
    WorkerCount     int
    QueueBufferSize int
    UseDAGEngine    bool // 是否使用DAG引擎
}
```

#### 3.4 DAG 引擎 (`dag_engine.go`)

**功能：**
- 构建和管理有向无环图
- 执行拓扑排序
- 处理复杂的任务依赖关系

**DAG 结构：**
```go
type DependencyGraph struct {
    Nodes    []*DAGNode
    Edges    []*DAGEdge
    Topology []string
    Metadata map[string]any
}
```

#### 3.5 工作池 (`worker_pool.go`)

**功能：**
- 管理并发执行的工作线程
- 负责任务队列管理
- 处理任务执行结果

**工作池配置：**
```go
type WorkerPool struct {
    workerCount    int
    taskQueue      *TaskQueue
    contextManager *ContextManager
    resultChan     chan *domain.SubTask
    taskChat       TaskChat
}
```

#### 3.6 上下文管理器 (`context_manager.go`)

**功能：**
- 管理每个任务的对话上下文
- 控制 Token 使用量
- 支持上下文压缩和摘要

**上下文结构：**
```go
type TaskContext struct {
    SystemPrompt string
    Messages     []domain.Message
    TokenCount   int
    Metadata     map[string]interface{}
}
```

### 4. 工具系统 (`agent/tools/`)

系统提供了丰富的工具集，支持各种操作。

#### 4.1 工具分类

**文件工具 (`tools/file/`):**
- `file_read` - 读取文件内容
- `file_write` - 写入文件内容
- `file_list` - 列出目录文件
- `file_delete` - 删除文件
- `file_rename` - 重命名文件
- `file_search` - 搜索文件内容
- `randomReadFile` - 随机读取文件内容

**Git 工具 (`tools/git/`):**
- `git_commit` - 提交代码
- `git_push` - 推送到远程仓库
- `git_diff` - 查看代码差异

**数据库工具 (`tools/database/`):**
- `database_query` - 执行数据库查询操作
- `database_execute` - 执行数据库写操作

**系统工具 (`tools/system/`):**
- `curr_time` - 获取当前时间
- `build_tools` - 构建项目
- `project_reboot` - 重启项目
- `log_view` - 查看日志
- `env_manage` - 管理环境变量
- `test_run` - 运行测试
- `dependency_check` - 检查依赖
- `performance_monitor` - 性能监控
- `api_tester` - API接口测试
- `web_search` - 联网搜索功能

**Agent 管理工具 (`tools/agent/`):**
- `agent_add` - 添加 Agent 配置
- `agent_remove` - 删除 Agent
- `agent_list` - 列出所有 Agent
- `agent_update` - 更新 Agent 配置
- `agent_test` - 测试 Agent 连接

#### 4.2 工具注册机制

所有工具通过统一的接口注册：
```go
func InitTools() {
    Once.Do(func() {
        Tools = append(Tools, file.FileTools...)
        Tools = append(Tools, CurrTimeTool)
        Tools = append(Tools, git.GitTools...)
        Tools = append(Tools, system.SystemTools...)
        Tools = append(Tools, agent.AgentTools...)
        Tools = append(Tools, database.DatabaseTools...)
    })
}
```

### 5. 任务处理流程

#### 5.1 普通对话流程

```
用户输入 → Agent.Chat() → Client → AI模型 → 返回结果
```

#### 5.2 任务模式流程

```
用户输入 → Agent.TaskChat()
    ↓
TaskOrchestrator.Execute()
    ↓
1. TaskAnalyzer.Analyze()    # 任务分析
    ↓
2. TaskSplitter.Split()      # 任务拆分
    ↓
3. 选择执行引擎:
   - 顺序引擎: executeSequential()
   - DAG引擎: dagEngine.ExecuteDAG()
    ↓
4. TaskSummarizer.Summarize() # 结果总结
    ↓
返回最终结果
```

#### 5.3 DAG 执行流程

```
1. 构建依赖图 (buildDAG)
2. 拓扑排序 (topologicalSort)
3. 按拓扑顺序执行 (executeTopology)
4. 收集依赖结果
5. 聚合最终结果
```

### 6. 数据库设计

#### 6.1 核心表结构

**用户表 (users):**
- ID, TelegramID, Username, ModelName, ChainMode
- CreatedAt, UpdatedAt

**Agent 配置表 (agent_configs):**
- ID, Name, BaseURL, Token, Models
- CreatedAt, UpdatedAt

**任务表 (tasks):**
- ID, SessionID, Description, Status, Result
- Subtasks (关联的子任务)

**子任务表 (sub_tasks):**
- ID, TaskID, Description, Status, Input, Output
- Dependencies, ExecutionOrder, CanParallel

#### 6.2 数据关系

```
用户 → 多个Agent配置
用户 → 多个任务记录
任务 → 多个子任务
子任务 → 多个依赖关系
```

### 7. 错误处理机制

系统采用多级错误处理机制：

#### 7.1 任务级别错误
- 子任务失败不影响其他任务执行
- 提供详细的错误信息和堆栈跟踪

#### 7.2 系统级别错误
- 数据库连接失败
- AI 模型调用失败
- 工具执行异常

#### 7.3 错误恢复策略
- 重试机制
- 降级处理
- 优雅降级

### 8. 性能优化

#### 8.1 并发处理
- WorkerPool 支持并发执行
- 任务队列缓冲机制
- 上下文独立管理

#### 8.2 内存管理
- 上下文压缩策略
- 结果缓存机制
- 资源池化

#### 8.3 数据库优化
- 连接池管理
- 批量操作
- 索引优化

### 9. 扩展性设计

#### 9.1 插件化架构
- 工具系统支持热插拔
- 模型系统支持动态配置
- 任务处理器可扩展

#### 9.2 接口设计
```go
type Agent interface {
    Chat(ctx context.Context, messages []domain.Message) string
    TeleGramChat(ctx context.Context, modelName string, message []llms.MessageContent) string
    TaskChat(ctx context.Context, question string) string
}

type TaskChat interface {
    Chat(ctx context.Context, messages []domain.Message) string
}

type Manager interface {
    SaveCallRecord(message []llms.MessageContent, resp *llms.ContentResponse, agentConfig *domain.AgentConfig) error
    TeleGramSendMessage(text string)
}
```

### 10. 安全性考虑

#### 10.1 输入验证
- 用户输入验证
- 文件路径安全检查
- SQL 注入防护

#### 10.2 权限控制
- 用户身份验证
- 操作权限检查
- 资源访问控制

#### 10.3 数据安全
- 敏感信息加密
- 日志脱敏
- 安全审计

### 11. 部署和运维

#### 11.1 环境要求
- Go 1.25.6+
- MySQL 5.7+
- Telegram Bot Token

#### 11.2 配置文件
```yaml
telegram:
  bot_token: "YOUR_BOT_TOKEN"
  proxy: "http://127.0.0.1:7897"
  prompt: "promet/telegram.md"

database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  database: "orange-agent"
  
logger:
  level: "debug"
  output: "both"
```

#### 11.3 监控和日志
- 多级别日志系统
- 性能监控
- 健康检查

### 12. 使用示例

#### 12.1 基本使用
```go
// 创建Agent实例
agent := agent.NewAgent()

// 普通对话
result := agent.Chat(ctx, messages)

// 任务模式
result := agent.TaskChat(ctx, "分析这个项目并生成报告")
```

#### 12.2 工具调用示例
```json
{
  "name": "file_read",
  "parameters": {
    "file_path": "./README.md"
  }
}
```

#### 12.3 任务处理示例
```
用户输入: "帮我分析项目结构，然后生成文档"

处理流程:
1. 分析任务 → 分析结果为"文档生成"，复杂度中等
2. 拆分任务 → 拆分为: ①分析结构 ②生成文档
3. 执行任务 → 顺序执行两个子任务
4. 总结结果 → 生成最终文档
```

### 13. 最佳实践

#### 13.1 开发规范
- 遵循 Go 语言最佳实践
- 统一的错误处理模式
- 完善的测试覆盖

#### 13.2 性能优化
- 合理设置 Worker 数量
- 控制上下文大小
- 优化数据库查询

#### 13.3 安全实践
- 定期更新依赖
- 安全配置检查
- 访问日志监控

### 14. 故障排除

#### 14.1 常见问题
1. **数据库连接失败**：检查配置和网络
2. **AI 模型调用失败**：检查 API 密钥和网络
3. **工具执行异常**：检查权限和路径
4. **任务执行超时**：调整超时设置或减少任务复杂度

#### 14.2 调试方法
```bash
# 查看日志
tail -f log/orange-agent.log

# 调试模式运行
LOG_LEVEL=debug ./orange-agent

# 性能监控
./orange-agent --monitor
```

### 15. 未来规划

#### 15.1 功能扩展
- 支持更多 AI 模型
- 增加更多工具类型
- 优化任务编排算法

#### 15.2 性能优化
- 分布式任务处理
- 结果缓存机制
- 异步处理优化

#### 15.3 用户体验
- 图形化界面
- 实时进度显示
- 结果可视化

---

## 总结

Orange Agent 是一个功能强大、架构灵活的智能代理系统，具有以下特点：

1. **模块化设计**：清晰的层次结构和模块划分
2. **强大的任务处理能力**：支持复杂的任务分解和执行
3. **丰富的工具集**：覆盖文件、Git、数据库、系统等操作
4. **灵活的扩展性**：支持插件化扩展和自定义工具
5. **良好的性能**：并发处理、内存优化、数据库优化
6. **完善的安全性**：输入验证、权限控制、数据安全

系统适用于各种自动化任务处理场景，特别是需要复杂任务分解和AI辅助的场景。