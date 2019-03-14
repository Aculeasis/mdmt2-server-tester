package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/gobwas/ws/wsutil"
)

// Server class
type Server struct {
	mutex   sync.Mutex
	wg      sync.WaitGroup
	connect bool
	work    bool
	restart bool
	con     anySocket
	parser  Parser
	srv     net.Listener
	args    *argParse
}

func (server *Server) makeServer() {
	var err error
	if server.srv, err = net.Listen("tcp", server.args.Address()); err != nil {
		panic(err)
	}
	fmt.Println("GO!")
	fmt.Println("")
}

// RunForever Server method
func (server *Server) RunForever() {
	server.parser = Parser{token: server.args.token}
	server.makeServer()
	server.work = true
	go server.loop()
	shell := Shell{srv: server}
	shell.RunForever()
	server.join()
}

func (server *Server) join() {
	server.wg.Wait()
}

func (server *Server) loop() {
	server.wg.Add(1)
	defer server.wg.Done()
	for {
		server.run()
		if server.restart {
			server.work = true
			server.restart = false
			server.args.HelloServer()
			server.makeServer()
		} else {
			break
		}
	}
	fmt.Println("")
	fmt.Println("BYE!")
	fmt.Println("")
}

func (server *Server) run() {
	defer server.srv.Close()
	for server.work {
		if c, err := server.srv.Accept(); err != nil {
			if server.work {
				fmt.Printf("Connecting error: %s\n", err)
			}
		} else {
			fmt.Printf("Connected %s ...\n", c.RemoteAddr())

			server.con, err = makeSocket(c)
			if err == nil {
				server.connParser()
			} else {
				fmt.Printf("Connection %s error: %v\n", c.RemoteAddr(), err)
				c.Close()
			}
			fmt.Printf("Disconnected %s.\n\n", c.RemoteAddr())
			server.con = nil

		}
	}
}

func (server *Server) connParser() {
	defer server.Close()
	server.connect = true
	server.parser.stage = 0
	for server.connect && server.work {
		if line, err := server.con.read(); err != nil {
			if _, ok := err.(wsutil.ClosedError); !ok && server.connect {
				fmt.Printf("Read error: %s\n", err)
			}
			return
		} else if line == "" {
			return
		} else if result, ok := server.parser.Parse(line); ok {
			server.Send(result)
		}
	}
}

// Close Server method
func (server *Server) Close() {
	server.connect = false
	defer server.mutex.Unlock()
	server.mutex.Lock()
	if server.con != nil {
		server.con.close()
	}
}

// Exit Server method
func (server *Server) Exit() {
	server.work = false
	server.Close()
	server.srv.Close()
}

func (server *Server) reload() {
	server.restart = true
	server.Exit()
}

// Send Server method
func (server *Server) Send(line string) {
	defer server.mutex.Unlock()
	server.mutex.Lock()
	if !server.connect || server.con == nil {
		fmt.Println("send -> no clients")
		return
	}
	if err := server.con.write(line); err != nil {
		fmt.Printf("Sending error: %s\n", err)
		return
	}
	fmt.Printf("send -> %s\n", line)
}
