package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Resp struct {
	reader *bufio.Reader
}

type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

// Constructor for Resp
func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readTillEndOfLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		if b == '\r' {
			r.reader.ReadByte()
			return line, n, nil
		}
		n += 1
		line = append(line, b)
	}
}

func (r *Resp) readInteger() (i int, n int, err error) {
	line, n, err := r.readTillEndOfLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

func (r *Resp) read() (value Value, err error) {
	dataType, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch dataType {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(dataType))
		return Value{}, nil
	}
}

func (r *Resp) readArray() (value Value, err error) {
	value.typ = "array"
	value.array = make([]Value, 0)
	numElements, _, err := r.readInteger()
	if err != nil {
		return Value{}, err
	}
	for i := 0; i < numElements; i++ {
		arrayElement, err := r.read()
		if err != nil {
			return Value{}, err
		}
		value.array = append(value.array, arrayElement)
	}
	return value, nil
}

func (r *Resp) readBulk() (value Value, err error) {
	value.typ = "bulk"
	_, _, err = r.readInteger()
	if err != nil {
		return Value{}, err
	}
	line, _, err := r.readTillEndOfLine()
	if err != nil {
		return Value{}, err
	}
	value.bulk = string(line)
	return value, nil
}
