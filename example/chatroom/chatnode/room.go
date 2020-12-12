package chatnode

import (
	"lucky/core/iduck"
	"lucky/core/inet"
)

var testChatRoom iduck.INode

func GetRoom() iduck.INode {
	if testChatRoom == nil {
		testChatRoom = inet.NewBroadcastNode()
		testChatRoom.Serve()
	}
	return testChatRoom
}
