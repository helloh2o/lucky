package chatnode

import (
	"github.com/helloh2o/lucky"
)

var testChatRoom lucky.INode

// GetRoom get net node
func GetRoom() lucky.INode {
	if testChatRoom == nil {
		testChatRoom = lucky.NewBroadcastNode()
		testChatRoom.Serve()
	}
	return testChatRoom
}
