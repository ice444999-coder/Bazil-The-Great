package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ares_api/internal/hedera"
	"ares_api/internal/merkle"
	"ares_api/internal/models"

	"gorm.io/gorm"
)

// MerkleBatchService handles batching and Merkle tree operations
type MerkleBatchService struct {
	db            *gorm.DB
	batchSize     int
	batchInterval time.Duration
	pendingLogs   []models.GlassBoxLog
	lastBatchTime time.Time
}

// NewMerkleBatchService creates a new Merkle batch service
func NewMerkleBatchService(db *gorm.DB) *MerkleBatchService {
	return &MerkleBatchService{
		db:            db,
		batchSize:     100,             // Submit to Hedera every 100 logs
		batchInterval: 5 * time.Minute, // Or every 5 minutes, whichever comes first
		pendingLogs:   []models.GlassBoxLog{},
		lastBatchTime: time.Now(),
	}
}

// AddToBatch adds a log entry to the pending batch
func (s *MerkleBatchService) AddToBatch(log models.GlassBoxLog) error {
	s.pendingLogs = append(s.pendingLogs, log)

	// Check if we should process the batch
	shouldProcess := len(s.pendingLogs) >= s.batchSize ||
		time.Since(s.lastBatchTime) >= s.batchInterval

	if shouldProcess {
		return s.ProcessBatch()
	}

	return nil
}

// ProcessBatch creates a Merkle tree and prepares for Hedera submission
func (s *MerkleBatchService) ProcessBatch() error {
	if len(s.pendingLogs) == 0 {
		return nil
	}

	log.Printf("🌲 Processing Merkle batch with %d entries", len(s.pendingLogs))

	// Convert to Merkle log entries
	var entries []merkle.LogEntry
	for _, glLog := range s.pendingLogs {
		entries = append(entries, merkle.LogEntry{
			LogID:        glLog.LogID,
			Actor:        glLog.Actor,
			ActionType:   glLog.ActionType,
			MessageHash:  glLog.InternalHash,
			Timestamp:    glLog.Timestamp,
			InternalHash: glLog.InternalHash,
		})
	}

	// Build Merkle tree
	tree, err := merkle.NewMerkleTree(entries)
	if err != nil {
		return fmt.Errorf("failed to build merkle tree: %w", err)
	}

	rootHash := tree.GetRootHash()
	log.Printf("🌲 Merkle root hash: %s", rootHash[:16]+"...")

	// Create batch record
	batch := models.MerkleBatch{
		RootHash:       rootHash,
		LeafCount:      tree.GetLeafCount(),
		TreeDepth:      tree.GetTreeDepth(),
		BatchStartTime: s.pendingLogs[0].Timestamp,
		BatchEndTime:   s.pendingLogs[len(s.pendingLogs)-1].Timestamp,
	}

	// Save batch to database
	if err := s.db.Create(&batch).Error; err != nil {
		return fmt.Errorf("failed to save merkle batch: %w", err)
	}

	log.Printf("🌲 Merkle batch created: ID=%d, Leaves=%d, Depth=%d",
		batch.BatchID, batch.LeafCount, batch.TreeDepth)

	// Update each log with Merkle proof
	for i, glLog := range s.pendingLogs {
		proof, err := tree.GenerateProof(i)
		if err != nil {
			log.Printf("⚠️ Failed to generate proof for log %d: %v", glLog.LogID, err)
			continue
		}

		// Store proof as JSON
		proofJSON, _ := json.Marshal(proof)

		// Update log entry
		updates := map[string]interface{}{
			"merkle_batch_id":   batch.BatchID,
			"merkle_leaf_index": i,
			"merkle_proof":      string(proofJSON),
		}

		if err := s.db.Model(&models.GlassBoxLog{}).
			Where("log_id = ?", glLog.LogID).
			Updates(updates).Error; err != nil {
			log.Printf("⚠️ Failed to update log %d with Merkle proof: %v", glLog.LogID, err)
		}
	}

	log.Printf("✅ All %d logs updated with Merkle proofs", len(s.pendingLogs))

	// Clear pending logs
	s.pendingLogs = []models.GlassBoxLog{}
	s.lastBatchTime = time.Now()

	// Submit root hash to Hedera (or mock)
	txid, seq, consTime, topicID, err := hedera.SubmitRootHash(rootHash)
	if err != nil {
		log.Printf("⚠️ Failed to submit root to Hedera: %v", err)
	} else {
		// Update batch with Hedera proof
		updates := map[string]interface{}{
			"hedera_topic_id":  topicID,
			"hedera_tx_id":     txid,
			"hedera_sequence":  seq,
			"hedera_timestamp": consTime,
			"submitted_at":     time.Now(),
			"verified":         true,
		}

		if err := s.db.Model(&models.MerkleBatch{}).
			Where("batch_id = ?", batch.BatchID).
			Updates(updates).Error; err != nil {
			log.Printf("⚠️ Failed to update batch %d with Hedera proof: %v", batch.BatchID, err)
		} else {
			log.Printf("✅ Merkle batch %d submitted to Hedera tx=%s seq=%d", batch.BatchID, txid, seq)
		}
	}

	return nil
}

