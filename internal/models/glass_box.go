package models

import "time"

// GlassBoxLog represents a single logged action with maximum security
type GlassBoxLog struct {
	LogID          uint   `gorm:"primaryKey;autoIncrement" json:"log_id"`
	Actor          string `gorm:"size:50;not null" json:"actor"`
	ActionType     string `gorm:"size:100;not null" json:"action_type"`
	FeatureTested  string `gorm:"size:100" json:"feature_tested"`
	MessageContent string `gorm:"type:text" json:"-"`  // Never exposed in API
	ActionDetails  string `gorm:"type:jsonb" json:"-"` // Never exposed in API

	// Internal hashing (stored in DB, never sent to Hedera)
	InternalHash string `gorm:"size:64;not null;index" json:"-"` // SHA-256 of content

	// Merkle tree association
	MerkleBatchID   *uint  `gorm:"index" json:"merkle_batch_id,omitempty"`
	MerkleLeafIndex *int   `json:"merkle_leaf_index,omitempty"`
	MerkleProof     string `gorm:"type:jsonb" json:"-"` // JSON array of proof hashes

	// Results
	Result       string `gorm:"size:20" json:"result"`
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`

	// Timestamps
	Timestamp time.Time `gorm:"not null;index" json:"timestamp"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (GlassBoxLog) TableName() string {
	return "glass_box_log"
}

// MerkleBatch represents a batch of logs submitted to Hedera
type MerkleBatch struct {
	BatchID   uint   `gorm:"primaryKey;autoIncrement" json:"batch_id"`
	RootHash  string `gorm:"size:64;not null;unique" json:"root_hash"` // Only this goes to Hedera
	LeafCount int    `gorm:"not null" json:"leaf_count"`
	TreeDepth int    `json:"tree_depth"`

	// Hedera verification (proof that root was submitted)
	HederaTopicID   string     `gorm:"size:20" json:"hedera_topic_id,omitempty"`
	HederaTxID      string     `gorm:"size:100;index" json:"hedera_tx_id,omitempty"`
	HederaSequence  *int64     `json:"hedera_sequence,omitempty"`
	HederaTimestamp *time.Time `json:"hedera_timestamp,omitempty"`

	// Batch metadata
	BatchStartTime time.Time  `gorm:"not null" json:"batch_start_time"`
	BatchEndTime   time.Time  `gorm:"not null" json:"batch_end_time"`
	SubmittedAt    *time.Time `json:"submitted_at,omitempty"`
	Verified       bool       `gorm:"default:false" json:"verified"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (MerkleBatch) TableName() string {
	return "merkle_batches"
}

// VerificationRequest represents a request to verify a log entry
type VerificationRequest struct {
	LogID uint `json:"log_id" binding:"required"`
}

// VerificationResponse represents the response with full verification chain
type VerificationResponse struct {
	LogID      uint      `json:"log_id"`
	Actor      string    `json:"actor"`
	ActionType string    `json:"action_type"`
	Timestamp  time.Time `json:"timestamp"`
	Result     string    `json:"result"`

	// Merkle verification (admin only)
	MerkleVerification *MerkleVerificationData `json:"merkle_verification,omitempty"`

	// Hedera proof (public)
	HederaProof *HederaProofData `json:"hedera_proof,omitempty"`

	Verified        bool   `json:"verified"`
	VerificationMsg string `json:"verification_message"`
}

// MerkleVerificationData contains the Merkle proof chain (admin only)
type MerkleVerificationData struct {
	InternalHash string   `json:"internal_hash"`
	MerkleProof  []string `json:"merkle_proof"`
	RootHash     string   `json:"root_hash"`
	LeafIndex    int      `json:"leaf_index"`
	BatchID      uint     `json:"batch_id"`
	ProofValid   bool     `json:"proof_valid"`
}

// HederaProofData contains the public Hedera verification (safe to expose)
type HederaProofData struct {
	TopicID        string    `json:"topic_id"`
	SequenceNumber int64     `json:"sequence_number"`
	ConsensusTime  time.Time `json:"consensus_timestamp"`
	ExplorerURL    string    `json:"explorer_url"`
	RootHash       string    `json:"root_hash"` // Only the root, not individual hashes
}
