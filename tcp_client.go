package main

import (
	"errors"
	"log"
	"net"
	"time"
)

type TCPClient struct {
	ch            chan int
	TCPAddress    *net.TCPAddr
	Conn          *net.TCPConn
	ReconnectWait time.Duration
	MaxReconnect  int
}

func NewTCPClient(address string, reconnectWait time.Duration, maxReconnect int) (*TCPClient, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	return &TCPClient{
		ch:            make(chan int, 1),
		TCPAddress:    tcpAddr,
		Conn:          conn,
		ReconnectWait: reconnectWait,
		MaxReconnect:  maxReconnect,
	}, nil
}

// based on
// https://github.com/ehsangolshani/data-forwarder/blob/ffcfcf54834e1fc09665a4fca493b70c8d50f112/filebeat/tcp.go
func (t *TCPClient) Write(data []byte) (int, error) {

	if len(t.ch) > 0 {
		return 0, errors.New("message discarded, connection is not available")
	}

	n, err := t.Conn.Write(data)
	if err != nil {
		if len(t.ch) > 0 {
			return 0, err
		}

		t.ch <- 0
		err = t.reconnect()
		<-t.ch

		return 0, err
	}

	return n, nil
}

func (t *TCPClient) reconnect() error {
	_ = t.Conn.Close()

	for i := 0; i < t.MaxReconnect; i++ {
		conn, err := net.DialTCP("tcp", nil, t.TCPAddress)
		if err != nil {
			log.Printf("[%s], reconnecting ...", err)
			time.Sleep(t.ReconnectWait)
		}

		t.Conn = conn
		return nil
	}

	return errors.New("failed to reconnect to tcp listener")
}
