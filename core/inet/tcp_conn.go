package inet

import (
	"encoding/binary"
	"io"
	"lucky-day/core/iduck"
	"lucky-day/log"
	"net"
	"time"
)

type TCPConn struct {
	net.Conn
	// 缓存队列
	writeQueue chan []byte
	readQueue  chan []byte
	// 消息处理器
	processor iduck.Processor
}

func NewTcpConn(conn net.Conn, processor iduck.Processor, maxQueueSize int) *TCPConn {
	if processor == nil || conn == nil {
		return nil
	}
	tc := &TCPConn{
		Conn:       conn,
		writeQueue: make(chan []byte, maxQueueSize),
		processor:  processor,
		// 单个缓存100个为处理的包
		readQueue: make(chan []byte, maxQueueSize),
	}
	// write q
	go func() {
		for pkg := range tc.writeQueue {
			// read over
			if pkg == nil {
				break
			}
			_, err := tc.Write(pkg)
			if err != nil {
				log.Error("tcp write %v", err)
				break
			}
		}
		// write over or error
		_ = conn.Close()
		log.Release("Conn %s <=> %s closed.", tc.Conn.LocalAddr(), tc.Conn.RemoteAddr())
	}()
	// read q
	go func() {
		for pkg := range tc.readQueue {
			// read over
			if pkg == nil {
				break
			}
			// the package
			tc.processor.OnReceivedPackage(tc, pkg)
		}
	}()
	return tc
}

func (tc *TCPConn) ReadMsg() {
	defer func() {
		tc.readQueue <- nil
		tc.writeQueue <- nil
	}()
	bf := make([]byte, 2048)
	// 第一个包默认5秒
	timeout := time.Second * 5
	for {
		_ = tc.SetReadDeadline(time.Now().Add(timeout))
		// read length
		_, err := io.ReadAtLeast(tc, bf[:2], 2)
		if err != nil {
			log.Error("TCPConn read message head error %s", err.Error())
			return
		}
		var ln uint16
		if tc.processor.GetBigEndian() {
			ln = binary.BigEndian.Uint16(bf[:2])
		} else {
			ln = binary.LittleEndian.Uint16(bf[:2])
		}
		if ln < 1 || ln > 2048 {
			log.Error("TCPConn message length %d invalid", ln)
			return
		}
		// read data
		_, err = io.ReadFull(tc, bf[:ln])
		if err != nil {
			log.Error("TCPConn read data err %s", err.Error())
			return
		}
		// clean
		_ = tc.SetDeadline(time.Time{})
		// write to cache queue
		select {
		case tc.readQueue <- append(make([]byte, 0), bf[:ln]...):
		default:
			log.Error("TCPConn read queue overflow err %s", err.Error())
			return
		}
		// after first pack | check heartbeat
		timeout = time.Second * 15
	}
}

func (tc *TCPConn) WriteMsg(message interface{}) {
	err, pkg := tc.processor.WarpMsg(message)
	if err != nil {
		log.Error("OnWarpMsg package error %s", err)
	} else {
		select {
		case tc.writeQueue <- pkg:
		default:
			log.Error(" =============== Drop message, write chan is full  %d  =============== ", len(tc.writeQueue))
		}
	}
}

func (tc *TCPConn) Close() error {
	return tc.Conn.Close()
}
