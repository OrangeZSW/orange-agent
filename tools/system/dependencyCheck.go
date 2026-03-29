package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
)

var DependencyCheckTool = common.BaseTool{
	Name:        "dependency_check",
	Description: "检查依赖包版本和更新情况",
	Parameters: map[string]interface{}{
		"check_outdated": map[string]interface{}{
			"type":        "boolean",
			"description": "是否检查过时的依赖",
		},
		"check_vulns": map[string]interface{}{
			"type":        "boolean",
			"description": "是否检查安全漏洞",
		},
	},
	Call: handlerDependencyCheck,
}

func handlerDependencyCheck(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
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
		out, err := cmd.CombinedOutput()
		if err != nil {
			output += fmt.Sprintf("过时依赖检查失败: %v\n", err)
		} else {
			output += "过时依赖:\n" + string(out) + "\n"
		}
	}

	if params.CheckVulns {
		cmd := exec.Command("govulncheck", "./...")
		out, err := cmd.CombinedOutput()
		if err != nil {
			output += fmt.Sprintf("安全漏洞检查失败: %v\n", err)
		} else {
			output += "安全漏洞检查:\n" + string(out)
		}
	}

	if output == "" {
		output = "未执行任何检查。请指定 check_outdated 或 check_vulns 为 true。"
	}

	return output, nil
}
