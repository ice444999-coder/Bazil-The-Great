package repositories

import (
	"database/sql"
	"encoding/json"

	"ares_api/internal/models"
)

type AgentRepository struct {
	db *sql.DB
}

func NewAgentRepository(db *sql.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// ========== AGENT OPERATIONS ==========

// GetAllAgents retrieves all registered agents
func (r *AgentRepository) GetAllAgents() ([]models.Agent, error) {
	query := `
		SELECT agent_id, agent_name, agent_type, capabilities, status,
		       current_task_id, total_tasks_completed, success_rate,
		       avg_completion_time_ms, last_active_at, created_at
		FROM agent_registry
		ORDER BY agent_name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []models.Agent
	for rows.Next() {
		var agent models.Agent
		err := rows.Scan(
			&agent.AgentID, &agent.AgentName, &agent.AgentType, &agent.Capabilities,
			&agent.Status, &agent.CurrentTaskID, &agent.TotalTasksCompleted,
			&agent.SuccessRate, &agent.AvgCompletionTimeMs, &agent.LastActiveAt, &agent.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

// GetAgentByName retrieves a specific agent by name
func (r *AgentRepository) GetAgentByName(name string) (*models.Agent, error) {
	query := `
		SELECT agent_id, agent_name, agent_type, capabilities, status,
		       current_task_id, total_tasks_completed, success_rate,
		       avg_completion_time_ms, last_active_at, created_at
		FROM agent_registry
		WHERE agent_name = $1
	`
	var agent models.Agent
	err := r.db.QueryRow(query, name).Scan(
		&agent.AgentID, &agent.AgentName, &agent.AgentType, &agent.Capabilities,
		&agent.Status, &agent.CurrentTaskID, &agent.TotalTasksCompleted,
		&agent.SuccessRate, &agent.AvgCompletionTimeMs, &agent.LastActiveAt, &agent.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// UpdateAgentStatus updates an agent's status
func (r *AgentRepository) UpdateAgentStatus(agentName, status string, taskID *string) error {
	query := `
		UPDATE agent_registry
		SET status = $1, current_task_id = $2, last_active_at = NOW()
		WHERE agent_name = $3
	`
	_, err := r.db.Exec(query, status, taskID, agentName)
	return err
}

// ========== TASK OPERATIONS ==========

// CreateTask creates a new task in the queue
func (r *AgentRepository) CreateTask(req *models.CreateTaskRequest, createdBy string) (string, error) {
	filePathsJSON, _ := json.Marshal(req.FilePaths)
	dependsOnJSON, _ := json.Marshal(req.DependsOnTaskIDs)
	contextJSON, _ := json.Marshal(req.Context)

	if req.Priority == 0 {
		req.Priority = 5 // Default priority
	}

	query := `
		INSERT INTO task_queue (task_type, description, priority, created_by, file_paths, depends_on_task_ids, context, deadline, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending')
		RETURNING task_id
	`
	var taskID string
	err := r.db.QueryRow(query, req.TaskType, req.Description, req.Priority, createdBy,
		filePathsJSON, dependsOnJSON, contextJSON, req.Deadline).Scan(&taskID)

	return taskID, err
}

// GetTaskByID retrieves a task by ID
func (r *AgentRepository) GetTaskByID(taskID string) (*models.Task, error) {
	query := `
		SELECT task_id, task_type, priority, status, created_by, assigned_to_agent,
		       file_paths, depends_on_task_ids, description, context,
		       created_at, assigned_at, started_at, completed_at, deadline, result, error_log, retry_count
		FROM task_queue
		WHERE task_id = $1
	`
	var task models.Task
	err := r.db.QueryRow(query, taskID).Scan(
		&task.TaskID, &task.TaskType, &task.Priority, &task.Status, &task.CreatedBy,
		&task.AssignedToAgent, &task.FilePaths, &task.DependsOnTaskIDs, &task.Description,
		&task.Context, &task.CreatedAt, &task.AssignedAt, &task.StartedAt, &task.CompletedAt,
		&task.Deadline, &task.Result, &task.ErrorLog, &task.RetryCount,
	)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetPendingTasks retrieves all pending tasks
func (r *AgentRepository) GetPendingTasks(limit int) ([]models.Task, error) {
	if limit == 0 {
		limit = 10
	}

	query := `
		SELECT task_id, task_type, priority, status, created_by, assigned_to_agent,
		       file_paths, depends_on_task_ids, description, context,
		       created_at, assigned_at, started_at, completed_at, deadline, result, error_log, retry_count
		FROM task_queue
		WHERE status = 'pending'
		ORDER BY priority DESC, created_at ASC
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.TaskID, &task.TaskType, &task.Priority, &task.Status, &task.CreatedBy,
			&task.AssignedToAgent, &task.FilePaths, &task.DependsOnTaskIDs, &task.Description,
			&task.Context, &task.CreatedAt, &task.AssignedAt, &task.StartedAt, &task.CompletedAt,
			&task.Deadline, &task.Result, &task.ErrorLog, &task.RetryCount,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// AssignTask assigns a task to an agent
func (r *AgentRepository) AssignTask(taskID, agentName string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update task
	_, err = tx.Exec(`
		UPDATE task_queue
		SET assigned_to_agent = $1, assigned_at = NOW(), status = 'assigned'
		WHERE task_id = $2
	`, agentName, taskID)
	if err != nil {
		return err
	}

	// Update agent
	_, err = tx.Exec(`
		UPDATE agent_registry
		SET current_task_id = $1, status = 'busy', last_active_at = NOW()
		WHERE agent_name = $2
	`, taskID, agentName)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CompleteTask marks a task as completed
func (r *AgentRepository) CompleteTask(taskID string, result map[string]interface{}) error {
	resultJSON, _ := json.Marshal(result)

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get task info
	var agentName string
	err = tx.QueryRow(`SELECT assigned_to_agent FROM task_queue WHERE task_id = $1`, taskID).Scan(&agentName)
	if err != nil {
		return err
	}

	// Update task
	_, err = tx.Exec(`
		UPDATE task_queue
		SET status = 'completed', completed_at = NOW(), result = $1
		WHERE task_id = $2
	`, resultJSON, taskID)
	if err != nil {
		return err
	}

	// Update agent
	_, err = tx.Exec(`
		UPDATE agent_registry
		SET status = 'idle', current_task_id = NULL, 
		    total_tasks_completed = total_tasks_completed + 1,
		    last_active_at = NOW()
		WHERE agent_name = $1
	`, agentName)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// FailTask marks a task as failed
func (r *AgentRepository) FailTask(taskID, errorLog string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get task info
	var agentName string
	var retryCount int
	err = tx.QueryRow(`
		SELECT assigned_to_agent, retry_count 
		FROM task_queue 
		WHERE task_id = $1
	`, taskID).Scan(&agentName, &retryCount)
	if err != nil {
		return err
	}

	// Update task
	_, err = tx.Exec(`
		UPDATE task_queue
		SET status = 'failed', error_log = $1, retry_count = retry_count + 1
		WHERE task_id = $2
	`, errorLog, taskID)
	if err != nil {
		return err
	}

	// Update agent
	_, err = tx.Exec(`
		UPDATE agent_registry
		SET status = 'idle', current_task_id = NULL, last_active_at = NOW()
		WHERE agent_name = $1
	`, agentName)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ========== FILE REGISTRY OPERATIONS ==========

// RegisterFile registers a new file in the registry
func (r *AgentRepository) RegisterFile(file *models.FileRegistry) error {
	depsJSON, _ := json.Marshal(file.Dependencies)

	query := `
		INSERT INTO file_registry 
		(file_path, file_type, file_hash, owner_agent, created_by, status, purpose, dependencies, size_bytes, line_count, language)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (file_path) DO UPDATE SET
			file_hash = EXCLUDED.file_hash,
			last_modified_by = EXCLUDED.owner_agent,
			updated_at = NOW()
		RETURNING file_id
	`
	return r.db.QueryRow(query, file.FilePath, file.FileType, file.FileHash, file.OwnerAgent,
		file.CreatedBy, file.Status, file.Purpose, depsJSON, file.SizeBytes, file.LineCount, file.Language).Scan(&file.FileID)
}

// GetFileByPath retrieves a file by path
func (r *AgentRepository) GetFileByPath(filePath string) (*models.FileRegistry, error) {
	query := `
		SELECT file_id, file_path, file_type, file_hash, owner_agent, created_by, last_modified_by,
		       status, purpose, dependencies, created_at, updated_at, last_tested_at, test_status,
		       build_required, deployed, size_bytes, line_count, language
		FROM file_registry
		WHERE file_path = $1
	`
	var file models.FileRegistry
	err := r.db.QueryRow(query, filePath).Scan(
		&file.FileID, &file.FilePath, &file.FileType, &file.FileHash, &file.OwnerAgent,
		&file.CreatedBy, &file.LastModifiedBy, &file.Status, &file.Purpose, &file.Dependencies,
		&file.CreatedAt, &file.UpdatedAt, &file.LastTestedAt, &file.TestStatus,
		&file.BuildRequired, &file.Deployed, &file.SizeBytes, &file.LineCount, &file.Language,
	)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetAllFiles retrieves all files from registry
func (r *AgentRepository) GetAllFiles() ([]models.FileRegistry, error) {
	query := `
		SELECT file_id, file_path, file_type, file_hash, owner_agent, created_by, last_modified_by,
		       status, purpose, dependencies, created_at, updated_at, last_tested_at, test_status,
		       build_required, deployed, size_bytes, line_count, language
		FROM file_registry
		ORDER BY file_path
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.FileRegistry
	for rows.Next() {
		var file models.FileRegistry
		err := rows.Scan(
			&file.FileID, &file.FilePath, &file.FileType, &file.FileHash, &file.OwnerAgent,
			&file.CreatedBy, &file.LastModifiedBy, &file.Status, &file.Purpose, &file.Dependencies,
			&file.CreatedAt, &file.UpdatedAt, &file.LastTestedAt, &file.TestStatus,
			&file.BuildRequired, &file.Deployed, &file.SizeBytes, &file.LineCount, &file.Language,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

// ========== HISTORY OPERATIONS ==========

// LogTaskHistory logs a task execution to history
func (r *AgentRepository) LogTaskHistory(history *models.AgentTaskHistory) error {
	query := `
		INSERT INTO agent_task_history 
		(agent_name, task_id, task_type, file_id, action_type, success, duration_ms, error_message, learned_pattern, cost_tokens)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING history_id
	`
	return r.db.QueryRow(query, history.AgentName, history.TaskID, history.TaskType, history.FileID,
		history.ActionType, history.Success, history.DurationMs, history.ErrorMessage,
		history.LearnedPattern, history.CostTokens).Scan(&history.HistoryID)
}

// GetAgentPerformance retrieves performance metrics for an agent
func (r *AgentRepository) GetAgentPerformance(agentName string, limit int) ([]models.AgentTaskHistory, error) {
	if limit == 0 {
		limit = 100
	}

	query := `
		SELECT history_id, agent_name, task_id, task_type, file_id, action_type,
		       success, duration_ms, error_message, learned_pattern, cost_tokens, created_at
		FROM agent_task_history
		WHERE agent_name = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(query, agentName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.AgentTaskHistory
	for rows.Next() {
		var h models.AgentTaskHistory
		err := rows.Scan(
			&h.HistoryID, &h.AgentName, &h.TaskID, &h.TaskType, &h.FileID, &h.ActionType,
			&h.Success, &h.DurationMs, &h.ErrorMessage, &h.LearnedPattern, &h.CostTokens, &h.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

// LogBuild logs a build to history
func (r *AgentRepository) LogBuild(build *models.BuildHistory) error {
	filesChangedJSON, _ := json.Marshal(build.FilesChanged)

	query := `
		INSERT INTO build_history 
		(triggered_by, files_changed, success, duration_ms, error_log, warnings, binary_hash, deployed, git_commit_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING build_id, build_number
	`
	return r.db.QueryRow(query, build.TriggeredBy, filesChangedJSON, build.Success, build.DurationMs,
		build.ErrorLog, build.Warnings, build.BinaryHash, build.Deployed, build.GitCommitHash).
		Scan(&build.BuildID, &build.BuildNumber)
}

// GetRecentBuilds retrieves recent builds
func (r *AgentRepository) GetRecentBuilds(limit int) ([]models.BuildHistory, error) {
	if limit == 0 {
		limit = 20
	}

	query := `
		SELECT build_id, build_number, triggered_by, files_changed, success, duration_ms,
		       error_log, warnings, binary_hash, deployed, git_commit_hash, created_at
		FROM build_history
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builds []models.BuildHistory
	for rows.Next() {
		var build models.BuildHistory
		err := rows.Scan(
			&build.BuildID, &build.BuildNumber, &build.TriggeredBy, &build.FilesChanged, &build.Success,
			&build.DurationMs, &build.ErrorLog, &build.Warnings, &build.BinaryHash, &build.Deployed,
			&build.GitCommitHash, &build.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}
	return builds, nil
}
