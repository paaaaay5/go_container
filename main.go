package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")
	}
}

func run() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s <rootfs path> <cgroup name> <command>\n", os.Args[0])
		os.Exit(1)
	}
	fmt.Printf("Running %v \n", os.Args[4])
	//go run main.go child [command...]
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//新たなnamespaceを作成
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error starting the command - %s\n", err)
		os.Exit(1)
	}
}

func child() {
	fmt.Printf("Child Running %v \n", os.Args[4])
	rootfs := os.Args[2]
	cgroupName := os.Args[3]
	cg(cgroupName)
	//rootディレクトリとカレントディレクトリをrootfsに設定
	syscall.Chroot(rootfs)
	syscall.Chdir("/")
	//新たなホストネームを設定
	syscall.Sethostname([]byte("container"))
	//procをマウント
	syscall.Mount("proc", "proc", "proc", 0, "")
	syscall.Mount("thing", "mytemp", "tmpfs", 0, "")
	cmd := exec.Command(os.Args[4], os.Args[5:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}

func cg(cgroupName string) {
	pid := os.Getpid()
	cgroupPath := filepath.Join("/sys/fs/cgroup/cpu", cgroupName)
	if err := os.Mkdir(cgroupPath, 0755); err != nil && !os.IsExist(err) {
		fmt.Printf("Error creating cgroup - %s\n", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(filepath.Join(cgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		fmt.Printf("Error adding the process to the cgroup - %s\n", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(filepath.Join(cgroupPath, "cpu.cfs_quota_us"), []byte("200000"), 0644); err != nil {
		fmt.Printf("Error setting the CPU quota - %s\n", err)
		os.Exit(1)
	}
}
