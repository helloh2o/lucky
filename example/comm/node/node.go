package node

import (
	"github.com/helloh2o/lucky/core/inet"
)

// TestNode for testing
var TestNode *inet.FrameNode

// NewTestNode init & run the test node
func NewTestNode() {
	TestNode = inet.NewFrameNode()
	TestNode.Serve()
}
