package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"path/filepath"
	"os/exec"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0 // file has execute permission 
}

func findInPath(cmdName string) string {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return ""
	}

	paths := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, dir := range paths {
		fullPath := filepath.Join(dir, cmdName)
		if isExecutable(fullPath) {
			return fullPath
		}
	}
	return ""
}

func main() {
	for {
	        fmt.Print("$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		command = strings.TrimSpace(command)

		if command == "exit" {
			os.Exit(0)
		}
		
		parts := strings.Fields(command)
		if len(parts) == 0 {
			continue
		}

		if parts[0] == "echo" {
			fmt.Println(strings.Join(parts[1:], " "))
			continue
		}

		if parts[0] == "type" {
			if len(parts) < 2 {
				continue
			}
			cmdName := parts[1]
			if cmdName == "echo" || cmdName == "exit" || cmdName == "type" {
				fmt.Printf("%s is a shell builtin\n", cmdName)
			} else {
				// search in PATH
				execPath := findInPath(cmdName)
				if execPath != "" {
					fmt.Printf("%s is %s\n", cmdName, execPath)
				} else {
					fmt.Printf("%s: not found\n", cmdName)
				}	
			}
			continue
		}

		execPath := findInPath(parts[0])
		if execPath != "" {
			// execute
			cmd := exec.Command(execPath, parts[1:]...)
			cmd.Args = parts
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			err := 	cmd.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr	, "Error executing command:", err)
			}
		} else {
			fmt.Println(parts[0] + ": command not found")
		}	
	}
}
