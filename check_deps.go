package main
import (
	"fmt"
	"os/exec"
)
func main() {
	cmd := exec.Command("go", "mod", "tidy")
	output, err := cmd.CombinedOutput()
	fmt.Printf("Output: %s\nError: %v\n", output, err)
}
