package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func Log(v ...interface{}) {
	log.Printf("%s", fmt.Sprint(v))
}

func waitIncomingMessage(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			Log("Read err", err)
			os.Exit(1)
			return
		}
		Log(strings.TrimSpace(line))

	}
}

func sendFile(fname string, conn net.Conn) {
	Log("Send file", fname, "to server")
	var err error
	file, err := os.Open(fname)
	if err != nil {
		Log("File open err", err)
		return
	}
	fi, err := file.Stat()
	if err != nil {
		Log("File get stat err", err)
		return
	}
	Log("File size:", fi.Size())
	conn.Write([]byte("/file " + fname + " " + strconv.FormatInt(fi.Size(), 10) + "\r\n"))

	defer file.Close()
	n, err := io.Copy(conn, file)
	if err != nil {
		Log("Send file err", err)
	}
	Log(n, "bytes sent")
}

func main() {
	name := flag.String("name", "Foo", "Nickname for chat")
	host := flag.String("host", "127.0.0.1", "Server host")
	port := flag.String("port", "7777", "Server port")
	flag.Parse()
	Log("Client start")
	conn, err := net.Dial("tcp", *host+":"+*port)
	if err != nil {
		Log("Connection err", err)
		return
	}
	defer conn.Close()
	conn.Write([]byte("/name " + *name + "\n"))
	go waitIncomingMessage(conn)
	reader := bufio.NewReader(os.Stdin)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			Log("Console input err", err)

		}
		if len(msg) > 5 && msg[:5] == "/send" {
			fname := strings.TrimSpace(strings.Split(msg, " ")[1])
			sendFile(fname, conn)
			continue
		}
		conn.Write([]byte(msg))
	}
}
