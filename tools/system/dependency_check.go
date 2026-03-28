package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
)

type DependencyCheckTools struct {
	common.BaseTool
}

func (d *DependencyCheckTools) Name() string {
	return "dependency_check"
}

func (d *DependencyCheckTools) Description() string {
	return "检查依赖包版本和更新情况"
}

func (d *DependencyCheckTools) Call(ctx context.Context, input string) (string, error) {
	var params struct {
		CheckOutdated bool `json:"check_outdated"`
		CheckVulns    bool `json:"check_vulns"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	var output string

	if params.CheckOutdated {
		cmd := exec.Command("go", "list", "-m", "-u", "all")
		out, _ := cmd.CombinedOutput()
		output += "过时依赖:\n" + string(out) + "\n"
	}

	if params.CheckVulns {
		cmd := exec.Command("govulncheck", "./...")
		out, _ := cmd.CombinedOutput()
		output += "安全漏洞检查:\n" + string(out)
	}

	if output == "" {
		output = "未执行任何检查。请指定 check_outdated 或 check_vulns 为 true。"
	}

	return output, nil
}

func (d *DependencyCheckTools) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"check_outdated": map[string]interface{}{
				"type":        "boolean",
				"description": "是否检查过时的依赖",
			},
			"check_vulns": map[string]interface{}{
				"type":        "boolean",
				"description": "是否检查安全漏洞",
			},
		},
		"required": []string{},
	}
}
