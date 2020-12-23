package aes

import (
	"lucky/log"
	"testing"
)

func TestAESCipher_Encode(t *testing.T) {
	cipher := NewAESCipher("BH1rStJwNP1YIvNIffffff")
	painText := []byte("hello � []world ��")
	encrypt := cipher.Encode(painText)
	log.Debug(string(encrypt))
	dencrypt := cipher.Decode(encrypt)
	log.Debug(string(dencrypt))
}
