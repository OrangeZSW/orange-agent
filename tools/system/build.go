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
	cmd := exec.Command("bash", "-c", "chmod +x ./build.sh && ./build.sh")
	output, err := cmd.CombinedOutput()
	
	result := "Build output:\n" + string(output)
	if err != nil {
		result += "\n\nBuild failed with error: " + err.Error()
		return result, err
	}
	result += "\n\nBuild completed successfully!"
	return result, nil
}

func (b *BuildTools) Parameters() interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}