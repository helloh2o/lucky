package iduck

type Encrypt interface {
	Encode(bs []byte) []byte
	Decode(bs []byte) []byte
}

type Processor interface {
	SetBigEndian()
	GetBigEndian() bool
	SetEncrypt(enc Encrypt)
	OnReceivedPackage(IConnection, []byte)
	WarpMsg(interface{}) (error, []byte)
	RegisterHandler(id int, entity interface{}, handle func(args ...interface{}))
}
