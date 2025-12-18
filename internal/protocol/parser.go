package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Parser struct {
	reader *bufio.Reader
}

func NewParser(rd io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(rd),
	}
}

// readLine reads until \r\n and returns the bytes without the suffix
func (p *Parser) readLine() ([]byte, error) {
	line, err := p.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(line) < 2 {
		return nil, errors.New("protocol error: line too short")
	}
	// Trim the \r\n from the end
	return line[:len(line)-2], nil
}

// readInt parses the number after a prefix (like *3 or $5)
func (p *Parser) readInt() (int, error) {
	line, err := p.readLine()
	if err != nil {
		return 0, err
	}
	// Convert everything after the first character (the prefix) to an int
	return strconv.Atoi(string(line[1:]))
}

func (p *Parser) Parse() ([]string, error) {
	prefix, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	// Skip any leading whitespace or newlines that might be in the buffer
	for prefix == '\r' || prefix == '\n' {
		prefix, err = p.reader.ReadByte()
		if err != nil {
			return nil, err
		}
	}

	// Standard Redis commands always start as an Array (*)
	if prefix != '*' {
		return nil, fmt.Errorf("expected '*', got %q", prefix)
	}

	// Put the * back so readInt can read the whole line
	p.reader.UnreadByte()
	count, err := p.readInt()
	if err != nil {
		return nil, err
	}

	args := make([]string, count)
	for i := 0; i < count; i++ {
		// Expect '$' for Bulk String
		bPrefix, err := p.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if bPrefix != '$' {
			return nil, fmt.Errorf("expected '$', got %q", bPrefix)
		}

		// Put $ back for readInt
		p.reader.UnreadByte()
		size, err := p.readInt()
		if err != nil {
			return nil, err
		}

		// Read the actual data (bulk string)
		data := make([]byte, size)
		_, err = io.ReadFull(p.reader, data)
		if err != nil {
			return nil, err
		}

		// CRITICAL: Consume the \r\n that follows the data block
		// We read exactly two bytes to avoid eating into the next command
		p.reader.ReadByte() // consumes \r
		p.reader.ReadByte() // consumes \n

		args[i] = string(data)
	}

	return args, nil
}
