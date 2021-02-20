package node

import (
	"github.com/helloh2o/lucky"
)

// TestNode for testing
var TestNode *lucky.FrameNode

// NewTestNode init & run the test node
func NewTestNode() {
	TestNode = lucky.NewFrameNode()
	TestNode.Serve()
}
