package route

import "github.com/helloh2o/lucky/im/server/immsg"

func InitHandler() {
	Processor.RegisterHandler(PeerMsg, &immsg.PeerMsg{}, func(args ...interface{}) {

	})
	Processor.RegisterHandler(GroupMsg, &immsg.PeerGroupMsg{}, func(args ...interface{}) {

	})
}
