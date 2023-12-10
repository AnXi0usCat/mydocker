package main

import (
	"io"
	"log"
	"os"
	"os/exec"
)

func copyTo(to, from string) error {
	ffi, err := os.Stat(from)
	if err != nil {
		log.Printf("failed to get the file stats %v", err)
		return err
	}

	fromFile, err := os.Open(from)
	if err != nil {
		log.Printf("failed to open the origanal command file %v", err)
		return err
	}
	defer fromFile.Close()

	toFile, err := os.Create(to)
	if err != nil {
		log.Printf("failed to create a destination file for the command %v", err)
		return err
	}
	defer toFile.Close()

	_, err = io.Copy(toFile, fromFile)
	if err != nil {
		log.Printf("failed to copy contents of the command file %v", err)
		return err
	}

	err = os.Chmod(to, ffi.Mode())
	if err != nil {
		log.Printf("failed to grant permissions to  target command file %v", err)
		return err
	}

	return nil
}

// Usage: ./mydocker run <image> <command> <arg1> <arg2> ...
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
