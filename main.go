package main

import (
	"log"
	"os"
	"os/exec"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	cmd := exec.Command(command, args...)
    
    cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
        log.Printf("Program failed %s", err.Error())
        switch err := err.(type) {
        case *exec.ExitError:
            os.Exit(err.ExitCode())
        default:
		    log.Printf("Child process exited abnormally: %v", err)
            os.Exit(255)
        }
	}
}
