package hedera

import (
	"fmt"
	"log"
	"os"
	"time"
)

// SubmitRootHash is a lightweight submission function that currently supports
// a MOCK mode (no external dependencies). To enable real Hedera integration,
// install the Hedera Go SDK and replace this implementation or add a build
// tag implementation that imports the SDK.
// Environment variables used (optional for mock):
//
//	HEDERA_OPERATOR_ID
//	HEDERA_OPERATOR_KEY
//	HEDERA_TOPIC_ID
//	HEDERA_NETWORK (testnet|mainnet)
func SubmitRootHash(rootHash string) (txID string, sequence int64, consensusTime time.Time, topicID string, err error) {
	topicID = os.Getenv("HEDERA_TOPIC_ID")
	if topicID == "" {
		topicID = MockHederaTopicID()
	}

	// If operator credentials are missing, return a MOCK response but no error.
	operatorID := os.Getenv("HEDERA_OPERATOR_ID")
	operatorKey := os.Getenv("HEDERA_OPERATOR_KEY")
	if operatorID == "" || operatorKey == "" {
		// Return a deterministic mock tx id and sequence number
		now := time.Now()
		txid := fmt.Sprintf("mock-%d", now.UnixNano())
		seq := now.Unix()
		log.Printf("[hedera] MOCK submit: root=%s topic=%s tx=%s seq=%d", rootHash[:16]+"...", topicID, txid, seq)
		return txid, seq, now, topicID, nil
	}

	// If credentials are present but the real SDK isn't linked, still mock and warn.
	log.Printf("[hedera] Hedera credentials provided but SDK integration not enabled in this build. Falling back to MOCK submission.")
	now := time.Now()
	txid := fmt.Sprintf("mock-enabled-%d", now.UnixNano())
	seq := now.Unix()
	return txid, seq, now, topicID, nil
}
