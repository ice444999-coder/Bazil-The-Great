package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Hash  string
	Data  string // Only for leaf nodes
}

// MerkleTree represents the complete tree
type MerkleTree struct {
	Root       *MerkleNode
	Leaves     []*MerkleNode
	LeafHashes []string
	CreatedAt  time.Time
}

// LogEntry represents a single entry to be included in the tree
type LogEntry struct {
	LogID        uint
	Actor        string
	ActionType   string
	MessageHash  string // SHA-256 of the actual message
	Timestamp    time.Time
	InternalHash string // Double-hashed for extra security
}

// NewMerkleTree creates a new Merkle tree from log entries
func NewMerkleTree(entries []LogEntry) (*MerkleTree, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("cannot create merkle tree with zero entries")
	}

	var leaves []*MerkleNode
	var leafHashes []string

	// Create leaf nodes from entries
	for _, entry := range entries {
		// Hash the entry with all its data (double-hashed)
		leafData := fmt.Sprintf("%d|%s|%s|%s|%d",
			entry.LogID,
			entry.Actor,
			entry.ActionType,
			entry.MessageHash,
			entry.Timestamp.Unix(),
		)

		// Double hash for security
		firstHash := sha256.Sum256([]byte(leafData))
		secondHash := sha256.Sum256(firstHash[:])
		leafHash := hex.EncodeToString(secondHash[:])

		node := &MerkleNode{
			Hash: leafHash,
			Data: leafData,
		}
		leaves = append(leaves, node)
		leafHashes = append(leafHashes, leafHash)
	}

	// Build the tree from leaves
	root := buildTree(leaves)

	return &MerkleTree{
		Root:       root,
		Leaves:     leaves,
		LeafHashes: leafHashes,
		CreatedAt:  time.Now(),
	}, nil
}

// buildTree recursively builds the Merkle tree
func buildTree(nodes []*MerkleNode) *MerkleNode {
	if len(nodes) == 1 {
		return nodes[0]
	}

	var parentNodes []*MerkleNode

	// Process pairs of nodes
	for i := 0; i < len(nodes); i += 2 {
		var left, right *MerkleNode
		left = nodes[i]

		// If odd number of nodes, duplicate the last one
		if i+1 < len(nodes) {
			right = nodes[i+1]
		} else {
			right = nodes[i]
		}

		// Create parent node
		parentHash := combineHashes(left.Hash, right.Hash)
		parent := &MerkleNode{
			Left:  left,
			Right: right,
			Hash:  parentHash,
		}
		parentNodes = append(parentNodes, parent)
	}

	return buildTree(parentNodes)
}

// combineHashes combines two hashes and creates a new hash
func combineHashes(left, right string) string {
	combined := left + right
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// GetRootHash returns the root hash of the tree
func (mt *MerkleTree) GetRootHash() string {
	if mt.Root == nil {
		return ""
	}
	return mt.Root.Hash
}

// GenerateProof generates a Merkle proof for a specific leaf index
func (mt *MerkleTree) GenerateProof(leafIndex int) ([]string, error) {
	if leafIndex < 0 || leafIndex >= len(mt.Leaves) {
		return nil, fmt.Errorf("invalid leaf index: %d", leafIndex)
	}

	var proof []string
	currentNodes := mt.Leaves
	currentIndex := leafIndex

	// Traverse up the tree
	for len(currentNodes) > 1 {
		var parentNodes []*MerkleNode

		for i := 0; i < len(currentNodes); i += 2 {
			left := currentNodes[i]
			var right *MerkleNode

			if i+1 < len(currentNodes) {
				right = currentNodes[i+1]
			} else {
				right = currentNodes[i]
			}

			// If current index is in this pair, add sibling to proof
			if i == currentIndex || i+1 == currentIndex {
				if currentIndex%2 == 0 {
					// Current is left, add right to proof
					proof = append(proof, right.Hash)
				} else {
					// Current is right, add left to proof
					proof = append(proof, left.Hash)
				}
			}

			parent := &MerkleNode{
				Hash: combineHashes(left.Hash, right.Hash),
			}
			parentNodes = append(parentNodes, parent)
		}

		currentIndex = currentIndex / 2
		currentNodes = parentNodes
	}

	return proof, nil
}

// VerifyProof verifies a Merkle proof
func VerifyProof(leafHash string, proof []string, rootHash string, leafIndex int) bool {
	currentHash := leafHash
	currentIndex := leafIndex

	for _, siblingHash := range proof {
		if currentIndex%2 == 0 {
			// Current is left child
			currentHash = combineHashes(currentHash, siblingHash)
		} else {
			// Current is right child
			currentHash = combineHashes(siblingHash, currentHash)
		}
		currentIndex = currentIndex / 2
	}

	return currentHash == rootHash
}

// GetTreeDepth returns the depth of the tree
func (mt *MerkleTree) GetTreeDepth() int {
	if mt.Root == nil {
		return 0
	}
	return getNodeDepth(mt.Root)
}

func getNodeDepth(node *MerkleNode) int {
	if node == nil || (node.Left == nil && node.Right == nil) {
		return 0
	}

	leftDepth := getNodeDepth(node.Left)
	rightDepth := getNodeDepth(node.Right)

	if leftDepth > rightDepth {
		return leftDepth + 1
	}
	return rightDepth + 1
}

// GetLeafCount returns the number of leaves in the tree
func (mt *MerkleTree) GetLeafCount() int {
	return len(mt.Leaves)
}
