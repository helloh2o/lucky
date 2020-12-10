package iduck

type Encryptor interface {
	Encode(bs []byte) []byte
	Decode(bs []byte) []byte
}

type Processor interface {
	SetBigEndian()
	GetBigEndian() bool
	SetEncrypt(enc Encryptor)
	OnReceivedPackage(IConnection, []byte)
	WarpMsg(interface{}) (error, []byte)
	RegisterHandler(id int, entity interface{}, handle func(args ...interface{}))
}
