//go:build linux

package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

const jail = "jail"
const cgroupPath = "/sys/fs/cgroup/"
const cgNameLen = 32
const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func root() (string, error) {
	err := os.MkdirAll(jail, os.FileMode(0777))
	if err != nil {
		log.Printf("Failed to create a temp directory %v", err)
		return "", err
	}
	return jail, nil
}

func name() string {
	s := rand.NewSource(time.Now().UnixMilli())
	r := rand.New(s)
	b := make([]byte, cgNameLen)

	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

func cgroup(cname string) error {
	pid := os.Getgid()

	cgPathName := cgroupPath + cname
	log.Printf("Creating a new cgroup for container with name: %v\n", cname)

	// create a new cgroup
	if err := os.Mkdir(cgPathName, 0755); err != nil {
		fmt.Printf("Error creating groups: %v\n", err)
		return err
	}

	// create a file with max pid limit
	pidsMaxPath := filepath.Join(cgPathName, "pids.max")
	if err := os.WriteFile(pidsMaxPath, []byte("20"), 0644); err != nil {
		fmt.Printf("Error creating pids.max file: %v\n", err)
		return err
	}
	// add acurrent process to the group
	cgroupProcsPath := filepath.Join(cgPathName, "cgroup.procs")
	if err := os.WriteFile(cgroupProcsPath, []byte(fmt.Sprintf("%v", pid)), 0644); err != nil {
		fmt.Printf("Error creating cgroup.procs file %v\n", err)
		return err
	}

	return nil
}

func run() error {
	image := os.Args[2]
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	fmt.Printf("Running command %v with args %v as %v\n", command, args, os.Getgid())

	jail, err := root()
	if err != nil {
		log.Printf("could not create a target directory, stopping execution")
		return err
	}
	download(image, jail)

	cmd := exec.Command("/proc/self/exe", append([]string{"child", image, command}, args...)...)

	// don't share PIDS's and hostname with the host
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	log.Printf("Syscall to clone PID's")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	// remove the working directory after child completes
	delete(jail)

	if err != nil {
		log.Printf("Encountered an error while doing `execute as a child` %s", err)
		return err
	}

	return nil
}

func child() error {
	fmt.Printf("Running command %v with args %v as %v\n", os.Args[3], os.Args[4:len(os.Args)], os.Getpid())

	// generate new cgroup name
	cname := name()
	// create a new cgroup
	err := cgroup(cname)
	if err != nil {
		log.Printf("Failed to create cgroup direcotry %s", err)
		return err
	}
	// isolate child process hostname
	syscall.Sethostname([]byte("container"))
	// isolate filsystem
	syscall.Chroot(jail)
	syscall.Chdir("/")
	// isolate the process PIDS
	syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(os.Args[3], os.Args[4:len(os.Args)]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	syscall.Unmount("proc", 0)

	if err != nil {
		return err
	}
	return nil
}

// Usage: ./mydocker run <image> <command> <arg1> <arg2> ...
func main() {
	log.Printf("starting the command")
	var err error

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		err = child()
	default:
		panic("Undefined command " + os.Args[1])
	}

	if err != nil {
		log.Printf("Program failed %s", err.Error())
		switch e := err.(type) {
		case *exec.ExitError:
			os.Exit(e.ExitCode())
		default:
			panic(err)
		}
	}
}
