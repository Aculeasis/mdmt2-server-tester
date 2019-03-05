package main

import (
	"flag"
	"fmt"
)

type argParse struct {
	ip    string
	port  uint
	token string
}

func newArgs() argParse {
	args := argParse{}
	flag.StringVar(&args.ip, "ip", "127.0.0.1", "Server IP (127.0.0.1)")
	flag.UintVar(&args.port, "port", 7575, "Server Port (7575)")
	flag.StringVar(&args.token, "token", "hello", "auth token (\"hello\")")
	flag.Parse()
	return args
}

func (args *argParse) Address() string {
	return fmt.Sprintf("%s:%d", args.ip, args.port)
}

func (args *argParse) Hello() {
	fmt.Printf("Server %s\n", args.Address())
	fmt.Printf("token: %s\n\n", args.token)
}
