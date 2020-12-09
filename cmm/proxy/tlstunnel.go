package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"time"
)

func main() {
	Run("goproxy.cn:443", ":1080")
}

type backend struct {
	Host     string
	Port     string
	Hostname string
	TLS      bool
	Insecure bool
}

var (
	dialer  = &net.Dialer{Timeout: 6 * time.Second}
	l       net.Listener
	nodes   = make(map[string]string)
	running = false
	bk      = new(backend)
)

func Run(remote, local string) {
	bk.TLS = true
	if running {
		switchbk(remote)
	}
	go initListener(remote, local)
	select {}
}
func SwitchNode(remote string) bool {
	switchbk(remote)
	return true
}

func switchbk(remote string) bool {
	host, port, err := net.SplitHostPort(remote)
	if err == nil {
		bk.Host = host
		bk.Port = port
		return true
	}
	return false
}

func initListener(remote, local string) {
	var err error
	if switchbk(remote) {
		l, err = net.Listen("tcp", local)
		if err != nil {
			log.Printf("Listen error :: %v\n", err)
			return
		} else {
			log.Printf("Listen on %s\n", local)
			running = true
		}
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println(err)
				l.Close()
				break
			}
			go handleConn(conn)
		}
	}
}

func handleConn(conn net.Conn) {
	var c net.Conn
	var err error

	remote := net.JoinHostPort(bk.Host, bk.Port)
	if bk.TLS {
		config := &tls.Config{
			ServerName:         bk.Host,
			InsecureSkipVerify: false,
		}
		c, err = tls.DialWithDialer(dialer, "tcp", remote, config)
	} else {
		c, err = dialer.Dial("tcp", remote)
	}

	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}

	pipeAndClose(conn, c)
}

func pipeAndClose(c1, c2 net.Conn) {
	defer c1.Close()
	defer c2.Close()

	ch := make(chan struct{}, 2)
	go func() {
		io.Copy(c1, c2)
		ch <- struct{}{}
	}()

	go func() {
		io.Copy(c2, c1)
		ch <- struct{}{}
	}()
	<-ch
}
