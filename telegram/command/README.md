# Telegram快捷命令模块

## 概述

该模块为Telegram机器人添加强大的快捷命令功能，允许用户通过简单的命令快速执行常见操作，无需通过AI助手进行复杂的对话。

## 功能特性

### 1. 命令分类

#### 📁 文件操作
- `/list` - 列出当前目录文件
- `/read <文件路径>` - 读取文件内容
- `/search <内容>` - 搜索文件内容

#### 🔧 Git操作
- `/git` - 查看Git状态和更改
- `/commit <消息>` - 提交更改
- `/push [分支]` - 推送到远程

#### 🏗️ 项目操作
- `/build` - 构建项目
- `/test [包路径]` - 运行测试
- `/reboot` - 重启项目
- `/deps` - 检查项目依赖
- `/logs` - 查看应用日志

#### 🗄️ 数据库操作
- `/db <SQL查询>` - 执行SELECT查询
- `/dbe <SQL语句>` - 执行写操作（需确认）

#### 🤖 Agent管理
- `/agents` - 列出所有Agent
- `/agenttest <名称>` - 测试Agent连接
- `/agentadd` - 添加新Agent
- `/agentremove <名称>` - 删除Agent
- `/agentupdate` - 更新Agent配置

#### 📋 系统命令
- `/help` - 显示所有可用命令
- `/status` - 查看系统状态
- `/tools` - 列出所有可用工具

### 2. 架构设计

```
telegram/command/
├── command.go          # 命令管理器核心
├── file_commands.go    # 文件操作命令
├── git_commands.go     # Git操作命令
├── project_commands.go # 项目操作命令
├── db_commands.go      # 数据库命令
├── agent_commands.go   # Agent管理命令
└── command_test.go     # 单元测试
```

### 3. 核心组件

#### CommandHandler 接口
```go
type CommandHandler interface {
    Command() string      // 命令名称
    Description() string  // 命令描述
    Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string
}
```

#### CommandManager 管理器
- 统一管理所有命令处理器
- 提供命令注册、查找和执行功能
- 支持命令自动发现和扩展

### 4. 使用示例

#### 基本使用
```
/help                    # 查看所有命令
/status                  # 查看系统状态
/list                    # 列出文件
```

#### 文件操作
```
/read main.go            # 读取main.go文件
/search function         # 搜索包含"function"的文件
```

#### Git操作
```
/git                     # 查看Git状态
/commit 修复bug          # 提交更改
/push main              # 推送到main分支
```

#### 项目操作
```
/build                   # 构建项目
/test                    # 运行测试
/logs                    # 查看日志
```

#### 数据库操作
```
/db SELECT * FROM users  # 查询用户表
/dbe UPDATE users SET status=1 WHERE id=1  # 更新用户状态
```

### 5. 集成方式

#### 集成到Telegram客户端
```go
// 在listenMessage中添加命令处理逻辑
func (c *client) listenMessage() {
    // 创建命令管理器
    cmdManager := command.NewCommandManager(c.repo)
    
    c.bot.Handle(telebot.OnText, func(t telebot.Context) error {
        text := t.Text()
        
        // 检查是否为命令
        if strings.HasPrefix(text, "/") {
            // 执行命令
            result := cmdManager.Execute(ctx, t, user, text)
            return t.Reply(result, telebot.ModeMarkdown)
        }
        
        // 原有消息处理逻辑
        // ...
    })
}
```

### 6. 扩展命令

#### 添加新命令步骤
1. 实现`CommandHandler`接口
2. 在`command.go`的`registerHandlers()`中注册
3. 在`command_test.go`中添加测试

#### 示例：添加天气命令
```go
// weather_command.go
type WeatherCommand struct{}

func (w *WeatherCommand) Command() string {
    return "weather"
}

func (w *WeatherCommand) Description() string {
    return "获取天气信息"
}

func (w *WeatherCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
    // 实现天气查询逻辑
    return "🌤️ 今日天气: 晴天，25°C"
}
```

### 7. 安全考虑

#### 权限控制
- 数据库写操作需要确认
- 敏感操作有明确警告
- 命令参数验证

#### 输入验证
- 检查SQL注入风险
- 验证文件路径安全性
- 限制命令执行权限

### 8. 未来扩展

#### 计划功能
- [ ] 命令别名系统
- [ ] 命令权限分级
- [ ] 命令历史记录
- [ ] 命令自动补全
- [ ] 命令组合执行

#### 性能优化
- 命令缓存机制
- 异步命令执行
- 结果分页显示

### 9. 注意事项

1. **命令冲突**：确保命令名称唯一
2. **权限管理**：敏感命令需要权限验证
3. **错误处理**：友好的错误提示
4. **日志记录**：记录所有命令执行

### 10. 测试

运行测试：
```bash
cd telegram/command
go test -v
```

## 总结

Telegram快捷命令模块极大地提升了用户体验，通过简单的命令即可完成复杂操作，同时保持了系统的安全性和可扩展性。