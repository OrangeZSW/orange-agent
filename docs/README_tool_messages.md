# 工具调用消息美化方案

## 概述
本方案重构了工具调用时发送到Telegram的消息格式，提供了更加美观、可读的消息展示。

## 主要改进

### 1. 消息格式美化
使用统一的格式化器 (`ToolMessageFormatter`) 来美化所有工具调用相关消息：

#### 工具调用消息 (FormatToolCallMessage)
```
🛠️ *工具调用*

📋 *工具名称*: File Read
⚙️ *参数*:
```json
{
  "file_path": "agent/task/orchestrator.go"
}
```
```

#### 工具调用成功消息 (FormatToolSuccessMessage)
```
✅ *工具调用成功*

📋 *工具名称*: Build Tools
⚙️ *参数*:
```json
{}
```
📊 *输出*:
```
构建成功
```
```

#### 工具调用失败消息 (FormatToolErrorMessage)
```
❌ *工具调用失败*

📋 *工具名称*: File Read
⚙️ *参数*:
```json
{
  "file_path": "nonexistent.txt"
}
```
💥 *错误*:
```
文件不存在
```
```

### 2. 功能特性

#### 工具名称美化
- 将下划线分隔的工具名转换为首字母大写的空格分隔形式
- 例如：`file_read` → `File Read`
- 例如：`database_query` → `Database Query`

#### 参数格式化
- JSON 参数自动格式化缩进
- 空参数显示为 `无参数`
- 使用代码块展示参数内容

#### 结果处理
- 自动识别并格式化 JSON 结果
- 长结果自动截断并添加省略号
- 使用代码块展示结果内容

#### 图标使用
- 🛠️ - 工具调用开始
- ✅ - 工具调用成功
- ❌ - 工具调用失败
- 📋 - 工具名称
- ⚙️ - 参数配置
- 📊 - 输出结果
- 💥 - 错误信息

## 实现细节

### 1. 核心类
```go
// ToolMessageFormatter - 工具消息格式化器
type ToolMessageFormatter struct{}

// 主要方法：
- FormatToolCallMessage()     // 格式化工具调用消息
- FormatToolSuccessMessage()  // 格式化成功消息
- FormatToolErrorMessage()    // 格式化错误消息
```

### 2. 集成位置
修改了 `agent/client/clietn.go` 文件：
```go
// 在 HandleToolCalls 方法中替换了原来的消息发送逻辑
callMessage := c.messageFormatter.FormatToolCallMessage(
    toolcall.FunctionCall.Name, 
    toolcall.FunctionCall.Arguments,
)
c.manager.TeleGramSendMessage(callMessage)
```

### 3. 测试覆盖
创建了完整的单元测试：
- 工具名称美化测试
- 参数格式化测试
- 结果处理测试
- 各种边界情况测试

## 使用示例

### 原始消息
```
调用工具:file_read,参数:{"file_path": "agent/task/orchestrator.go"}
调用工具:build_tools,参数:{}
```

### 美化后消息
```
🛠️ *工具调用*

📋 *工具名称*: File Read
⚙️ *参数*:
```json
{
  "file_path": "agent/task/orchestrator.go"
}
```

✅ *工具调用成功*

📋 *工具名称*: Build Tools
⚙️ *参数*:
```json
{}
```
📊 *输出*:
```
构建工具执行完成
```
```

## 优势

1. **可读性提升** - 结构化展示，一目了然
2. **美观性** - 使用图标和格式美化
3. **可维护性** - 集中化的格式化逻辑
4. **可扩展性** - 易于添加新的格式化类型
5. **一致性** - 统一的消息风格

## 配置选项

消息格式化器支持以下配置（可通过扩展实现）：
- 最大结果长度（默认500字符）
- 是否显示完整结果
- 自定义图标
- 消息模板定制

## 未来改进

1. **支持更多消息类型** - 如工具链调用、批量操作等
2. **国际化支持** - 多语言消息格式
3. **主题定制** - 支持不同的显示风格
4. **交互式消息** - 支持按钮和交互操作

## 迁移指南

从旧的消息格式迁移到新格式：
1. 引入 `agent/utils/tool_message_formatter.go`
2. 更新 `agent/client/clietn.go` 中的消息发送逻辑
3. 运行测试验证功能正常

## 兼容性
- 完全向后兼容，不影响现有功能
- 只改变消息展示方式，不改变业务逻辑
- 支持所有现有工具调用场景