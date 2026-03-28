package common

type FileNode struct {
	Name     string
	IsDir    bool
	Path     string
	Children []*FileNode
}
