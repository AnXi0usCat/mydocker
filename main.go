package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
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

	// create a chroot jail
	tmp, err := os.MkdirTemp(os.TempDir(), "*")
	if err != nil {
		log.Fatalf("Failed to create a new root directory %v", err)
	}
	defer os.RemoveAll(tmp)

	// find executable for the command we are tring to Run
	command, err = exec.LookPath(command)
	if err != nil {
		log.Fatalf("Failed to find executable of the command %v", err)
	}

	// copy the command executable to chroot jail
	commandChRoot := filepath.Join(tmp, filepath.Base(command))
	err = copyTo(commandChRoot, command)
	if err != nil {
		log.Fatalf("Failed to copy the command %v", err)
	}

	err = syscall.Chroot(tmp)
	if err != nil {
		log.Fatalf("Failed to create a new root %v", err)
	}

	commandChRoot = filepath.Join("/", filepath.Base(command))
	cmd := exec.Command(commandChRoot, args...)

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
