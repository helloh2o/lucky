package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/helloh2o/lucky/core/inet"
	"github.com/helloh2o/lucky/example/comm/msg"
	"io/ioutil"
	"math/big"
	"os"
)

func main() {
	msg.SetEncrypt(msg.Processor)
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	err = ioutil.WriteFile("k.key", keyPEM, os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("c.cert", certPEM, os.ModePerm)
	if err != nil {
		panic(err)
	}
	pem, err := tls.LoadX509KeyPair("./c.cert", "./k.key")
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{pem},
		NextProtos:   []string{"quic-hello-example"},
	}
	if s, err := inet.NewQUICServer("localhost:2024", msg.Processor, tlsConfig); err != nil {
		panic(err)
	} else {
		err = s.Run()
	}
}
