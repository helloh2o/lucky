package node

import (
	"github.com/helloh2o/lucky/core/inet"
)

var TestNode *inet.FrameNode

func NewTestNode() {
	TestNode = inet.NewFrameNode()
	TestNode.Serve()
}
