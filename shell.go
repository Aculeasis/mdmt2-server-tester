package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Shell class
type Shell struct {
	srv  *Server
	work bool
}

// RunForever Shell method
func (shell *Shell) RunForever() {
	reader := bufio.NewReader(os.Stdin)
	shell.work = true
	defer shell.srv.Exit()
	for shell.work && shell.srv.work {
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("ERROR: %s", err)
			break
		}
		text = strings.TrimRight(text, "\r\n")
		if len(text) == 0 {
			continue
		}
		shell.parseLine(text)
	}
}

func (shell *Shell) parseLine(text string) {
	if text == "close" {
		shell.srv.Close()
	} else if text == "exit" {
		shell.work = false
	} else if text == "ping" {
		if ping, ok := makePingRequest(); ok {
			shell.srv.Send(ping)
		}
	} else if strings.HasPrefix(text, "token ") {
		shell.srv.parser.token = strings.SplitN(text, " ", 2)[1]
		fmt.Printf("New token: \"%s\"\n", shell.srv.parser.token)
	} else {
		shell.srv.Send(text)
	}
}
