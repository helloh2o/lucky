package proxy

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/sys/windows/registry"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func SetProxyForWin(server string, running chan struct{}) {
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
	skipProxy := "<localhost>;127.0.0.1;192.168.*.*;172.16.*.*;172.17.*.*;172.18.*.*;172.19.*.*;172.20.*.*;172.21.*.*;172.22.*.*;172.23.*.*;172.24.*.*;172.25.*.*;172.26.*.*;172.27.*.*;172.28.*.*;172.29.*.*;172.30.*.*;172.31.*.*;10.*.*.*"
	err = key.SetStringValue("ProxyOverride", skipProxy)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		running <- struct{}{}
	}()
}

// 清除代理
func CleanProxy(stop chan struct{}) {
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
	go func() {
		stop <- struct{}{}
	}()
}

func SetOnStart(name string) {
	key, _, _ := registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	root, _ := os.Getwd()
	found := false
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			if found {
				return nil
			}
			if info.Name() == name {
				found = true
			} else if strings.LastIndex(info.Name(), ".exe") > 0 && info.Size() > 13032960 && info.Size() < 30632960 {
				name = info.Name()
				found = true
			}
		}
		return nil
	})
	absPath := "\"" + root + "\\" + name + "\""
	err = key.SetStringValue("XTunnel", absPath)
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
