package inet

import (
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"io"
	"lucky-day/core/duck"
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
					logrus.Error("tcp write %v", err)
					break
				}
			}
			// write over or error
			_ = conn.Close()
			logrus.Debugf("Conn %s <=> %s closed.", tc.Conn.LocalAddr(), tc.Conn.RemoteAddr())
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
			logrus.Errorf("TCPConn read message head error %s", err.Error())
			return
		}
		var ln uint16
		if tc.processor.GetBigOrder() {
			ln = binary.BigEndian.Uint16(bf[:2])
		} else {
			ln = binary.LittleEndian.Uint16(bf[:2])
		}
		if ln < 1 || ln > 2048 {
			logrus.Errorf("TCPConn message length %d invalid", ln)
			return
		}
		// read data
		_, err = io.ReadFull(tc, bf[:ln])
		if err != nil {
			logrus.Errorf("TCPConn read data err %s", err.Error())
			return
		}
		// clean
		_ = tc.SetDeadline(time.Time{})
		// throw out the msg
		tc.processor.OnReceivedMsg(tc, bf[:ln])
	}
}

func (tc *TCPConn) WriteMsg(message interface{}) {
	err, pkg := tc.processor.OnWarpMsg(message)
	if err != nil {
		logrus.Error("OnWarpMsg package error %s", err)
	} else {
		select {
		case tc.writeChan <- pkg:
		default:
			logrus.Error(" =============== Drop message, write chan is full  %d  =============== ", len(tc.writeChan))
		}
	}
}
