/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package glassbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// DecisionTracer manages decision traces and spans with cryptographic verification
type DecisionTracer struct {
	db     *sql.DB
	hasher *SpanHasher
}

// NewDecisionTracer creates a new decision tracer
func NewDecisionTracer(db *sql.DB) *DecisionTracer {
	return &DecisionTracer{
		db:     db,
		hasher: &SpanHasher{},
	}
}

// Trace represents a complete decision flow
type Trace struct {
	ID              int
	TradeID         *int
	TraceType       string
	Status          string
	StartTime       time.Time
	EndTime         *time.Time
	TotalDurationMs *int
	FinalDecision   *string
	ConfidenceScore *float64
	IsAnchored      bool
	MerkleRoot      *string
}

// Span represents an individual node in the decision tree
type Span struct {
	ID                  int
	TraceID             int
	ParentSpanID        *int
	SpanName            string
	SpanType            string
	ChainPosition       int
	StartTime           time.Time
	EndTime             *time.Time
	DurationMs          *int
	InputData           map[string]interface{}
	OutputData          map[string]interface{}
	DecisionReasoning   *string
	ConfidenceScore     *float64
	Status              string
	ErrorMessage        *string
	ContextFromPrevious map[string]interface{}
	ContextToNext       map[string]interface{}
	SHA256Hash          string
	PreviousHash        string
	DataSnapshot        string
}

// StartTrace creates a new decision trace
func (dt *DecisionTracer) StartTrace(ctx context.Context, traceType string, tradeID *int) (*Trace, error) {
	var trace Trace
	err := dt.db.QueryRowContext(ctx,
		`INSERT INTO decision_traces (trace_type, trade_id, status, start_time)
         VALUES ($1, $2, 'in_progress', NOW())
         RETURNING id, trace_type, trade_id, status, start_time, is_anchored`,
		traceType, tradeID,
	).Scan(&trace.ID, &trace.TraceType, &trace.TradeID, &trace.Status, &trace.StartTime, &trace.IsAnchored)

	return &trace, err
}

