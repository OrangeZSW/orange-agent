package system

import (
	"context"
	"orange-agent/common"
	"os/exec"
)

type BuildTools struct {
	common.BaseTool
}

func (b *BuildTools) Name() string {
	return "build_tools"
}

func (b *BuildTools) Description() string {
	return "build tools"
}

func (b *BuildTools) Call(ctx context.Context, input string) (string, error) {
	cmd := exec.Command("go", "build", "orange-agent")
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), nil
	}
	return err.Error(), nil
}

func (b *BuildTools) Parameters() interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}
