package logger

import (
	"fmt"
	"io"
	"orange-agent/config"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	levelStrings = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	levelColors  = []string{"\033[36m", "\033[32m", "\033[33m", "\033[31m", "\033[35m"}
	resetColor   = "\033[0m"
)

// Logger 日志记录器
type Logger struct {
	level      LogLevel
	output     io.Writer
	file       *os.File
	mu         sync.Mutex
	module     string
	showCaller bool
}

// Config 日志配置

var (
	defaultLogger *Logger
	once          sync.Once
)

// NewLogger 创建新的日志记录器
func NewLogger(config config.Logger) (*Logger, error) {
	logger := &Logger{
		module:     config.Module,
		showCaller: config.ShowCaller,
	}
	// 设置日志级别
	switch strings.ToLower(config.Level) {
	case "debug":
		logger.level = DEBUG
	case "info":
		logger.level = INFO
	case "warn":
		logger.level = WARN
	case "error":
		logger.level = ERROR
	case "fatal":
		logger.level = FATAL
	default:
		logger.level = INFO
	}

	// 设置输出目标
	var writers []io.Writer

	if config.Output == "console" || config.Output == "both" {
		writers = append(writers, os.Stdout)
	}

	if config.Output == "file" || config.Output == "both" {
		if config.FilePath == "" {
			config.FilePath = "./logs"
		}
		if config.FileName == "" {
			config.FileName = "app.log"
		}

		// 创建日志目录
		if err := os.MkdirAll(config.FilePath, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %v", err)
		}

		filePath := filepath.Join(config.FilePath, config.FileName)
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件失败: %v", err)
		}

		logger.file = file
		writers = append(writers, file)
	}

	if len(writers) == 0 {
		logger.output = os.Stdout
	} else if len(writers) == 1 {
		logger.output = writers[0]
	} else {
		logger.output = io.MultiWriter(writers...)
	}
	return logger, nil
}

// InitDefaultLogger 初始化默认日志记录器
func InitDefaultLogger(config config.Logger) error {
	var err error
	once.Do(func() {
		defaultLogger, err = NewLogger(config)
	})
	return err
}

// GetLogger 获取默认日志记录器
func GetLogger() *Logger {
	if defaultLogger == nil {
		// 使用默认配置
		config := config.Logger{
			Level:  "info",
			Output: "console",
		}
		defaultLogger, _ = NewLogger(config)
	}
	return defaultLogger
}

// WithModule 创建带模块名的日志记录器
func (l *Logger) WithModule(module string) *Logger {
	return &Logger{
		level:      l.level,
		output:     l.output,
		file:       l.file,
		module:     module,
		showCaller: l.showCaller,
	}
}

// log 内部日志方法
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	//判断日志是否需要轮转
	if time.Now().Format("2006-01-02") != l.GetLastModifiedTime() {
		if err := l.Rotate(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
		}
	}

	now := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := levelStrings[level]
	levelColor := levelColors[level]

	// 构建日志前缀
	var prefix string
	if l.module != "" {
		prefix = fmt.Sprintf("[%s] [%-5s] [%s]", now, levelStr, l.module)
	} else {
		prefix = fmt.Sprintf("[%s] [%-5s]", now, levelStr)
	}

	// 添加调用者信息
	if l.showCaller {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			// 只显示文件名，不显示完整路径
			shortFile := filepath.Base(file)
			prefix += fmt.Sprintf(" [%10s:%-3d]", shortFile, line)
		}
	}

	// 构建完整日志消息
	message := fmt.Sprintf(format, args...)

	logLine := fmt.Sprintf("%s %s\n", prefix, message)

	// 添加颜色（仅控制台输出）
	if l.output == os.Stdout || l.output == os.Stderr {
		logLine = fmt.Sprintf("%s%s%s %s\n", levelColor, prefix, resetColor, message)
	}

	// 写入日志
	fmt.Fprint(l.output, logLine)

	// 如果是FATAL级别，退出程序
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug 调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn 警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error 错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal 致命错误日志
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// 全局日志函数
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GetLogger().Fatal(format, args...)
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Rotate 轮转日志文件
func (l *Logger) Rotate() error {
	if l.file == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 关闭当前文件
	if err := l.file.Close(); err != nil {
		return err
	}

	// 重命名旧文件 app-2024-06-01.log
	oldPath := l.file.Name()
	newPath := fmt.Sprintf("%s-%s.log", strings.TrimSuffix(oldPath, ".log"), l.GetLastModifiedTime())
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	// 创建新文件
	if err := os.MkdirAll(filepath.Dir(oldPath), 0755); err != nil {
		return err
	}

	// 打开新文件
	file, err := os.OpenFile(oldPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.file = file

	return nil
}

// 获取当前日志文件的最后修改时间
func (l *Logger) GetLastModifiedTime() string {
	if l.file == nil {
		return ""
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	fileInfo, err := l.file.Stat()
	if err != nil {
		return ""
	}

	return fileInfo.ModTime().Format("2006-01-02")
}