// StartSpan creates a new span with automatic hash chaining
func (dt *DecisionTracer) StartSpan(ctx context.Context, traceID int, parentSpanID *int, spanName, spanType string, inputData map[string]interface{}) (*Span, error) {
	inputJSON, _ := json.Marshal(inputData)

	// Get chain position and previous hash
	var chainPosition int
	var previousHash string

	err := dt.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(chain_position), -1) + 1, 
                COALESCE((SELECT sha256_hash FROM decision_spans 
                         WHERE trace_id = $1 
                         ORDER BY chain_position DESC LIMIT 1), '')
         FROM decision_spans WHERE trace_id = $1`,
		traceID,
	).Scan(&chainPosition, &previousHash)

	if err != nil {
		return nil, err
	}

	// Create temporary span for initial hashing
	tempSpan := &Span{
		TraceID:       traceID,
		SpanName:      spanName,
		SpanType:      spanType,
		ChainPosition: chainPosition,
		StartTime:     time.Now(),
		InputData:     inputData,
		OutputData:    make(map[string]interface{}),
	}

	// Calculate initial hash (will be updated when span completes)
	initialHash, dataSnapshot := dt.hasher.HashSpan(tempSpan, previousHash)

	var span Span
	err = dt.db.QueryRowContext(ctx,
		`INSERT INTO decision_spans 
         (trace_id, parent_span_id, span_name, span_type, chain_position, input_data, 
          status, start_time, previous_hash, sha256_hash, data_snapshot)
         VALUES ($1, $2, $3, $4, $5, $6, 'running', NOW(), $7, $8, $9)
         RETURNING id, trace_id, parent_span_id, span_name, span_type, chain_position,
                   status, start_time, previous_hash, sha256_hash`,
		traceID, parentSpanID, spanName, spanType, chainPosition, inputJSON,
		previousHash, initialHash, dataSnapshot,
	).Scan(
		&span.ID, &span.TraceID, &span.ParentSpanID, &span.SpanName, &span.SpanType,
		&span.ChainPosition, &span.Status, &span.StartTime, &span.PreviousHash, &span.SHA256Hash,
	)

	span.InputData = inputData
	span.OutputData = make(map[string]interface{})

	return &span, err
}

// EndSpan completes a span and recalculates final hash
func (dt *DecisionTracer) EndSpan(ctx context.Context, spanID int, outputData map[string]interface{}, reasoning string, confidence float64, status string) error {
	outputJSON, _ := json.Marshal(outputData)

	// Get span data for rehashing
	var span Span
	var inputJSON []byte

	err := dt.db.QueryRowContext(ctx,
		`SELECT trace_id, span_name, span_type, chain_position, start_time, 
                input_data, previous_hash
         FROM decision_spans WHERE id = $1`,
		spanID,
	).Scan(
		&span.TraceID, &span.SpanName, &span.SpanType, &span.ChainPosition,
		&span.StartTime, &inputJSON, &span.PreviousHash,
	)

	if err != nil {
		return err
	}

	// Unmarshal input data
	json.Unmarshal(inputJSON, &span.InputData)
	span.OutputData = outputData
	span.DecisionReasoning = &reasoning
	span.ConfidenceScore = &confidence
	span.ID = spanID

	// Recalculate final hash with all data
	finalHash, dataSnapshot := dt.hasher.HashSpan(&span, span.PreviousHash)

	// Update span with final data and hash
	_, err = dt.db.ExecContext(ctx,
		`UPDATE decision_spans
         SET end_time = NOW(),
             duration_ms = EXTRACT(EPOCH FROM (NOW() - start_time)) * 1000,
             output_data = $1,
             decision_reasoning = $2,
             confidence_score = $3,
             status = $4,
             sha256_hash = $5,
             data_snapshot = $6
         WHERE id = $7`,
		outputJSON, reasoning, confidence, status, finalHash, dataSnapshot, spanID,
	)

	return err
}

// EndTrace completes a trace and calculates merkle root
func (dt *DecisionTracer) EndTrace(ctx context.Context, traceID int, finalDecision string, confidence float64) error {
	// Get all span hashes
	rows, err := dt.db.QueryContext(ctx,
		`SELECT sha256_hash FROM decision_spans 
         WHERE trace_id = $1 ORDER BY chain_position`,
		traceID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var spanHashes []string
	for rows.Next() {
		var hash string
		rows.Scan(&hash)
		spanHashes = append(spanHashes, hash)
	}

	// Calculate merkle root
	merkleRoot := dt.hasher.CalculateMerkleRoot(spanHashes)

	// Update trace
	_, err = dt.db.ExecContext(ctx,
		`UPDATE decision_traces
         SET end_time = NOW(),
             total_duration_ms = EXTRACT(EPOCH FROM (NOW() - start_time)) * 1000,
             final_decision = $1,
             confidence_score = $2,
             status = 'completed',
             merkle_root = $3
         WHERE id = $4`,
		finalDecision, confidence, merkleRoot, traceID,
	)

	return err
}

// VerifyTrace checks hash chain integrity
func (dt *DecisionTracer) VerifyTrace(ctx context.Context, traceID int) (bool, error) {
	// Get all spans in order
	rows, err := dt.db.QueryContext(ctx,
		`SELECT id, trace_id, span_name, span_type, chain_position, start_time,
                input_data, output_data, decision_reasoning, confidence_score,
                sha256_hash, previous_hash, data_snapshot
         FROM decision_spans
         WHERE trace_id = $1
         ORDER BY chain_position`,
		traceID,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var spans []Span
	for rows.Next() {
		var span Span
		var inputJSON, outputJSON []byte
		var reasoning sql.NullString
		var confidence sql.NullFloat64

		err := rows.Scan(
			&span.ID, &span.TraceID, &span.SpanName, &span.SpanType,
			&span.ChainPosition, &span.StartTime, &inputJSON, &outputJSON,
			&reasoning, &confidence, &span.SHA256Hash, &span.PreviousHash,
			&span.DataSnapshot,
		)
		if err != nil {
			return false, err
		}

		json.Unmarshal(inputJSON, &span.InputData)
		json.Unmarshal(outputJSON, &span.OutputData)

		if reasoning.Valid {
			span.DecisionReasoning = &reasoning.String
		}
		if confidence.Valid {
			span.ConfidenceScore = &confidence.Float64
		}

		spans = append(spans, span)
	}

	// Verify chain
	isValid, err := dt.hasher.VerifyChain(spans)

	// Log verification
	dt.db.ExecContext(ctx,
		`INSERT INTO hash_chain_verifications 
         (trace_id, verification_type, is_valid, error_message, verified_by)
         VALUES ($1, 'chain_integrity', $2, $3, 'system')`,
		traceID, isValid, func() *string {
			if err != nil {
				s := err.Error()
				return &s
			}
			return nil
		}(),
	)

	return isValid, err
}

// RecordMetric logs a performance metric
func (dt *DecisionTracer) RecordMetric(ctx context.Context, traceID int, spanID *int, metricName string, value float64, unit string) error {
	_, err := dt.db.ExecContext(ctx,
		`INSERT INTO decision_metrics (trace_id, span_id, metric_name, metric_value, metric_unit)
         VALUES ($1, $2, $3, $4, $5)`,
		traceID, spanID, metricName, value, unit,
	)
	return err
}

// GetTrace retrieves a complete trace with all spans
func (dt *DecisionTracer) GetTrace(ctx context.Context, traceID int) (*Trace, []Span, error) {
	// Get trace
	var trace Trace
	err := dt.db.QueryRowContext(ctx,
		`SELECT id, trade_id, trace_type, status, start_time, end_time,
                total_duration_ms, final_decision, confidence_score, is_anchored, merkle_root
         FROM decision_traces WHERE id = $1`,
		traceID,
	).Scan(
		&trace.ID, &trace.TradeID, &trace.TraceType, &trace.Status,
		&trace.StartTime, &trace.EndTime, &trace.TotalDurationMs,
		&trace.FinalDecision, &trace.ConfidenceScore, &trace.IsAnchored, &trace.MerkleRoot,
	)
	if err != nil {
		return nil, nil, err
	}

	// Get spans
	rows, err := dt.db.QueryContext(ctx,
		`SELECT id, trace_id, parent_span_id, span_name, span_type, chain_position,
                start_time, end_time, duration_ms, input_data, output_data,
                decision_reasoning, confidence_score, status, error_message,
                sha256_hash, previous_hash, data_snapshot
         FROM decision_spans
         WHERE trace_id = $1
         ORDER BY chain_position`,
		traceID,
	)
	if err != nil {
		return &trace, nil, err
	}
	defer rows.Close()

	var spans []Span
	for rows.Next() {
		var span Span
		var inputJSON, outputJSON []byte
		var reasoning, errMsg sql.NullString
		var confidence sql.NullFloat64
		var parentID sql.NullInt64
		var endTime sql.NullTime
		var duration sql.NullInt64

		err := rows.Scan(
			&span.ID, &span.TraceID, &parentID, &span.SpanName, &span.SpanType,
			&span.ChainPosition, &span.StartTime, &endTime, &duration,
			&inputJSON, &outputJSON, &reasoning, &confidence, &span.Status, &errMsg,
			&span.SHA256Hash, &span.PreviousHash, &span.DataSnapshot,
		)
		if err != nil {
			return &trace, nil, err
		}

		json.Unmarshal(inputJSON, &span.InputData)
		json.Unmarshal(outputJSON, &span.OutputData)

		if parentID.Valid {
			pid := int(parentID.Int64)
			span.ParentSpanID = &pid
		}
		if endTime.Valid {
			span.EndTime = &endTime.Time
		}
		if duration.Valid {
			d := int(duration.Int64)
			span.DurationMs = &d
		}
		if reasoning.Valid {
			span.DecisionReasoning = &reasoning.String
		}
		if confidence.Valid {
			span.ConfidenceScore = &confidence.Float64
		}
		if errMsg.Valid {
			span.ErrorMessage = &errMsg.String
		}

		spans = append(spans, span)
	}

	return &trace, spans, nil
}
