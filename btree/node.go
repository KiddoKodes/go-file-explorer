package btree

import indexer_struct "file-system-index/indexer-structs"

type Node struct {
    Keys     []indexer_struct.FileInfo
    Children []*Node
    Leaf     bool
}

func NewNode(leaf bool) *Node {
    return &Node{
        Keys:     make([]indexer_struct.FileInfo, 0),
        Children: make([]*Node, 0),
        Leaf:     leaf,
    }
}