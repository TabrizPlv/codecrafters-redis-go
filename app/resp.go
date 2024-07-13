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
	NULL    = '_'
)

type Resp struct {
	reader *bufio.Reader
}

type Value struct {
	typ   string  //type of RESP Value
	str   string  //data of STRING
	num   int     //data of INTEGER
	bulk  string  //data of BULK
	array []Value //data of ARRAY
}

type Writer struct {
	writer io.Writer
}

// Constructor for Resp
func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
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

func (value Value) marshal() []byte {
	switch value.typ {
	case "array":
		return value.marshalArray()
	case "bulk":
		return value.marshalBulk()
	case "string":
		return value.marshalString()
	case "null":
		return value.marshalNull()
	case "error":
		return value.marshalError()
	default:
		return []byte{}
	}
}

func (value Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, value.str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (value Value) marshalInteger() []byte {
	var bytes []byte
	bytes = append(bytes, INTEGER)
	bytes = append(bytes, strconv.Itoa(value.num)...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (value Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(value.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, value.bulk...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (value Value) marshalArray() []byte {
	var bytes []byte
	len := len(value.array)
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')
	for i := 0; i < len; i++ {
		data := value.array[i].marshal()
		bytes = append(bytes, data...)
	}
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (value Value) marshalNull() []byte {
	var bytes []byte
	bytes = append(bytes, NULL)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (value Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, value.str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (w *Writer) Write(value Value) error {
	var bytes = value.marshal()
	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
