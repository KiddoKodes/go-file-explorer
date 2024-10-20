package btree

import (
	indexer_struct "file-system-index/indexer-structs"
	"sync"
	"unsafe"
)
func (t *BTree) EstimateSize() int64 {
    return t.estimateNodeSize(t.Root)
}

func (t *BTree) estimateNodeSize(node *Node) int64 {
    if node == nil {
        return 0
    }

    size := int64(unsafe.Sizeof(*node))

    // Estimate size of keys
    size += int64(cap(node.Keys)) * int64(unsafe.Sizeof(indexer_struct.FileInfo{}))

    // Estimate size of children pointers
    size += int64(cap(node.Children)) * int64(unsafe.Sizeof(&Node{}))

    // Recursively estimate size of children
    for _, child := range node.Children {
        size += t.estimateNodeSize(child)
    }

    return size
}
const MinDegree = 2

type BTree struct {
    Root  *Node
    mutex sync.RWMutex
}

func NewBTree() *BTree {
    return &BTree{
        Root: NewNode(true),
    }
}

func (t *BTree) Insert(info indexer_struct.FileInfo) {
    t.mutex.Lock()
    defer t.mutex.Unlock()

    root := t.Root
    if len(root.Keys) == 2*MinDegree-1 {
        newRoot := NewNode(false)
        newRoot.Children = append(newRoot.Children, root)
        t.splitChild(newRoot, 0)
        t.Root = newRoot
    }
    t.insertNonFull(t.Root, info)
}

func (t *BTree) insertNonFull(n *Node, info indexer_struct.FileInfo) {
    i := len(n.Keys) - 1
    if n.Leaf {
        n.Keys = append(n.Keys, indexer_struct.FileInfo{})
        for i >= 0 && info.Hash.PathHash < n.Keys[i].Hash.PathHash {
            n.Keys[i+1] = n.Keys[i]
            i--
        }
        n.Keys[i+1] = info
    } else {
        for i >= 0 && info.Hash.PathHash < n.Keys[i].Hash.PathHash {
            i--
        }
        i++
        if len(n.Children[i].Keys) == 2*MinDegree-1 {
            t.splitChild(n, i)
            if info.Hash.PathHash > n.Keys[i].Hash.PathHash {
                i++
            }
        }
        t.insertNonFull(n.Children[i], info)
    }
}

func (t *BTree) splitChild(parent *Node, i int) {
    child := parent.Children[i]
    newChild := NewNode(child.Leaf)
    parent.Children = append(parent.Children, nil)
    copy(parent.Children[i+2:], parent.Children[i+1:])
    parent.Children[i+1] = newChild
    parent.Keys = append(parent.Keys, indexer_struct.FileInfo{})
    copy(parent.Keys[i+1:], parent.Keys[i:])
    parent.Keys[i] = child.Keys[MinDegree-1]
    newChild.Keys = append(newChild.Keys, child.Keys[MinDegree:]...)
    child.Keys = child.Keys[:MinDegree-1]
    if !child.Leaf {
        newChild.Children = append(newChild.Children, child.Children[MinDegree:]...)
        child.Children = child.Children[:MinDegree]
    }
}

func (t *BTree) Search(hash string) *indexer_struct.FileInfo {
    t.mutex.RLock()
    defer t.mutex.RUnlock()
    return t.searchNode(t.Root, hash)
}

func (t *BTree) searchNode(n *Node, hash string) *indexer_struct.FileInfo {
    i := 0
    for i < len(n.Keys) && hash > n.Keys[i].Hash.PathHash {
        i++
    }
    if i < len(n.Keys) && hash == n.Keys[i].Hash.PathHash {
        return &n.Keys[i]
    }
    if n.Leaf {
        return nil
    }
    return t.searchNode(n.Children[i], hash)
}

func (t *BTree) InOrderTraversal(fn func(indexer_struct.FileInfo) bool) {
    t.mutex.RLock()
    defer t.mutex.RUnlock()
    t.inOrderTraversalNode(t.Root, fn)
}

func (t *BTree) inOrderTraversalNode(n *Node, fn func(indexer_struct.FileInfo) bool) bool {
    if n == nil {
        return true
    }
    for i, key := range n.Keys {
        if !n.Leaf {
            if !t.inOrderTraversalNode(n.Children[i], fn) {
                return false
            }
        }
        if !fn(key) {
            return false
        }
    }
    if !n.Leaf {
        return t.inOrderTraversalNode(n.Children[len(n.Keys)], fn)
    }
    return true
}