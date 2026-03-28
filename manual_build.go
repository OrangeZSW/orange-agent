package main
import (
	"fmt"
	"os"
	"os/exec"
)
func main() {
	fmt.Println("Running go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("go mod tidy failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("\nRunning go build...")
	cmd2 := exec.Command("go", "build", "-o", "orange-agent", "main.go")
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	err = cmd2.Run()
	if err != nil {
		fmt.Printf("\nbuild failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("\nBuild completed successfully!")
}
