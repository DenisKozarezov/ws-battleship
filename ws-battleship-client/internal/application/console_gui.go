package application

import (
	"os"
	"os/exec"
	"runtime"
)

func runCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func clearTerminal() {
	switch runtime.GOOS {
	case "darwin", "linux":
		runCmd("clear")
	case "windows":
		runCmd("cmd", "/c", "cls")
	default:
		runCmd("clear")
	}
}
