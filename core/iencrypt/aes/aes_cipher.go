package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"github.com/helloh2o/lucky/log"
)

// AESCipher one of Encryptor implement
type AESCipher struct {
	key []byte
}

// NewAESCipher return a AESCipher
func NewAESCipher(key string) *AESCipher {
	size := len(key) / 8
	if size < 2 {
		log.Error("incorrect key, need 16(aes-128), 24(aes-192), 32(aes-256) length string")
		return nil
	} else if size > 4 {
		size = 4
	}
	key = key[:size*8]
	return &AESCipher{key: []byte(key)}
}

// Decode src
func (aes *AESCipher) Decode(src []byte) []byte {
	encrypt, err := aesDeCrypt(src, aes.key)
	if err != nil {
		log.Error("Aes Decode error %s", err)
		return src
	}
	return encrypt
}

// Encode src
func (aes *AESCipher) Encode(src []byte) []byte {
	encrypt, err := aesEcrypt(src, aes.key)
	if err != nil {
		log.Error("Aes Encode error %s", err)
		return src
	}
	return encrypt
}

//PKCS7 填充模式
func pKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	//Repeat()函数的功能是把切片[]byte{byte(padding)}复制padding个，然后合并成新的字节切片返回
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//填充的反向操作，删除填充字符串
func pKCS7UnPadding(origData []byte) ([]byte, error) {
	//获取数据长度
	length := len(origData)
	if length == 0 {
		return nil, errors.New("pKCS7UnPadding error")
	} else {
		//获取填充字符串长度
		unpadding := int(origData[length-1])
		//截取切片，删除填充字节，并且返回明文
		return origData[:(length - unpadding)], nil
	}
}

//实现加密
func aesEcrypt(origData []byte, key []byte) ([]byte, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块的大小
	blockSize := block.BlockSize()
	//对数据进行填充，让数据长度满足需求
	origData = pKCS7Padding(origData, blockSize)
	//采用AES加密方法中CBC加密模式
	blocMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	//执行加密
	blocMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//实现解密
func aesDeCrypt(cypted []byte, key []byte) ([]byte, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块大小
	blockSize := block.BlockSize()
	//创建加密客户端实例
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(cypted))
	//这个函数也可以用来解密
	blockMode.CryptBlocks(origData, cypted)
	//去除填充字符串
	origData, err = pKCS7UnPadding(origData)
	if err != nil {
		return nil, err
	}
	return origData, err
}
