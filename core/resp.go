package core

import (
	"errors"
	"fmt"
	"io"
	"log"
)

func readSimpleString(data []byte) (string, int, error) {

	// first char is '+'
	pos := 1

	for ; data[pos] != '\r'; pos++ {
	}

	return string(data[1:pos]), pos + 1 + 1, nil
	// 							current + \r \n

}

func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

func readInt64(data []byte) (int64, int, error) {
	pos := 1
	var value int64 = 0
	for ; data[pos] != '\r'; pos++ {
		b := data[pos] - '0'
		value = value*10 + int64(b)
	}
	return value, pos + 1 + 1, nil
}

func readLength(data []byte) (int, int) {

	pos, length := 0, 0

	for pos = range data {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 1 + 1
		}
		length = length*10 + int(b-'0')
	}

	return 0, 0

}

func readBulkString(data []byte) (string, int, error) {

	pos := 1

	len, delta := readLength(data[pos:])
	pos += delta

	end := pos + len

	return string(data[pos:end]), end + 1 + 1, nil

}

func readArray(data []byte) (interface{}, int, error) {

	pos := 1

	count, delta := readLength(data[pos:])
	pos += delta

	var elements []interface{} = make([]interface{}, count)

	for i := range elements {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elements[i] = elem
		pos += delta
	}
	return elements, pos, nil
}

func DecodeOne(data []byte) (interface{}, int, error) {

	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}

	log.Printf("Input = %q", data)

	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case ':':
		return readInt64(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	}
	return nil, 0, nil
}

func Decode(data []byte) (interface{}, error) {

	if len(data) == 0 {
		return nil, errors.New("no data")
	}

	value, _, err := DecodeOne(data)

	return value, err

}

func DecodeArrayString(data []byte) ([]string, error) {
	value, err := Decode(data)
	log.Println("from Decode(): ", value, err)
	if err != nil {
		return nil, err
	}

	ts := value.([]interface{})

	tokens := make([]string, len(ts))
	for i := range tokens {
		tokens[i] = ts[i].(string)
	}

	log.Printf("Tokens: %s", tokens)

	return tokens, nil

}

func EvalAndRespond(c io.ReadWriter, cmd *RedisCmd) error {

	switch cmd.Cmd {
	case "PING":
		return evalPING(c, cmd.Args)
	default:
		return evalPING(c, cmd.Args)
	}
}

func evalPING(c io.ReadWriter, args []string) error {
	var b []byte

	if len(args) >= 2 {
		return errors.New("wrong number of arguments")
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := c.Write(b)
	return err
}

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	}
	return []byte{}
}
