package node

import (
	"lucky/core/inet"
)

var TestNode *inet.FrameNode

func NewTestNode() {
	TestNode = inet.NewFrameNode()
	TestNode.Serve()
}
