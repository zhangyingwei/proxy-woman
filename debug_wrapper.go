//go:build ignore

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func main() {
	// 设置环境变量
	os.Setenv("CGO_ENABLED", "1")
	os.Setenv("CGO_LDFLAGS", "-framework UniformTypeIdentifiers")
	
	// 构建并运行应用
	cmd := exec.Command("go", "run", "-tags", "dev", "-gcflags", "all=-N -l", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.Sys().(syscall.WaitStatus).ExitStatus())
		}
		os.Exit(1)
	}
}
