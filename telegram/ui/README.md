# Telegram UI 模块

## 概述

该模块为 Orange Agent 提供了基于 Telegram 的交互式用户界面，支持点击式菜单和按钮操作，极大地改善了用户体验。

## 功能特性

### 1. 🎯 核心功能

#### 点击式菜单
- **主菜单**：分类显示所有功能
- **子菜单**：按功能分类的详细操作
- **返回导航**：支持菜单层级返回

#### 按钮操作
- **直接命令执行**：点击按钮执行无参数命令
- **参数提示**：需要参数的命令会提示用户输入
- **状态保持**：记住用户当前状态和操作历史

#### 智能响应
- **自动菜单切换**：根据命令类型自动显示对应菜单
- **参数处理**：支持复杂参数格式
- **错误恢复**：用户输入错误时提供清晰指引

### 2. 🗺️ 菜单结构

#### 主菜单 (menu_main)
```
📁 文件   🔧 Git   🏗️ 项目
🗄️ 数据库 🤖 Agent ⚙️ 系统
🛠️ 工具   🤖 模型  ❓ 帮助
```

#### 文件菜单 (menu_file)
```
📋 列出文件  📄 读取文件
🔍 搜索文件  📝 写入文件
⬅️ 返回主菜单
```

#### Git菜单 (menu_git)
```
📊 Git状态  💾 提交更改
📤 推送代码  📥 差异对比
⬅️ 返回主菜单
```

#### 项目管理菜单 (menu_project)
```
🔨 构建项目  🧪 运行测试
🔄 重启项目  📦 检查依赖
📋 查看日志  📝 环境变量
⬅️ 返回主菜单
```

#### 数据库菜单 (menu_db)
```
🔍 查询数据  ✏️ 执行SQL
⬅️ 返回主菜单
```

#### Agent管理菜单 (menu_agent)
```
📋 列出Agent  🧪 测试Agent
➕ 添加Agent  ✖️ 删除Agent
🔄 更新Agent  ⬅️ 返回主菜单
```

#### 系统工具菜单 (menu_system)
```
📊 系统状态  🛠️ 工具列表
⏰ 当前时间  📈 性能监控
🌐 Web搜索  🔗 API测试
⬅️ 返回主菜单
```

#### 模型菜单 (menu_model)
```
📋 查看模型  🔄 切换模型
⬅️ 返回主菜单
```

#### 帮助菜单 (menu_help)
```
📖 命令帮助  📋 快速指南
🛠️ 使用示例  📝 使用技巧
⬅️ 返回主菜单
```

### 3. ⚙️ 技术架构

#### 核心组件
- **MenuManager**：管理菜单键盘和按钮布局
- **UIManager**：处理用户交互和状态管理
- **UserState**：维护用户会话状态

#### 状态管理
```go
type UserState struct {
    LastMenu     string                 // 最后显示的菜单
    LastCommand  string                 // 最后执行的命令
    LastMessage  string                 // 最后的消息
    LastResponse string                 // 最后的响应
    StateData    map[string]interface{} // 状态数据
    CreatedAt    time.Time              // 创建时间
    UpdatedAt    time.Time              // 更新时间
}
```

#### 工作流程
1. **用户点击按钮** → 触发回调处理
2. **检查按钮类型** → 菜单导航或命令执行
3. **处理参数输入** → 如果需要则提示用户
4. **执行命令** → 调用对应的命令处理器
5. **更新状态** → 记录用户操作历史
6. **返回结果** → 显示结果和适当菜单

### 4. 🚀 集成指南

#### 1. 初始化UI管理器
```go
// 在Telegram客户端中初始化
uiManager := ui.NewUIManager(cm, answer)
```

#### 2. 集成消息处理
```go
// 修改消息处理逻辑
c.bot.Handle(telebot.OnText, func(t telebot.Context) error {
    // 使用UI管理器处理消息
    result, menu, err := uiManager.HandleMessage(ctx, t, user, messageText)
    if err != nil {
        return t.Reply("❌ 处理消息时出错", telebot.ModeMarkdown)
    }
    
    // 发送带菜单的响应
    if menu != nil {
        return t.Reply(result, menu, telebot.ModeMarkdown)
    }
    return t.Reply(result, telebot.ModeMarkdown)
})

// 处理按钮回调
c.bot.Handle(telebot.OnCallback, func(t telebot.Context) error {
    result, menu, err := uiManager.HandleCallback(ctx, t, user, t.Data())
    if err != nil {
        return t.Respond(&telebot.CallbackResponse{
            Text: "❌ 处理回调时出错",
        })
    }
    
    // 发送响应
    if menu != nil {
        return t.Reply(result, menu, telebot.ModeMarkdown)
    }
    return t.Reply(result, telebot.ModeMarkdown)
})
```

