package glassbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// TEMPORARILY DISABLED: Hedera SDK has dependency issues
	// Will be re-enabled in Phase 4
	// hedera "github.com/hiero-ledger/hiero-sdk-go/v2"
	"github.com/lib/pq"
)

// PLACEHOLDER: Hedera SDK types until dependency is resolved
type hederaClient struct{}
type hederaTopicID struct{}

func (t hederaTopicID) String() string { return "" }

// HederaAnchorer manages blockchain anchoring of decision trees
// TEMPORARILY DISABLED: Will be fully implemented in Phase 4
type HederaAnchorer struct {
	db           *sql.DB
	hederaClient *hederaClient // Placeholder
	topicID      hederaTopicID // Placeholder
}

// NewHederaAnchorer creates a new Hedera anchorer
// TEMPORARILY DISABLED: Returns stub until SDK dependency resolved
func NewHederaAnchorer(db *sql.DB, hederaClient *hederaClient, topicID hederaTopicID) *HederaAnchorer {
	return &HederaAnchorer{
		db:           db,
		hederaClient: hederaClient,
		topicID:      topicID,
	}
}

// AnchorTrace submits trace merkle root to Hedera
// TEMPORARILY DISABLED: Returns not implemented error
func (ha *HederaAnchorer) AnchorTrace(ctx context.Context, traceID int) error {
	return fmt.Errorf("Hedera anchoring temporarily disabled - will be enabled in Phase 4")
}

// VerifyTraceFromHedera verifies trace against Hedera anchor
// TEMPORARILY DISABLED: Skips blockchain verification, only checks database
func (ha *HederaAnchorer) VerifyTraceFromHedera(ctx context.Context, traceID int) (bool, error) {
	// 1. Get anchor from database (if exists)
	var anchor struct {
		MerkleRoot string
		LeafHashes []string
	}

	err := ha.db.QueryRowContext(ctx,
		`SELECT merkle_root, leaf_hashes
         FROM hedera_anchors WHERE trace_id = $1`,
		traceID,
	).Scan(&anchor.MerkleRoot, pq.Array(&anchor.LeafHashes))

	if err != nil {
		// No anchor found - that's okay, blockchain is optional
		return false, nil
	}

	// 2. Recalculate merkle root from current database
	rows, err := ha.db.QueryContext(ctx,
		`SELECT sha256_hash FROM decision_spans 
         WHERE trace_id = $1 ORDER BY chain_position`,
		traceID,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var currentHashes []string
	for rows.Next() {
		var hash string
		rows.Scan(&hash)
		currentHashes = append(currentHashes, hash)
	}

	hasher := &SpanHasher{}
	currentRoot := hasher.CalculateMerkleRoot(currentHashes)

	// 3. Compare
	isValid := currentRoot == anchor.MerkleRoot && len(currentHashes) == len(anchor.LeafHashes)

	// 4. Log verification
	ha.db.ExecContext(ctx,
		`INSERT INTO hash_chain_verifications 
         (trace_id, verification_type, is_valid, error_message, verified_by)
         VALUES ($1, 'database_match', $2, $3, 'system')`,
		traceID,
		isValid,
		func() *string {
			if !isValid {
				s := fmt.Sprintf("Merkle root mismatch: expected %s, got %s", anchor.MerkleRoot, currentRoot)
				return &s
			}
			return nil
		}(),
	)

	return isValid, nil
}

// GetAnchorInfo retrieves Hedera anchor details for display
func (ha *HederaAnchorer) GetAnchorInfo(ctx context.Context, traceID int) (*HederaAnchor, error) {
	var anchor HederaAnchor

	err := ha.db.QueryRowContext(ctx,
		`SELECT trace_id, merkle_root, span_count, hedera_topic_id, hedera_txn_id,
                hedera_consensus_timestamp, hedera_sequence_number, verification_url,
                verification_status, verified_at, anchored_at
         FROM hedera_anchors WHERE trace_id = $1`,
		traceID,
	).Scan(
		&anchor.TraceID, &anchor.MerkleRoot, &anchor.SpanCount,
		&anchor.HederaTopicID, &anchor.HederaTxnID, &anchor.HederaConsensusTimestamp,
		&anchor.HederaSequenceNumber, &anchor.VerificationURL,
		&anchor.VerificationStatus, &anchor.VerifiedAt, &anchor.AnchoredAt,
	)

	return &anchor, err
}

// HederaAnchor represents blockchain anchor information
type HederaAnchor struct {
	TraceID                  int
	MerkleRoot               string
	SpanCount                int
	HederaTopicID            string
	HederaTxnID              string
	HederaConsensusTimestamp time.Time
	HederaSequenceNumber     uint64
	VerificationURL          string
	VerificationStatus       string
	VerifiedAt               *time.Time
	AnchoredAt               time.Time
}

// ListRecentAnchors retrieves recent Hedera anchors for monitoring
func (ha *HederaAnchorer) ListRecentAnchors(ctx context.Context, limit int) ([]HederaAnchor, error) {
	rows, err := ha.db.QueryContext(ctx,
		`SELECT trace_id, merkle_root, span_count, hedera_topic_id, hedera_txn_id,
                hedera_consensus_timestamp, hedera_sequence_number, verification_url,
                verification_status, verified_at, anchored_at
         FROM hedera_anchors
         ORDER BY anchored_at DESC
         LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anchors []HederaAnchor
	for rows.Next() {
		var anchor HederaAnchor
		err := rows.Scan(
			&anchor.TraceID, &anchor.MerkleRoot, &anchor.SpanCount,
			&anchor.HederaTopicID, &anchor.HederaTxnID, &anchor.HederaConsensusTimestamp,
			&anchor.HederaSequenceNumber, &anchor.VerificationURL,
			&anchor.VerificationStatus, &anchor.VerifiedAt, &anchor.AnchoredAt,
		)
		if err != nil {
			return nil, err
		}
		anchors = append(anchors, anchor)
	}

	return anchors, nil
}
