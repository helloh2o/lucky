package oor

import (
	"bytes"
	"lucky/cmm/utils"
	"lucky/log"
	"testing"
)

func TestNewXORCipher(t *testing.T) {
	cipher := NewXORCipher(utils.RandString(10))
	painText := bytes.Repeat([]byte("hello world"), 8192)
	encrypt := cipher.Encode(painText)
	log.Debug(string(encrypt))
	dencrypt := cipher.Decode(encrypt)
	log.Debug(string(dencrypt))
}
