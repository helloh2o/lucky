package route

import "github.com/helloh2o/lucky"

var Processor lucky.Processor

func JsonHandler() lucky.Processor {
	Processor = lucky.NewJSONProcessor()
	return Processor
}

func ProtobufHandler() lucky.Processor {
	Processor = lucky.NewPBProcessor()
	return Processor
}
