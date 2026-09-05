package server

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/kartheek0107/GoDis/internal/persistence"
	"github.com/kartheek0107/GoDis/internal/protocol"
	"github.com/kartheek0107/GoDis/internal/store"
)

type Server struct {
	addr  string
	store *store.Store
	aof   *persistence.AOF
}

func NewServer(addr string, store *store.Store, aof *persistence.AOF) *Server {
	return &Server{
		addr:  addr,
		store: store,
		aof:   aof,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	fmt.Println("📡 GoDis is listening on %s\n", s.addr)
	defer ln.Close()
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

		switch strings.ToUpper(cmd[0]) {
		case "SET":
			if len(cmd) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
				break
			}
			s.store.Set(cmd[1], cmd[2])
			if err := s.aof.Write(cmd); err != nil {
				fmt.Println("Error writing to AOF:", err)
				conn.Write([]byte("-ERR AOF write failed\r\n"))
				break
			}
			conn.Write([]byte("+OK\r\n"))
		case "GET":
			if len(cmd) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				break
			}
			value, found := s.store.Get(cmd[1])
			if !found {
				conn.Write([]byte("$-1\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)))
			}
		case "DEL":
			if len(cmd) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'del' command\r\n"))
				break
			}
			deleted := s.store.Delete(cmd[1])
			if deleted {
				if err := s.aof.Write(cmd); err != nil {
					fmt.Println("Error writing to AOF:", err)
				}
				conn.Write([]byte(":1\r\n"))
			} else {
				conn.Write([]byte(":0\r\n"))
			}
		case "EXISTS":
			if len(cmd) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'exists' command\r\n"))
				break
			}
			if s.store.Exists(cmd[1]) {
				conn.Write([]byte(":1\r\n"))
			} else {
				conn.Write([]byte(":0\r\n"))
			}
		// List Commands
		case "LPUSH":
			if len(cmd) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'lpush' command\r\n"))
				break
			}
			res := s.store.LPush(cmd[1], cmd[2:]...)
			if res == -1 {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				if err := s.aof.Write(cmd); err != nil {
					fmt.Println("Error writing to AOF:", err)
				}
				conn.Write([]byte(fmt.Sprintf(":%d\r\n", res)))
			}
		case "RPUSH":
			if len(cmd) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'rpush' command\r\n"))
				break
			}
			res := s.store.RPush(cmd[1], cmd[2:]...)
			if res == -1 {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				if err := s.aof.Write(cmd); err != nil {
					fmt.Println("Error writing to AOF:", err)
				}
				conn.Write([]byte(fmt.Sprintf(":%d\r\n", res)))
			}
		case "LPOP":
			if len(cmd) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'lpop' command\r\n"))
				break
			}
			val, ok := s.store.LPop(cmd[1])
			if !ok {
				conn.Write([]byte("$-1\r\n"))
			} else {
				if err := s.aof.Write(cmd); err != nil {
					fmt.Println("Error writing to AOF:", err)
				}
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
			}
		case "RPOP":
			if len(cmd) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'rpop' command\r\n"))
				break
			}
			val, ok := s.store.RPop(cmd[1])
			if !ok {
				conn.Write([]byte("$-1\r\n"))
			} else {
				if err := s.aof.Write(cmd); err != nil {
					fmt.Println("Error writing to AOF:", err)
				}
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
			}
		case "LRANGE":
			if len(cmd) != 4 {
				conn.Write([]byte("-ERR wrong number of arguments for 'lrange' command\r\n"))
				break
			}
			start, err1 := strconv.Atoi(cmd[2])
			stop, err2 := strconv.Atoi(cmd[3])
			if err1 != nil || err2 != nil {
				conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
				break
			}
			vals, ok := s.store.LRange(cmd[1], start, stop)
			if !ok {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf("*%d\r\n", len(vals))))
				for _, v := range vals {
					conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)))
				}
			}
		// Hash Commands
		case "HSET":
			if len(cmd) != 4 {
				conn.Write([]byte("-ERR wrong number of arguments for 'hset' command\r\n"))
				break
			}
			res := s.store.HSet(cmd[1], cmd[2], cmd[3])
			if res == -1 {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				if err := s.aof.Write(cmd); err != nil {
					fmt.Println("Error writing to AOF:", err)
				}
				conn.Write([]byte(fmt.Sprintf(":%d\r\n", res)))
			}
		case "HGET":
			if len(cmd) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'hget' command\r\n"))
				break
			}
			val, ok := s.store.HGet(cmd[1], cmd[2])
			if !ok {
				conn.Write([]byte("$-1\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
			}
		case "HGETALL":
			if len(cmd) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'hgetall' command\r\n"))
				break
			}
			hash, ok := s.store.HGetAll(cmd[1])
			if !ok {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf("*%d\r\n", len(hash)*2)))
				for k, v := range hash {
					conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n$%d\r\n%s\r\n", len(k), k, len(v), v)))
				}
			}
		// Set Commands
		case "SADD":
			if len(cmd) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'sadd' command\r\n"))
				break
			}
			res := s.store.SAdd(cmd[1], cmd[2:]...)
			if res == -1 {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				if err := s.aof.Write(cmd); err != nil {
					fmt.Println("Error writing to AOF:", err)
				}
				conn.Write([]byte(fmt.Sprintf(":%d\r\n", res)))
			}
		case "SISMEMBER":
			if len(cmd) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'sismember' command\r\n"))
				break
			}
			isMem, ok := s.store.SIsMember(cmd[1], cmd[2])
			if !ok {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else if isMem {
				conn.Write([]byte(":1\r\n"))
			} else {
				conn.Write([]byte(":0\r\n"))
			}
		case "SMEMBERS":
			if len(cmd) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'smembers' command\r\n"))
				break
			}
			members, ok := s.store.SMembers(cmd[1])
			if !ok {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf("*%d\r\n", len(members))))
				for _, m := range members {
					conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(m), m)))
				}
			}
		// Probabilistic Commands
		case "PFADD":
			if len(cmd) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'pfadd' command\r\n"))
				break
			}
			res, _ := s.store.PfAdd(cmd[1], cmd[2:]...)
			if res == -1 {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				if err := s.aof.Write(cmd); err != nil {
					fmt.Println("Error writing to AOF:", err)
				}
				conn.Write([]byte(fmt.Sprintf(":%d\r\n", res)))
			}
		case "PFCOUNT":
			if len(cmd) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'pfcount' command\r\n"))
				break
			}
			res, _ := s.store.PfCount(cmd[1])
			if res == -1 {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				if err := s.aof.Write(cmd); err != nil {
					fmt.Println("Error writing to AOF:", err)
				}
				conn.Write([]byte(fmt.Sprintf(":%d\r\n", res)))
			}
		case "BF.ADD":
			if len(cmd) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'bf.add' command\r\n"))
				break
			}
			res, _ := s.store.BfAdd(cmd[1], cmd[2])
			if res == -1 {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf(":%d\r\n", res)))
			}
		case "BF.EXISTS":
			if len(cmd) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'bf.exists' command\r\n"))
				break
			}
			res, _ := s.store.BfExists(cmd[1], cmd[2])
			if res == -1 {
				conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf(":%d\r\n", res)))
			}
		// Heatmap Commands
		case "HEATMAP.SET":
			if len(cmd) != 5 {
				conn.Write([]byte("-ERR wrong number of arguments for 'heatmap.set' command\r\n"))
				break
			}
			x, err1 := strconv.Atoi(cmd[2])
			y, err2 := strconv.Atoi(cmd[3])
			val, err3 := strconv.Atoi(cmd[4])
			if err1 != nil || err2 != nil || err3 != nil {
				conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
				break
			}
			s.store.HeatmapSet(cmd[1], x, y, val)
			if err := s.aof.Write(cmd); err != nil {
				fmt.Println("Error writing to AOF:", err)
			}
			conn.Write([]byte("+OK\r\n"))
		case "HEATMAP.GET":
			if len(cmd) != 4 {
				conn.Write([]byte("-ERR wrong number of arguments for 'heatmap.get' command\r\n"))
				break
			}
			x, err1 := strconv.Atoi(cmd[2])
			y, err2 := strconv.Atoi(cmd[3])
			if err1 != nil || err2 != nil {
				conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
				break
			}
			val, ok := s.store.HeatmapGet(cmd[1], x, y)
			if !ok {
				conn.Write([]byte("$-1\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf(":%d\r\n", val)))
			}
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "COMMAND":
			// Ignore COMMAND DOCS / COMMAND for now
			conn.Write([]byte("+OK\r\n"))
		default:
			conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", cmd[0])))
		}
	}
}
