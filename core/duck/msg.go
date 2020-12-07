package duck

type Encrypt interface {
	Encode(bs []byte)
	Decode(bs []byte)
}

type Processor interface {
	SetBytesOrder(big bool)
	GetBigOrder() bool //big
	SetEncrypt(enc Encrypt)
	GetEncrypt() Encrypt
	OnReceivedMsg(IConnection, []byte)
	OnWarpMsg(interface{}) (error, []byte)
	RegisterHandler(id int, entity interface{}, handle func(args ...interface{}))
}
