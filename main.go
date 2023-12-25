package main

import (
	"log"
	"os"
	"os/exec"
)

func createRootDir() (string, error) {
    jail := "jail"
    command :=  exec.Command("mkdir", "-p", jail)
    err := command.Run()
    if err != nil {
        log.Printf("Failed to create a temp directory %v", err)
        return "",  err
    }
    return jail, nil
}

func copyDir(command string, jail string) {
    mkDirC := exec.Command("mkdir", "-p",  jail + "/usr/local/bin")
    mkDirC.Stdout = os.Stdout
    mkDirC.Stderr = os.Stderr
    
    err := mkDirC.Run()
    if err !=nil {
        log.Printf("Failed on mkdr -p %v", jail)
        panic("Failed on mkdr -p " + jail)
    }
    
    cp := exec.Command("cp", command, jail + "/usr/local/bin")
    cp.Stdout = os.Stdout
    cp.Stderr = os.Stderr
    cp.Stdin = os.Stdin
    
    err = cp.Run()
    if err !=nil {
        log.Printf("Failed on cp %v /usr/local/bin", command)
        panic("Failed on cp %v /usr/local/bin" + jail)
    }
}

func runCommand(jail string, command string, args []string) error {
    chRootArgs := []string {jail, command}
    chRootArgs = append(chRootArgs, args...)

    cmd := exec.Command("chroot", chRootArgs...)
    cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

    err := cmd.Run()
    if err != nil {
        log.Printf("failed to run chroot cmd: %v, args %v", command, args)
        return err
    }
    return nil
}

// Usage: ./mydocker run <image> <command> <arg1> <arg2> ...
func main() {
    log.Printf("statrting the command")
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
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
