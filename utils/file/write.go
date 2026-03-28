package file

import "os"

// 或者使用更灵活的方式
func WriteFile(filePath string, content string) error {
	// 创建或打开文件（如果不存在则创建，存在则清空）
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close() // 确保文件关闭

	// 写入内容
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	// 确保数据写入磁盘
	return file.Sync()
}
