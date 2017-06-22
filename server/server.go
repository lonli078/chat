package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"strconv"
	"io"
	"flag"
)

const BUF_SIZE = 1024 * 1024 * 20
const MAX_FILE_SIZE = 1024 * 1024 * 1024

func Log(v ...interface{}) {
	log.Printf("Server: %s", fmt.Sprint(v))
}

type Client struct {
	Name       string
	conn       net.Conn
	ClientList *ClientList
}

func (c *Client) Read() {
	reader := bufio.NewReader(c.conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			Log("Read err", err, c, c.Name)
			c.conn.Close()
			c.ClientList.Leave_ch <- c
			return
		}
		line = strings.TrimSpace(line)
		if len(line) < 1 {
			continue
		}
		if len(line) > 7 && line[:5] == "/name" {
			c.Name = strings.TrimSpace(strings.Split(line, " ")[1])
			continue
		}
		if len(line) > 6 && line[:5] == "/file" {
			params := strings.SplitN(line, " ", -1)
			Log(params)
			fname := params[1]
			size, _ := strconv.ParseInt(strings.SplitN(line, " ", -1)[2], 10, 64)
			if size > MAX_FILE_SIZE {
				c.Write("Max size 1 GB\r\n")
				c.conn.Close()
				c.ClientList.Leave_ch <- c
				return
			}
			err := c.getFile(fname, size)
			if err != nil {
				c.Write("File Server error\r\n")
				c.conn.Close()
				c.ClientList.Leave_ch <- c
				return
			}
			continue
		}
		c.ClientList.Broadcast(c, line)

	}
}

func (c *Client) Write(msg string) {
	_, err := c.conn.Write([]byte(msg))
	if err != nil {
		Log("Write err", err, c, c.Name)
	}
}

func (c *Client) getFile(fname string, size int64) (err error){
	var currentByte int64
	var n int
	buf := make([]byte, BUF_SIZE)
	reader := bufio.NewReader(c.conn)
	_ = os.Mkdir("tmp/", 0777)
	file, err := os.Create("tmp/" + strings.TrimSpace(fname))
	if err != nil {
		Log("Create file err", err)
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for err == nil || err != io.EOF {
		n, err = reader.Read(buf)
		if _, err2 := w.Write(buf[:n]); err2 != nil {
			Log("Write file err", err2)
			return err2
		}
		w.Flush()
		currentByte += int64(n)
		if currentByte >= size {
			break
		}
	}
	if err == io.EOF {
		return err
	}
	Log("File saved", fname, "size:", size)
	return
}

type ClientList struct {
	clients  []*Client
	Join_ch  chan net.Conn
	Leave_ch chan *Client
}

func (cl *ClientList) addClient(conn net.Conn) {
	newclient := &Client{Name: "", conn: conn, ClientList: cl}
	cl.clients = append(cl.clients, newclient)
	go newclient.Read()
	Log("Client list:", cl.clients)
}

func (cl *ClientList) removeClient(c *Client) {
	for i, client := range cl.clients {
		if client == c {
			cl.clients = append(cl.clients[:i], cl.clients[i+1:]...)
			Log("Client list:", cl.clients)
			return
		}
	}

}
func (cl *ClientList) Broadcast(c *Client, msg string) {
	Log("Send msg from client:", c.Name, "Message:", msg)
	for _, client := range cl.clients {
		if c != client {
			client.Write(c.Name + ": " + msg + "\r\n")
		}

	}
}

func (cl *ClientList) Listen() {
	for {
		select {
		case conn := <-cl.Join_ch:
			cl.addClient(conn)
		case c := <-cl.Leave_ch:
			cl.removeClient(c)
		}
	}
}

func NewClientList() *ClientList {
	cl := &ClientList{
		clients:  make([]*Client, 0),
		Join_ch:  make(chan net.Conn),
		Leave_ch: make(chan *Client),
	}
	go cl.Listen()
	return cl
}

func main() {
	host := flag.String("host", "0.0.0.0", "Server host")
	port := flag.String("port", "7777", "Server port")
	flag.Parse()
	Log("Server Start")
	cl := NewClientList()
	listener, err := net.Listen("tcp", *host + ":" + *port)
	if err != nil {
		Log("Server bind err", err)
		return
	}
	for {
		conn, _ := listener.Accept()
		cl.Join_ch <- conn
	}
}
