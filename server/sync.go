package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/pranavgore09/lkv/config"
	"github.com/pranavgore09/lkv/core"
)

func readCommand(c io.ReadWriter) (*core.RedisCmd, error) {
	var buf []byte = make([]byte, 512)
	num_bytes_read, err := c.Read(buf[:])
	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buf[:num_bytes_read])
	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil
}

func respondError(c io.ReadWriter, e error) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", e)))
}

func sendResponse(c io.ReadWriter, cmd *core.RedisCmd) {

	err := core.EvalAndRespond(c, cmd)
	if err != nil {
		respondError(c, err)
	}
	// Echo Server OLD
	// byte_resp := []byte(cmd)
	// _, err := c.Write(byte_resp)
}

func Run() {
	var connection_str string = config.Host + ":" + strconv.Itoa(config.Port)

	log.Println("Config: ", connection_str)

	listener, err := net.Listen("tcp", connection_str)
	if err != nil {
		log.Println("Error: ", err)
		return
	}

	connection_count := 0

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection Error: ", err)
		}

		connection_count++
		log.Println("Current Connection Count = ", connection_count)

		for {
			cmd, err := readCommand(conn)

			if err != nil {
				connection_count -= 1
				conn.Close()
				log.Println("Client Disconnected: ", conn.RemoteAddr(), " connection count = ", connection_count)
				if err == io.EOF {
					break
				}
				log.Println(err)
			}

			log.Printf("Command = %q", cmd)

			sendResponse(conn, cmd)
			// if err != nil {
			// 	log.Println("Respond Error:", err)
			// }
		}
	}
}
