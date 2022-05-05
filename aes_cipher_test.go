package lucky

import (
	"github.com/helloh2o/lucky/log"
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

func TestAESCipher_24key(t *testing.T) {
	cipher := NewAESCipher("BH1rStJwNP1YIvNIffffff11")
	painText := []byte("hello � []world ��")
	encrypt := cipher.Encode(painText)
	log.Debug(string(encrypt))
	dencrypt := cipher.Decode(encrypt)
	log.Debug(string(dencrypt))
}

func TestAESCipher_36key(t *testing.T) {
	cipher := NewAESCipher("BH1rStJwNP1YIvNIffffff1122334455")
	painText := []byte("hello � []world ��")
	encrypt := cipher.Encode(painText)
	log.Debug(string(encrypt))
	dencrypt := cipher.Decode(encrypt)
	log.Debug(string(dencrypt))
}