#### 3. 添加启动菜单
```go
// 在/start命令中添加主菜单
c.bot.Handle("/start", func(t telebot.Context) error {
    welcomeMsg := "🤖 *欢迎使用 Orange Agent!*\n\n请选择要执行的操作："
    return t.Reply(welcomeMsg, uiManager.menuManager.GetMainMenu(), telebot.ModeMarkdown)
})
```

### 5. 📝 参数处理

#### 参数提示机制
- **需要参数的命令**：按钮显示为 `_prompt` 后缀
- **用户点击后**：显示参数输入提示
- **输入参数后**：自动执行完整命令

#### 支持的参数格式
```go
// 文件读取
/read main.go

// Git提交
/commit 修复了一个重要的bug

// 数据库查询
/db SELECT * FROM users WHERE status = 1

// Agent添加
/agentadd myagent|https://api.example.com|key123|openai|gpt-4
```

#### 参数验证
- **文件路径**：验证路径安全性
- **SQL语句**：检查基本语法
- **API密钥**：验证格式
- **模型名称**：检查是否可用

### 6. 🔧 扩展开发

#### 添加新菜单
1. 在 `menu_manager.go` 中添加新菜单方法
2. 在 `getMenuByData` 中添加映射
3. 在 `getMenuTitle` 中添加标题

#### 添加新命令按钮
1. 在对应菜单中添加按钮
2. 在 `GetCommandByData` 中添加处理逻辑
3. 在 `GetPromptMessage` 中添加参数提示

#### 自定义按钮样式
```go
// 使用表情符号和文字
telebot.Btn{Text: "📁 文件", Data: "menu_file"}

// 使用回调数据
telebot.Btn{Text: "执行", CallbackData: "cmd_execute"}

// 调整键盘布局
keyboard.Row(btn1, btn2, btn3)
keyboard.Row(btn4, btn5)
```

### 7. 🛡️ 安全考虑

#### 状态管理安全
- **会话超时**：自动清理过期会话
- **状态隔离**：不同用户状态完全隔离
- **数据清理**：定期清理状态数据

#### 输入验证
- **参数过滤**：防止注入攻击
- **路径限制**：限制文件访问范围
- **权限检查**：敏感操作需要确认

#### 错误处理
- **优雅降级**：UI错误时回退到文本模式
- **用户反馈**：清晰的错误提示
- **日志记录**：记录所有交互

### 8. 📊 性能优化

#### 内存管理
- **状态清理**：定期清理不活跃用户状态
- **缓存优化**：菜单键盘缓存
- **资源回收**：及时释放不再使用的资源

#### 响应时间
- **异步处理**：耗时操作异步执行
- **结果缓存**：常用结果缓存
- **预加载**：预加载常用菜单

### 9. 🧪 测试建议

#### 单元测试
```go
func TestMenuManager(t *testing.T) {
    mm := NewMenuManager(nil)
    
    // 测试菜单生成
    menu := mm.GetMainMenu()
    assert.NotNil(t, menu)
    
    // 测试命令映射
    cmd, args := mm.GetCommandByData("cmd_list")
    assert.Equal(t, "list", cmd)
    assert.Empty(t, args)
}
```

#### 集成测试
- 测试完整的用户交互流程
- 测试状态管理正确性
- 测试错误处理机制

#### 用户测试
- 邀请真实用户测试易用性
- 收集用户反馈改进UI
- A/B测试不同菜单布局

### 10. 📈 监控和日志

#### 监控指标
- **用户活跃度**：菜单使用频率
- **命令成功率**：命令执行成功率
- **响应时间**：UI响应延迟

#### 日志记录
- **用户操作**：记录所有按钮点击
- **状态变化**：记录状态变化历史
- **错误信息**：详细记录错误信息

## 总结

Telegram UI 模块通过点击式菜单和按钮操作，极大地提升了 Orange Agent 的用户体验。用户无需记忆复杂的命令，只需点击按钮即可执行常见操作，同时保留了命令输入功能以满足高级用户需求。该模块设计灵活、易于扩展，为未来的功能增强提供了良好的基础架构。