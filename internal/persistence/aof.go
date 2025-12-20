package persistence

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type AOF struct {
	// Implementation details would go here
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}

func NewAof(path string) (*AOF, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return &AOF{
		file: file,
		rd:   bufio.NewReader(file),
	}, nil
}

func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Close()
}

func (a *AOF) Write(args []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	resp := fmt.Sprintf("*%d\r\n", len(args))
	for _, arg := range args {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}
	_, err := a.file.WriteString(resp)
	if err != nil {
		return err
	}
	return a.file.Sync()
}
