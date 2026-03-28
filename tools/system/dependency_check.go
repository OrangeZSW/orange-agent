package system

import (
	"orange-agent/common"
	"os/exec"
)

var DependencyCheckTool = common.BaseTool{
	Name:        "dependency_check",
	Description: "检查依赖包版本和更新情况",
	Parameters: map[string]interface{}{
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
		"required": [],
	},
}

func CheckDependencies(checkOutdated, checkVulns bool) (string, error) {
	var output string
	var err error

	if checkOutdated {
		cmd := exec.Command("go", "list", "-m", "-u", "-u=all", "all")
		out, _ := cmd.CombinedOutput()
		output += "过时依赖:\n" + string(out) + "\n"
	}

	if checkVulns {
		cmd := exec.Command("govulncheck", "./...")
		out, _ := cmd.CombinedOutput()
		output += "安全漏洞检查:\n" + string(out)
	}

	if output == "" {
		output = "未执行任何检查。请指定检查选项。"
	}

	return output, err
}
