package server

import (
	"fmt"
	"io"
	"net"

	"github.com/kartheek0107/GoDis/internal/protocol"
	"github.com/kartheek0107/GoDis/internal/store"
)

type Server struct {
	addr  string
	store *store.Store
}

func NewServer(addr string, store *store.Store) *Server {
	return &Server{
		addr:  addr,
		store: store,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	fmt.Println("📡 GoDis is listening on %s\n", s.addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	parser := protocol.NewParser(conn)
	for {
		cmd, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error parsing command:", err)
			break
		}
		fmt.Println("Received command:", cmd)

		if len(cmd) == 0 {
			continue
		}

		switch cmd[0] {
		case "SET":
			if len(cmd) != 3 {
				conn.Write([]byte("Wrong no.of arguments"))
			}
			s.store.Set(cmd[1], cmd[2])
			conn.Write([]byte("+OK\r\n"))
		case "GET":
			if len(cmd) != 2 {
				conn.Write([]byte("Wrong no.of arguments"))
			}
			value, found := s.store.Get(cmd[1])
			if !found {
				conn.Write([]byte("$-1\r\n"))
			} else {
				response := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
				conn.Write([]byte(response))
			}
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "COMMAND DOCS":
			conn.Write([]byte("*0\r\n"))
		default:
			conn.Write([]byte("-ERR\r\n"))

		}
	}
}
