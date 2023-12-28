//go:build linux
// +build linux

package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func createRootDir() (string, error) {
	jail := "jail"
	command := exec.Command("mkdir", "-p", jail)
	err := command.Run()
	if err != nil {
		log.Printf("Failed to create a temp directory %v", err)
		return "", err
	}
	return jail, nil
}

func copyDir(command string, jail string) {
	mkDirC := exec.Command("mkdir", "-p", jail+"/usr/local/bin")
	mkDirC.Stdout = os.Stdout
	mkDirC.Stderr = os.Stderr

	err := mkDirC.Run()
	if err != nil {
		log.Printf("Failed on mkdir -p %v", jail)
		panic("Failed on mkdir -p " + jail)
	}

	// locaty full path of command
	path, err := exec.LookPath(command)
	if err != nil {
		log.Printf("Failed to locate full path of %v on the host os", command)
	}

	cp := exec.Command("cp", path, jail+"/usr/local/bin")
	cp.Stdout = os.Stdout
	cp.Stderr = os.Stderr
	cp.Stdin = os.Stdin

	err = cp.Run()
	if err != nil {
		log.Printf("Failed on cp %v /usr/local/bin", command)
		panic("Failed on cp %v /usr/local/bin" + jail)
	}
}

func runCommand(jail string, command string, args []string) error {
	chRootArgs := []string{jail, command}
	chRootArgs = append(chRootArgs, args...)

	cmd := exec.Command("chroot", chRootArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Usage: ./mydocker run <image> <command> <arg1> <arg2> ...
func main() {
	log.Printf("starting the command")
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	jail, err := createRootDir()
	if err != nil {
		panic("could not create a target directory, stopping execution")
	}
	copyDir(command, jail)
	err = runCommand(jail, command, args)

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
