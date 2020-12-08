package duck

type Encrypt interface {
	Encode(bs []byte)
	Decode(bs []byte)
}

type Processor interface {
	SetBigEndian(big bool)
	GetBigEndian() bool
	SetEncrypt(enc Encrypt)
	OnReceivedMsg(IConnection, []byte)
	WarpMsg(interface{}) (error, []byte)
	RegisterHandler(id int, entity interface{}, handle func(args ...interface{}))
}
