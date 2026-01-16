package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

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
				fmt.Printf("%s: not found\n", cmdName)
			}
			continue
		}

		fmt.Println(command + ": command not found")
	}
}
