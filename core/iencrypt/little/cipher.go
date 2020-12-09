package little

import "lucky/log"

type Cipher struct {
	// 编码用的密码
	encodePassword *password
	// 解码用的密码
	decodePassword *password
}

// 加密原数据
func (cipher *Cipher) Encode(bs []byte) {
	for i, v := range bs {
		bs[i] = cipher.encodePassword[v]
	}
}

// 解码加密后的数据到原数据
func (cipher *Cipher) Decode(bs []byte) {
	for i, v := range bs {
		bs[i] = cipher.decodePassword[v]
	}
}

var CipherX *Cipher

// 新建一个编码解码器
func InitCipher(pw string) {
	encodePassword, err := ParsePassword(pw)
	if err != nil {
		log.Fatal("Init cipher password error %v", err)
	}
	decodePassword := &password{}
	for i, v := range encodePassword {
		encodePassword[i] = v
		decodePassword[v] = byte(i)
	}
	CipherX = &Cipher{
		encodePassword: encodePassword,
		decodePassword: decodePassword,
	}
}

// 新建一个编码解码器
func NewCipher(encodePassword *password) *Cipher {
	decodePassword := &password{}
	for i, v := range encodePassword {
		encodePassword[i] = v
		decodePassword[v] = byte(i)
	}
	return &Cipher{
		encodePassword: encodePassword,
		decodePassword: decodePassword,
	}
}
