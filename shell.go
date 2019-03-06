package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

// Shell class
type Shell struct {
	srv       *Server
	work      bool
	completer *readline.PrefixCompleter
	liner     *readline.Instance
}

func (shell *Shell) usage() {
	io.WriteString(shell.liner.Stderr(), "commands:\n")
	io.WriteString(shell.liner.Stderr(), shell.completer.Tree("    "))
}

// RunForever Shell method
func (shell *Shell) RunForever() {
	shell.completer = readline.NewPrefixCompleter(
		readline.PcItem("close"),
		readline.PcItem("exit"),
		readline.PcItem("ping"),
		readline.PcItem("token"),
		readline.PcItem("help"),
		readline.PcItem("ip"),
		readline.PcItem("port"),
	)
	var err error
	shell.liner, err = readline.NewEx(&readline.Config{AutoComplete: shell.completer})
	if err != nil {
		panic(err)
	}
	shell.work = true
	defer func() {
		shell.liner.Close()
		shell.srv.restart = false
		shell.srv.Exit()
	}()

	for shell.work && (shell.srv.work || shell.srv.restart) {
		line, err := shell.liner.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err != nil {
			break
		}
		line = strings.TrimLeft(line, " ")
		shell.parseLine(line)
	}
}

func (shell *Shell) parseLine(text string) {
	if text == "help" {
		shell.usage()
	} else if text == "close" {
		shell.srv.Close()
	} else if text == "exit" {
		shell.work = false
	} else if text == "ping" {
		if ping, ok := makePingRequest(); ok {
			shell.srv.Send(ping)
		}

	} else if text == "port" {
		fmt.Printf("Current port: \"%d\"\n", shell.srv.args.port)
	} else if strings.HasPrefix(text, "port ") {
		port := strings.SplitN(text, " ", 2)[1]
		if portUint, err := strconv.ParseUint(port, 10, 16); err != nil {
			fmt.Printf("Wrong port \"%s\": \"%v\"\n", port, err)
		} else {
			shell.srv.args.port = uint(portUint)
			fmt.Printf("New port: \"%d\"\n", shell.srv.args.port)
			shell.srv.reload()
		}

	} else if text == "ip" {
		fmt.Printf("Current ip: \"%s\"\n", shell.srv.args.ip)
	} else if strings.HasPrefix(text, "ip ") {
		shell.srv.args.ip = strings.SplitN(text, " ", 2)[1]
		fmt.Printf("New ip: \"%s\"\n", shell.srv.args.ip)
		shell.srv.reload()

	} else if text == "token" {
		fmt.Printf("Current token: \"%s\"\n", shell.srv.parser.token)
	} else if strings.HasPrefix(text, "token ") {
		shell.srv.parser.token = strings.SplitN(text, " ", 2)[1]
		fmt.Printf("New token: \"%s\"\n", shell.srv.parser.token)

	} else {
		shell.srv.Send(text)
	}
}
