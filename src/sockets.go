package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/textproto"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type anySocket interface {
	write(string) error
	read() (string, error)
	close() error
}

type webSocket struct {
	rw     *bufio.ReadWriter
	closer func() error
}

func (wss *webSocket) read() (string, error) {
	var data []byte
	var err error
	if data, err = wsutil.ReadClientText(wss.rw); err != nil {
		return "", err
	}
	return string(data), nil
}

func (wss *webSocket) write(s string) error {
	if err := wsutil.WriteServerText(wss.rw, []byte(s)); err != nil {
		return err
	}
	return wss.rw.Flush()
}

func (wss *webSocket) close() error {
	msg := make([]byte, 2)
	// STATUS_NORMAL = 1000
	binary.BigEndian.PutUint16(msg, 1000)
	err1 := wsutil.WriteServerMessage(wss.rw, ws.OpClose, msg)
	err2 := wss.rw.Flush()
	err3 := wss.closer()
	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}
	return err3
}

type tcpSocket struct {
	rw     *bufio.ReadWriter
	closer func() error
	tpr    *textproto.Reader
}

func (tcp *tcpSocket) read() (s string, err error) {
	return tcp.tpr.ReadLine()
}

func (tcp *tcpSocket) write(s string) error {
	if _, err := tcp.rw.WriteString(s + "\r\n"); err != nil {
		return err
	}
	return tcp.rw.Flush()
}

func (tcp *tcpSocket) close() error {
	return tcp.closer()
}

func makeSocket(conn net.Conn) (anySocket, error) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	makeWebSocket := func() (anySocket, error) {
		_, err := ws.Upgrade(rw)
		if err != nil {
			conn.Write([]byte(fmt.Sprintf("%v\r\n\r\n", err)))
			return nil, err
		}
		fmt.Println("Upgrade: TCPSocket -> WebSocket")
		rw.Flush()
		return &webSocket{rw, conn.Close}, nil
	}
	makeTCPSocket := func() (anySocket, error) {
		tp := textproto.NewReader(rw.Reader)
		return &tcpSocket{rw, conn.Close, tp}, nil
	}
	if isHTTP(rw.Reader) {
		return makeWebSocket()
	}
	return makeTCPSocket()
}

func isHTTP(r *bufio.Reader) bool {
	http := []byte("GET ")
	head, err := r.Peek(4)
	return err == nil && bytes.Equal(http, head)
}
