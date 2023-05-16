package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <rootname> <command>\n", os.Args[0])
		os.Exit(1)
	}

	//ルートディレクトリの変更
	syscall.Chroot(os.Args[1])
	syscall.Chdir("/")

	cmd := exec.Command(os.Args[2])
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create new namespaces for the container
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
		os.Exit(1)
	}
}
