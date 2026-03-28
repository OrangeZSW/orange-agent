package file

import (
	"fmt"
	"os"
	"path/filepath"
)

// 或者使用更灵活的方式
func WriteFile(filePath string, content string) error {
	// 关键：创建父目录（如果不存在）
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}

	// 创建或打开文件（如果不存在则创建，存在则清空）
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败 %s: %w", filePath, err)
	}
	defer file.Close()

	// 写入内容
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入内容失败: %w", err)
	}

	// 确保数据写入磁盘
	return file.Sync()
}
