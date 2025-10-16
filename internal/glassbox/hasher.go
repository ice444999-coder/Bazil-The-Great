package glassbox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// SpanHasher provides cryptographic hashing for decision spans
// Implements blockchain-style hash chaining for tamper detection
type SpanHasher struct{}

// SpanHashData represents data to be hashed
type SpanHashData struct {
	TraceID           int
	SpanID            int
	SpanName          string
	SpanType          string
	ChainPosition     int
	StartTime         string
	InputData         string
	OutputData        string
	DecisionReasoning string
	ConfidenceScore   float64
	PreviousHash      string
}

// HashSpan creates SHA-256 hash of span data with previous hash chaining
// Returns: (hashHex, dataSnapshot)
// - hashHex: The calculated SHA-256 hash
// - dataSnapshot: The canonical string used for hashing (for verification)
func (sh *SpanHasher) HashSpan(span *Span, previousHash string) (string, string) {
	// Serialize data in canonical format (deterministic)
	inputJSON, _ := json.Marshal(span.InputData)
	outputJSON, _ := json.Marshal(span.OutputData)

	reasoning := ""
	if span.DecisionReasoning != nil {
		reasoning = *span.DecisionReasoning
	}

	confidence := 0.0
	if span.ConfidenceScore != nil {
		confidence = *span.ConfidenceScore
	}

	// Create deterministic data string (order matters for hash consistency)
	dataSnapshot := fmt.Sprintf(
		"TRACE:%d|SPAN:%d|NAME:%s|TYPE:%s|POS:%d|TIME:%s|IN:%s|OUT:%s|REASON:%s|CONF:%.2f|PREV:%s",
		span.TraceID,
		span.ID,
		span.SpanName,
		span.SpanType,
		span.ChainPosition,
		span.StartTime.Format(time.RFC3339Nano),
		string(inputJSON),
		string(outputJSON),
		reasoning,
		confidence,
		previousHash,
	)

	// Calculate SHA-256
	hash := sha256.Sum256([]byte(dataSnapshot))
	hashHex := hex.EncodeToString(hash[:])

	return hashHex, dataSnapshot
}

// VerifyChain checks if entire span chain is unmodified
// Returns (isValid, error)
// Walks through spans and recalculates each hash to verify integrity
func (sh *SpanHasher) VerifyChain(spans []Span) (bool, error) {
	for i, span := range spans {
		var previousHash string
		if i > 0 {
			previousHash = spans[i-1].SHA256Hash
		} else {
			previousHash = ""
		}

		calculatedHash, _ := sh.HashSpan(&span, previousHash)

		if calculatedHash != span.SHA256Hash {
			return false, fmt.Errorf(
				"hash chain broken at span %d (%s): expected %s, got %s",
				span.ID, span.SpanName, span.SHA256Hash, calculatedHash,
			)
		}
	}
	return true, nil
}

// CalculateMerkleRoot creates merkle tree root from span hashes
// Uses bottom-up merkle tree construction
// Returns: merkle root hash (empty string if no hashes)
func (sh *SpanHasher) CalculateMerkleRoot(spanHashes []string) string {
	if len(spanHashes) == 0 {
		return ""
	}

	// Build merkle tree bottom-up
	currentLevel := spanHashes

	for len(currentLevel) > 1 {
		var nextLevel []string

		// Pair up hashes and hash them together
		for i := 0; i < len(currentLevel); i += 2 {
			var combined string

			if i+1 < len(currentLevel) {
				// Pair exists
				combined = currentLevel[i] + currentLevel[i+1]
			} else {
				// Odd one out, hash with itself
				combined = currentLevel[i] + currentLevel[i]
			}

			hash := sha256.Sum256([]byte(combined))
			nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
		}

		currentLevel = nextLevel
	}

	return currentLevel[0]
}

// VerifyMerkleProof verifies a span is part of the merkle tree
// Used for efficient verification without downloading entire tree
// Parameters:
// - spanHash: The hash of the span to verify
// - merkleRoot: The known merkle root
// - proof: Array of sibling hashes from leaf to root
// - index: Position of span in the original array
func (sh *SpanHasher) VerifyMerkleProof(spanHash, merkleRoot string, proof []string, index int) bool {
	currentHash := spanHash

	for _, proofHash := range proof {
		var combined string
		if index%2 == 0 {
			// Current hash is on the left
			combined = currentHash + proofHash
		} else {
			// Current hash is on the right
			combined = proofHash + currentHash
		}

		hash := sha256.Sum256([]byte(combined))
		currentHash = hex.EncodeToString(hash[:])
		index /= 2
	}

	return currentHash == merkleRoot
}

// GenerateMerkleProof generates proof path for a span at given index
// Returns array of sibling hashes needed to verify the span
func (sh *SpanHasher) GenerateMerkleProof(spanHashes []string, index int) []string {
	if len(spanHashes) == 0 || index >= len(spanHashes) {
		return nil
	}

	var proof []string
	currentLevel := spanHashes
	currentIndex := index

	for len(currentLevel) > 1 {
		var nextLevel []string
		var siblingIndex int

		// Determine sibling index
		if currentIndex%2 == 0 {
			siblingIndex = currentIndex + 1
		} else {
			siblingIndex = currentIndex - 1
		}

		// Add sibling to proof (if it exists)
		if siblingIndex < len(currentLevel) {
			proof = append(proof, currentLevel[siblingIndex])
		}

		// Build next level
		for i := 0; i < len(currentLevel); i += 2 {
			var combined string
			if i+1 < len(currentLevel) {
				combined = currentLevel[i] + currentLevel[i+1]
			} else {
				combined = currentLevel[i] + currentLevel[i]
			}

			hash := sha256.Sum256([]byte(combined))
			nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
		}

		currentLevel = nextLevel
		currentIndex /= 2
	}

	return proof
}
