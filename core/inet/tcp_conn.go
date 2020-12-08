package inet

import (
	"encoding/binary"
	"io"
	"lucky-day/core/duck"
	"lucky-day/log"
	"net"
	"time"
)

type TCPConn struct {
	net.Conn
	writeChan chan []byte
	processor duck.Processor
}

func NewTcpConn(conn net.Conn, processor duck.Processor) *TCPConn {
	if processor == nil || conn == nil {
		return nil
	}
	tc := &TCPConn{
		Conn:      conn,
		writeChan: make(chan []byte, 100),
		processor: processor,
	}
	go func() {
		for {
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
			log.Debug("Conn %s <=> %s closed.", tc.Conn.LocalAddr(), tc.Conn.RemoteAddr())
		}
	}()
	return tc
}

func (tc *TCPConn) ReadMsg() {
	defer func() {
		tc.writeChan <- nil
	}()
	bf := make([]byte, 2048)
	for {
		_ = tc.SetReadDeadline(time.Now().Add(time.Second * 15))
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
		// throw out the msg
		tc.processor.OnReceivedMsg(tc, bf[:ln])
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
