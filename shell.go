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
	var cmd, value string
	splitted := strings.SplitN(text, " ", 2)
	cmd = splitted[0]
	isValue := len(splitted) == 2 && splitted[1] != ""
	if isValue {
		value = splitted[1]
		if value == " " {
			value = ""
		}
	}
	if cmd == "help" {
		shell.usage()
	} else if cmd == "close" {
		shell.srv.Close()
	} else if cmd == "exit" {
		shell.work = false
	} else if cmd == "ping" {
		if ping, ok := makePingRequest(); ok {
			shell.srv.Send(ping)
		}

	} else if cmd == "port" {
		if !isValue {
			fmt.Printf("Current port: \"%d\"\n", shell.srv.args.port)
		} else if portUint, err := strconv.ParseUint(value, 10, 16); err != nil {
			fmt.Printf("Wrong port \"%s\": \"%v\"\n", value, err)
		} else {
			shell.srv.args.port = uint(portUint)
			fmt.Printf("New port: \"%d\"\n", shell.srv.args.port)
			shell.srv.reload()
		}
	} else if cmd == "ip" {
		if !isValue {
			fmt.Printf("Current ip: \"%s\"\n", shell.srv.args.ip)
		} else {
			shell.srv.args.ip = value
			fmt.Printf("New ip: \"%s\"\n", shell.srv.args.ip)
			shell.srv.reload()
		}
	} else if cmd == "token" {
		if !isValue {
			fmt.Printf("Current token: \"%s\"\n", shell.srv.parser.token)
		} else {
			shell.srv.parser.token = value
			fmt.Printf("New token: \"%s\"\n", shell.srv.parser.token)
		}
	} else {
		if cmd == "remote_log" {
			shell.srv.parser.stage = 3
			text = cmd
		}
		shell.srv.Send(text)
	}
}
