package file

import (
	"bufio"
	"io"
	"orange-agent/common"
	"os"
	"path/filepath"
)

// 场景	推荐方法
// 小文件（<10MB）	os.ReadFile
// 按行处理	bufio.Scanner
// 大文件处理	bufio.Reader 分块读取
// JSON/CSV等格式	对应的 encoding 包
// 需要随机访问	os.File 配合 Seek

// 读取文件
func ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// 按行处理

func ReadLine(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// 大文件处理
func ReadBigFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buffer []byte
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		buffer = append(buffer, line...)
	}
	return buffer, nil
}

// 按块读取
func ReadBigFileByChunk(filePath string, chunkSize int) ([][]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var chunks [][]byte
	buffer := make([]byte, chunkSize)

	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// 只读取有效的字节数
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buffer[:n])
			chunks = append(chunks, chunk)
		}
	}

	return chunks, nil
}

// 随机访问
func ReadRandomAccess(filePath string, offset int64, length int64) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 跳转到指定位置
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// 读取指定长度的数据
	buffer := make([]byte, length)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buffer[:n], nil
}

func GetFileTree(rootPath string) ([]*common.FileNode, error) {
	var result []*common.FileNode

	// 读取目录内容
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == ".git" {
			continue
		}

		node := &common.FileNode{
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			Path:     filepath.Join(rootPath, entry.Name()),
			Children: []*common.FileNode{},
		}

		// 如果是目录，递归获取子文件
		if entry.IsDir() {
			children, err := GetFileTree(node.Path)
			if err != nil {
				// 如果无法读取子目录，可以选择继续或返回错误
				// 这里选择继续，只记录错误
				continue
			}
			node.Children = children
		}

		result = append(result, node)
	}

	return result, nil
}

func GetFileList(rootPath string) ([]common.FileInfo, error) {
	res := []common.FileInfo{}

	files, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, item := range files {
		fullPath := filepath.Join(rootPath, item.Name())

		if item.IsDir() {
			// 递归获取子目录中的文件
			children, err := GetFileList(fullPath)
			if err != nil {
				continue // 跳过无法读取的目录
			}
			res = append(res, children...)
		} else {
			// 只添加文件，不添加目录
			res = append(res, common.FileInfo{
				Name: item.Name(),
				Path: fullPath,
			})
		}
	}
	return res, nil
}
