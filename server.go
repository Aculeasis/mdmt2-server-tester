package main

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
	"sync"
)

// Server class
type Server struct {
	mutex   sync.Mutex
	wg      sync.WaitGroup
	connect bool
	work    bool
	restart bool
	con     net.Conn
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
	server.wg.Add(1)
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
	defer func() {
		server.srv.Close()
		server.wg.Done()
	}()
	for server.work {
		if c, err := server.srv.Accept(); err != nil {
			if server.work {
				fmt.Printf("Connecting error: %s\n", err)
			}
		} else {
			fmt.Printf("Connected %s ...\n", c.RemoteAddr())
			server.con = c
			server.connParser()
			fmt.Printf("Disconnected %s.\n\n", c.RemoteAddr())
			server.con = nil

		}
	}
}

func (server *Server) connParser() {
	defer server.Close()
	server.connect = true
	server.parser.stage = 0
	reader := bufio.NewReader(server.con)
	tp := textproto.NewReader(reader)
	for server.connect && server.work {
		if line, err := tp.ReadLine(); err != nil {
			if server.connect {
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
		server.con.Close()
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
	buffer := []byte(line + "\r\n")
	bufferLen := len(buffer)
	count := 0

	if !server.connect {
		fmt.Println("send -> no clients")
		return
	}
	defer server.mutex.Unlock()
	server.mutex.Lock()
	for count < bufferLen {
		if !server.connect {
			return
		} else if send, err := server.con.Write(buffer[count:]); err != nil {
			fmt.Printf("Sending error: %s\n", err)
			return
		} else {
			count += send
		}
	}
	fmt.Printf("send -> %s\n", line)
}
