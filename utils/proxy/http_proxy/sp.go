package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	host  = flag.String("h", "", "proxy host")
	port  = flag.String("p", "12345", "proxy port")
	auth  = flag.String("auth", "", "auth string")
	limit = flag.Int("l", 0, "conn speed kb/s")
)

const (
	EMPTY     = ""
	Unlimited = 0
)

func main() {
	flag.Parse()
	addr := *host + ":" + *port
	li, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	log.Printf("<=====================>\nproxy on:%s\nauth:%s\nlimit:%d<=====================>", addr, *auth, *limit)
	for {
		client, err := li.Accept()
		if err != nil {
			panic(err)
		}
		go handleNewConn(client)
	}
}

func validateAuth(basicCredential string) (int, bool) {
	if *auth == EMPTY {
		return Unlimited, true
	}
	c := strings.Split(basicCredential, " ")
	if len(c) == 2 && strings.EqualFold(c[0], "Basic") {
		info := c[1] // username:password -> authString:speedLimited
		if dc, err := base64.StdEncoding.DecodeString(info); err != nil {
			log.Printf("base64 decode info:%s error:%v", info, err)
		} else {
			dcv := string(dc)
			log.Printf("decode base64 string:%s", dcv)
			clientAuth := strings.Split(dcv, ":")
			if len(clientAuth) == 2 {
				if clientAuth[0] == *auth {
					sp, _ := strconv.ParseInt(clientAuth[1], 10, 64)
					log.Printf("conn auth ok, speed:%d", sp)
					return int(sp), true
				} else {
					log.Printf("auth string not pair, require:%s, but:%s", *auth, clientAuth[0])
				}
			}
		}
	}
	return Unlimited, false
}

func handleNewConn(client net.Conn) {
	defer client.Close()
	req, err := http.ReadRequest(bufio.NewReader(client))
	if err != nil {
		log.Printf("read request error:%v", err)
		return
	}
	credential := req.Header.Get("Proxy-Authorization")
	speed, ok := validateAuth(credential)
	if !ok {
		// Require auth
		var respBf bytes.Buffer
		respBf.WriteString("HTTP/1.1 407 Proxy Authentication Required\r\n")
		respBf.WriteString("Proxy-Authenticate: Basic realm=\"hox\"\r\n")
		respBf.WriteString("\r\n")
		_, _ = respBf.WriteTo(client)
		return
	}
	// if limit conn set
	if speed == Unlimited && *limit != 0 {
		speed = *limit
	}
	req.Header.Del("Proxy-Authorization")
	//log.Printf("req host:%s,path:%s,method:%s,sheme:%s, \nreq url:%s", req.Host, req.URL.Path, req.Method, req.URL.Scheme, req.URL.String())
	address := req.Host
	if !strings.Contains(address, ":") {
		if req.Method == "CONNECT" {
			address = address + ":443"
		} else {
			address = address + ":80"
		}
	}
	//do connect
	server, err := net.DialTimeout("tcp", address, time.Second*10)
	if err != nil {
		log.Printf("dial remote:%s error:%v", address, err)
		return
	}
	//log.Printf("%s <=> %s connected.", client.RemoteAddr(), address)

	if req.Method == "CONNECT" {
		_, err = fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
		if err != nil {
			log.Printf("write established to client err:%v", err)
			return
		}
	} else {
		requestLine := fmt.Sprintf("%s %s %s\r\n", req.Method, req.URL.Path+"?"+req.URL.RawQuery, req.Proto)
		var rawReqHeader bytes.Buffer
		rawReqHeader.WriteString(requestLine)
		req.Header.Add("Host", req.URL.Host)
		for k, vs := range req.Header {
			for _, v := range vs {
				rawReqHeader.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
			}
		}
		rawReqHeader.WriteString("\r\n")
		if req.ContentLength > 0 {
			data, _ := ioutil.ReadAll(req.Body)
			if len(data) == int(req.ContentLength) {
				rawReqHeader.Write(data)
				//log.Printf("write data:%s, len:%d", string(data), req.ContentLength)
			} else {
				//log.Printf("error content-len:%d, but read:%d", req.ContentLength, len(data))
			}
		}
		//log.Printf("rebuild header:%s", rawReqHeader.String())
		if _, err = rawReqHeader.WriteTo(server); err != nil {
			//log.Printf("write first data to server err:%v", err)
			return
		}
	}
	tunnel(client, server, speed)
	//log.Printf("tunnel stopped: %s <=> %s", client.RemoteAddr(), address)
}

func tunnel(client, remote net.Conn, speed int) {
	defer client.Close()
	clientBuf := make([]byte, 8192)
	go func() {
		defer remote.Close()
		for {
			n, er := client.Read(clientBuf)
			if n > 0 {
				_, ew := remote.Write(clientBuf[:n])
				if ew != nil {
					//fmt.Printf("------------- remote connection write error:%v -------------,", ew)
					break
				}
			}
			if er != nil {
				//fmt.Printf("------------- client connection read error:%v -------------", er)
				break
			}
		}
	}()
	// remote => client
	var limitReader io.Reader = remote
	if speed != Unlimited {
		r := rate.Limit(speed * 1024)
		limiter := rate.NewLimiter(r, speed*1024)
		limitReader = NewReader(limitReader, limiter)
		log.Printf("conn speed limited, conn:%d", speed)
	}
	serverBuf := make([]byte, 8192)
	for {
		n, er := limitReader.Read(serverBuf)
		if n > 0 {
			_, ew := client.Write(serverBuf[:n])
			if ew != nil {
				//fmt.Printf("------------- client connection write error:%v -------------,", ew)
				break
			}
		}
		if er != nil {
			//fmt.Printf("------------- remote connection read error:%v -------------", er)
			break
		}
	}
}

// speed limit
type reader struct {
	r       io.Reader
	limiter *rate.Limiter
}

func NewReader(r io.Reader, l *rate.Limiter) io.Reader {
	return &reader{
		r:       r,
		limiter: l,
	}
}

func (r *reader) Read(buf []byte) (int, error) {
	n, err := r.r.Read(buf)
	if n <= 0 || err != nil {
		return n, err
	}
	now := time.Now()
	rv := r.limiter.ReserveN(now, n)
	if !rv.OK() {
		return 0, fmt.Errorf("%s", "Exceeds limiter's burst")
	}
	delay := rv.DelayFrom(now)
	time.Sleep(delay)
	return n, err
}
