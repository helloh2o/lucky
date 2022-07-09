package proxy

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/sys/windows/registry"
	"log"
)

// 设置代理
func SetProxyForWin(server string) {
	key, _, _ := registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.ALL_ACCESS)
	defer key.Close()
	err := key.SetBinaryValue("ProxyEnable", IntToBytes(1))
	if err != nil {
		log.Fatal(err)
	}
	err = key.SetStringValue("ProxyServer", server)
	if err != nil {
		log.Fatal(err)
	}
}

// 清除代理
func CleanProxy() {
	key, _, _ := registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.ALL_ACCESS)
	defer key.Close()
	err := key.SetBinaryValue("ProxyEnable", IntToBytes(0))
	if err != nil {
		log.Fatal(err)
	}
	err = key.SetStringValue("ProxyServer", "")
	if err != nil {
		log.Fatal(err)
	}
}

func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}