// VerifyLog verifies a log entry against its Merkle proof and batch
func (s *MerkleBatchService) VerifyLog(logID uint) (*models.VerificationResponse, error) {
	var log models.GlassBoxLog
	if err := s.db.Where("log_id = ?", logID).First(&log).Error; err != nil {
		return nil, fmt.Errorf("log not found: %w", err)
	}

	response := &models.VerificationResponse{
		LogID:      log.LogID,
		Actor:      log.Actor,
		ActionType: log.ActionType,
		Timestamp:  log.Timestamp,
		Result:     log.Result,
	}

	// Check if log is part of a batch
	if log.MerkleBatchID == nil {
		response.Verified = false
		response.VerificationMsg = "Log not yet included in Merkle batch"
		return response, nil
	}

	// Get batch information
	var batch models.MerkleBatch
	if err := s.db.Where("batch_id = ?", *log.MerkleBatchID).First(&batch).Error; err != nil {
		return nil, fmt.Errorf("batch not found: %w", err)
	}

	// Parse Merkle proof
	var proof []string
	if err := json.Unmarshal([]byte(log.MerkleProof), &proof); err != nil {
		return nil, fmt.Errorf("failed to parse merkle proof: %w", err)
	}

	// Verify Merkle proof
	proofValid := merkle.VerifyProof(
		log.InternalHash,
		proof,
		batch.RootHash,
		*log.MerkleLeafIndex,
	)

	response.MerkleVerification = &models.MerkleVerificationData{
		InternalHash: log.InternalHash,
		MerkleProof:  proof,
		RootHash:     batch.RootHash,
		LeafIndex:    *log.MerkleLeafIndex,
		BatchID:      batch.BatchID,
		ProofValid:   proofValid,
	}

	// Add Hedera proof if available
	if batch.HederaTxID != "" && batch.HederaSequence != nil {
		explorerURL := fmt.Sprintf("https://hashscan.io/testnet/topic/%s/message/%d",
			batch.HederaTopicID, *batch.HederaSequence)

		response.HederaProof = &models.HederaProofData{
			TopicID:        batch.HederaTopicID,
			SequenceNumber: *batch.HederaSequence,
			ConsensusTime:  *batch.HederaTimestamp,
			ExplorerURL:    explorerURL,
			RootHash:       batch.RootHash,
		}
	}

	response.Verified = proofValid && batch.Verified
	if response.Verified {
		response.VerificationMsg = "Log verified: Merkle proof valid and root hash confirmed on Hedera"
	} else if proofValid {
		response.VerificationMsg = "Merkle proof valid, awaiting Hedera confirmation"
	} else {
		response.VerificationMsg = "Verification failed: Invalid Merkle proof"
	}

	return response, nil
}

// GetPendingBatchInfo returns information about the current pending batch
func (s *MerkleBatchService) GetPendingBatchInfo() map[string]interface{} {
	return map[string]interface{}{
		"pending_logs":    len(s.pendingLogs),
		"batch_size":      s.batchSize,
		"next_batch_in":   s.batchInterval - time.Since(s.lastBatchTime),
		"last_batch_time": s.lastBatchTime,
	}
}
