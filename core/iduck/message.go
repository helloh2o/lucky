package iduck

// Encryptor interface
type Encryptor interface {
	Encode(bs []byte) []byte
	Decode(bs []byte) []byte
}

// Processor interface
type Processor interface {
	SetBigEndian()
	GetBigEndian() bool
	SetEncryptor(enc Encryptor)
	OnReceivedPackage(interface{}, []byte)
	WarpMsg(interface{}) (error, []byte)
	RegisterHandler(id int, entity interface{}, handle func(args ...interface{}))
}
