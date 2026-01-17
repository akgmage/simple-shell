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

func parseCommand(input string) []string {
	var args []string
	var currentArg strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < len(input); i++ {
		char := input[i]

		if char == '\'' && !inDoubleQuote {
			// toggle single quote mode only
			inSingleQuote = !inSingleQuote
		} else if char == '"' && !inSingleQuote {
			// toggle double quote mode only
			inDoubleQuote = !inDoubleQuote	
		} else if char == ' ' || char == '\t' {
			if inSingleQuote || inDoubleQuote {
				// preserve whitespace since its inside quotes
				currentArg.WriteByte(char)
			} else {
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
			}
		} else {
			// wriite regular character
			currentArg.WriteByte(char)
		}
	}
	// add the last arg if any
	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
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
		
		parts := parseCommand(command)
		if len(parts) == 0 {
			continue
		}

		if parts[0] == "echo" {
			fmt.Println(strings.Join(parts[1:], " "))
			continue
		}

		if parts[0] == "pwd" {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error getting current directory:", err)
			} else {
				fmt.Println(cwd)
			}
			continue
		}

		if parts[0] == "cd" {
			if len(parts) < 2 {
				fmt.Fprintln(os.Stderr, "cd: missing argument")
				continue
			} 

			targetDir := parts[1]
			
			if targetDir == "~" {
				homeDir := os.Getenv("HOME")
				if homeDir == "" {
					fmt.Fprintln(os.Stderr, "cd: HOME not set")
					continue
				}
				targetDir = homeDir
			}

			err := os.Chdir(targetDir)
			if err != nil {
				fmt.Printf("cd: %s: No such file or directory\n", targetDir)
			}
			continue
		}

		if parts[0] == "type" {
			if len(parts) < 2 {
				continue
			}
			cmdName := parts[1]
			if cmdName == "echo" || cmdName == "exit" || cmdName == "type" || cmdName == "pwd" {
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
