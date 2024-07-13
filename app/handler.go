package main

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"SET":  set,
	"GET":  get,
	"ECHO": echo,
}
var SETs = map[string]string{}
var SETEXPIRYs = map[string]time.Time{}
var SETsMu = sync.RWMutex{}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "PONG"}
	}
	return Value{typ: "string", str: args[0].bulk}
}

func set(args []Value) Value {
	key := args[0].bulk
	value := args[1].bulk
	if len(args) == 4 {
		px := args[2].bulk
		if strings.ToUpper(px) != "PX" {
			return Value{typ: "error", str: "no such option " + px}
		}
		millisec, err := strconv.Atoi(args[3].bulk)
		if err != nil {
			return Value{typ: "error", str: "Invalid PX value"}
		}
		expiry := time.Now().Add(time.Millisecond * time.Duration(millisec))
		SETsMu.Lock()
		SETEXPIRYs[key] = expiry
		SETsMu.Unlock()
	}
	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()
	return Value{typ: "string", str: "OK"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'get' command"}
	}
	key := args[0].bulk
	SETsMu.Lock()
	expiry, ok := SETEXPIRYs[key]
	SETsMu.Unlock()
	if ok {
		if time.Now().After(expiry) {
			return Value{typ: "null"}
		}
	}
	SETsMu.Lock()
	value, ok := SETs[key]
	SETsMu.Unlock()
	if !ok {
		return Value{typ: "null"}
	}
	return Value{typ: "bulk", bulk: value}
}

func echo(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'echo' command"}
	}
	value := args[0].bulk
	return Value{typ: "bulk", bulk: value}
}
