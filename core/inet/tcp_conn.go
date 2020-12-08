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
	writeChan chan []byte
	processor iduck.Processor
}

func NewTcpConn(conn net.Conn, processor iduck.Processor) *TCPConn {
	if processor == nil || conn == nil {
		return nil
	}
	tc := &TCPConn{
		Conn:      conn,
		writeChan: make(chan []byte, 100),
		processor: processor,
	}
	go func() {
		for pkg := range tc.writeChan {
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
	return tc
}

func (tc *TCPConn) ReadMsg() {
	defer func() {
		tc.writeChan <- nil
		tc.processor.Close()
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
		// the package
		tc.processor.OnReceivedPackage(tc, bf[:ln])
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
		case tc.writeChan <- pkg:
		default:
			log.Error(" =============== Drop message, write chan is full  %d  =============== ", len(tc.writeChan))
		}
	}
}

func (tc *TCPConn) Close() error {
	return tc.Conn.Close()
}
