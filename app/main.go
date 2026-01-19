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
		
		if char == '\\' && !inSingleQuote {
			if inDoubleQuote {
				if i+1 < len(input) {
					nextChar := input[i+1]
					if nextChar == '"' || nextChar == '\\' || nextChar == '$' || nextChar == '`' || nextChar == '\n' {
						i++
						currentArg.WriteByte(nextChar)
					} else {
						currentArg.WriteByte('\\')
					}
				}
				continue
			} else {
				if i+1 < len(input) {
					i++
					nextChar := input[i]
					currentArg.WriteByte(nextChar)
				}	
			continue
			}
		}
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

func openOutputFile(filename string, appendMode bool) (*os.File, error) {
	if appendMode {
		return os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	return os.Create(filename)
}

func handlePwd() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting current directory:", err)
	} else {
		fmt.Println(cwd)
	}
}

func parseRedirection(parts []string) ([]string, string, string, bool, bool) {
	var cmdArgs []string
	var outputFile string
	var errorFile string
	var appendMode bool
	var appendErrorMode bool

	for i := 0; i < len(parts); i++ {
		arg := parts[i]
		if arg == "2>>" {
			if i+1 < len(parts) {
				errorFile = parts[i+1]
				appendErrorMode = true // append stderr
				i++
			}
		} else if strings.HasPrefix(arg, "2>>") {
			errorFile = arg[3:]
			appendErrorMode = true
		} else if arg == ">>" || arg == "1>>" {
			if i+1 < len(parts) {
				outputFile = parts[i+1]
				i++
				appendMode = true
			}
		} else if strings.HasPrefix(arg, ">>") || strings.HasPrefix(arg, "1>>") {
			if strings.HasPrefix(arg, ">>") {
				outputFile = arg[2:]
			} else {
				outputFile = arg[3:]
			}
			appendMode = true
		} else if arg == "2>" {
			// next arg is stderr file
			if i+1 < len(parts) {
				errorFile = parts[i+1]
				i++
				appendErrorMode = false
			}
		} else if strings.HasPrefix(arg, "2>") {
			errorFile = arg[2:]
			appendErrorMode = false
		} else if arg == ">" || arg == "1>" {
			if i+1 < len(parts) {
				outputFile = parts[i+1]
				i++
				appendMode = false
			}
		} else if strings.HasPrefix(arg, ">") || strings.HasPrefix(arg, "1>") {
			if strings.HasPrefix(arg, "1>") {
				outputFile = arg[2:]
			} else {
				outputFile = arg[1:]
			}
			appendMode = false
		} else {
			cmdArgs = append(cmdArgs, arg)
		}
	}
	return cmdArgs, outputFile, errorFile, appendMode, appendErrorMode
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
		// handle redirection
		cmdParts, outputFile, errorFile, appendMode, appendErrorMode := parseRedirection(parts)
		if len(cmdParts) == 0 {
			continue
		}

		if cmdParts[0] == "echo" {
			output := strings.Join(cmdParts[1:], " ")

			if outputFile != "" {
				var file *os.File
				var err error

				if appendMode {
					// append mode
					file, err = os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				} else {
					// overwrite mode
					file, err = os.Create(outputFile)
				}
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error creating file:", err)
					continue
				}
				fmt.Fprintln(file, output)
				file.Close()
			} else {
				fmt.Println(output)
			}

			if errorFile != "" {
				var file *os.File
				var err error

				if appendErrorMode {
					file, err = os.OpenFile(errorFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				} else {
					file, err = os.Create(errorFile)
				}
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error creating file:", err)
					continue
				}
				file.Close()
			}
			continue
		}

		if cmdParts[0] == "pwd" {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error getting current directory:", err)
			} else {
				fmt.Println(cwd)
			}
			continue
		}

		if cmdParts[0] == "cd" {
			if len(cmdParts) < 2 {
				fmt.Fprintln(os.Stderr, "cd: missing argument")
				continue
			} 

			targetDir := cmdParts[1]
			
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

		if cmdParts[0] == "type" {
			if len(cmdParts) < 2 {
				continue
			}
			cmdName := cmdParts[1]
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

		execPath := findInPath(cmdParts[0])
		if execPath != "" {
			// execute
			cmd := exec.Command(execPath, cmdParts[1:]...)
			cmd.Args = cmdParts

			var stdoutDest *os.File
			
			if outputFile != "" {
				var err error
				
				if appendMode {
					// append modee
					stdoutDest, err = os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				} else {
					// overwrite mode
					stdoutDest, err = os.Create(outputFile)
				}
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error creating file:", err)
					continue
				}

				defer stdoutDest.Close()

			} else {
				stdoutDest = os.Stdout
			}

			var stderrDest *os.File
			if errorFile != "" {
				if appendErrorMode {
					// append to stderr file
					stderrDest, err = os.OpenFile(errorFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				} else {
					// overwrite stderr file
					stderrDest, err = os.Create(errorFile)
				}

				if err != nil {
					fmt.Println(os.Stderr, "Error creating file:", err)
					continue
				}
				
				defer stderrDest.Close()
			} else {
				stderrDest = os.Stderr
			}

			cmd.Stdout = stdoutDest
			cmd.Stderr = stderrDest
			cmd.Stdin = os.Stdin

			cmd.Run()
		} else {
			fmt.Println(cmdParts[0] + ": command not found")
		}	
	}
}
