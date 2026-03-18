package persistence

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

type AOF struct {
	// Implementation details would go here
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
	c    chan []string  // in memory queue for AOF
	wg   sync.WaitGroup // ensures file doesn't close before worker flushes
}

func NewAof(path string) (*AOF, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	aof := &AOF{
		file: file,
		rd:   bufio.NewReader(file),
		c:    make(chan []string, 1024),
	}

	aof.wg.Add(1)

	return aof, nil
}

func (a *AOF) worker() {
	defer a.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case args, ok := <-a.c:
			if !ok {
				a.file.Sync()
				return
			}

			resp := fmt.Sprintf("*%d\r\n", len(args))
			for _, arg := range args {
				resp += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
			}

			a.file.WriteString(resp)

		case <-ticker.C:
			a.file.Sync()
		}
	}
}

func (a *AOF) Close() error {
	close(a.c)
	a.wg.Wait()
	return a.file.Close()
}

func (a *AOF) Write(args []string) error {
	a.c <- args
	return nil
}
