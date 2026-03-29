# Langchain 重构说明

## 重构目标

将原来耦合严重的 langchain 模块进行解耦，遵循单一职责原则和依赖倒置原则。

## 架构设计

```
langchain/
├── interfaces/          # 核心接口定义
│   └── interfaces.go    # 定义所有核心接口
├── llm/                # LLM 提供者层
│   ├── provider.go     # LLM 提供者接口
│   └── openai.go       # OpenAI 实现
├── memory/             # 记忆管理层
│   ├── manager.go      # 记忆管理接口
│   └── db_memory.go    # 数据库记忆实现
├── message/            # 消息处理层
│   ├── builder.go      # 消息构建器
│   ├── cleaner.go      # 消息清理器
│   └── token_counter.go # Token 计数器
├── tool/               # 工具管理层
│   ├── executor.go     # 工具执行器
│   └── manager.go      # 工具管理器
├── chain/              # 链式调用层
│   └── chain.go        # 主链实现
└── handler/            # 处理器层
    └── answer_handler.go # 答案处理器
```

## 各层职责

### 1. Interfaces 层
- 定义核心接口，确保各层解耦
- 提供统一的抽象层

### 2. LLM 层
- **provider.go**: 定义 LLM 提供者接口
- **openai.go**: OpenAI API 的具体实现
- 职责：管理 LLM 连接、配置和调用

### 3. Memory 层
- **manager.go**: 记忆管理接口
- **db_memory.go**: 基于数据库的记忆存储
- 职责：管理对话历史和用户记忆

### 4. Message 层
- **builder.go**: 构建对话消息
- **cleaner.go**: 清理和优化消息（基于 token 或数量）
- **token_counter.go**: 计算 token 数量
- 职责：处理消息格式和优化

### 5. Tool 层
- **executor.go**: 执行具体工具
- **manager.go**: 管理工具调用流程
- 职责：工具调用和结果处理

### 6. Chain 层
- **chain.go**: 协调各层完成完整的对话流程
- 职责：编排整个处理流程

### 7. Handler 层
- **answer_handler.go**: 处理用户问题并返回答案
- 职责：对外提供统一的调用接口

## 依赖关系

```
Handler → Chain → [LLM, Memory, Message, Tool]
                      ↓
                 Interfaces
```

## 使用示例

```go
// 创建处理器
answerHandler := handler.NewAnswerHandler()

// 处理用户问题
answer := answerHandler.AnswerQuestion(user, memory, prompt)
```

## 优势

1. **单一职责**: 每个模块只负责一个明确的功能
2. **易于测试**: 各层可以独立进行单元测试
3. **易于扩展**: 可以轻松添加新的 LLM 提供者或记忆存储方式
4. **低耦合**: 通过接口解耦，各层可以独立变化
5. **高内聚**: 相关功能聚合在同一模块中

## 迁移说明

原有的 `langchain` 包中的文件已重构到新的目录结构中：
- `lanchain.go` → `llm/openai.go` + `chain/chain.go`
- `handeler.go` → `handler/answer_handler.go`
- `message.go` → `message/` 目录下的多个文件
- `executeTool.go` → `tool/` 目录下的文件

外部调用只需更新导入路径：
```go
// 旧代码
import "orange-agent/langchain"

// 新代码
import "orange-agent/langchain/handler"
```
