package workflow

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Mock repository for testing
type mockWorkflowRepository struct {
	workflows map[int64]*Workflow
	nextID    int64
}

func newMockWorkflowRepository() *mockWorkflowRepository {
	return &mockWorkflowRepository{
		workflows: make(map[int64]*Workflow),
		nextID:    1,
	}
}

func (m *mockWorkflowRepository) CreateWorkflow(ctx context.Context, workflow *Workflow) (int64, error) {
	workflow.Id = m.nextID
	workflow.CreatedAt = time.Now().Unix()
	workflow.UpdatedAt = time.Now().Unix()
	m.workflows[m.nextID] = workflow
	m.nextID++
	return workflow.Id, nil
}

func (m *mockWorkflowRepository) UpdateWorkflow(ctx context.Context, workflow *Workflow) error {
	if _, exists := m.workflows[workflow.Id]; !exists {
		return gerror.New("workflow not found")
	}
	workflow.UpdatedAt = time.Now().Unix()
	m.workflows[workflow.Id] = workflow
	return nil
}

func (m *mockWorkflowRepository) DeleteWorkflow(ctx context.Context, workflowId int64) error {
	if _, exists := m.workflows[workflowId]; !exists {
		return gerror.New("workflow not found")
	}
	delete(m.workflows, workflowId)
	return nil
}

func (m *mockWorkflowRepository) GetWorkflowById(ctx context.Context, workflowId int64) (*Workflow, error) {
	workflow, exists := m.workflows[workflowId]
	if !exists {
		return nil, nil
	}
	return workflow, nil
}

func (m *mockWorkflowRepository) ListWorkflows(ctx context.Context, page, pageSize int, keyword string, status int) ([]*Workflow, int, error) {
	var allWorkflows []*Workflow
	for _, workflow := range m.workflows {
		allWorkflows = append(allWorkflows, workflow)
	}

	var filteredWorkflows []*Workflow
	for _, workflow := range allWorkflows {
		// Filter by status
		if status != -1 && workflow.Status != status {
			continue
		}
		// Filter by keyword
		if keyword != "" && !strings.Contains(workflow.Name, keyword) && !strings.Contains(workflow.Description, keyword) {
			continue
		}
		filteredWorkflows = append(filteredWorkflows, workflow)
	}

	total := len(filteredWorkflows)

	// Simple pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(filteredWorkflows) {
		return []*Workflow{}, total, nil
	}
	if end > len(filteredWorkflows) {
		end = len(filteredWorkflows)
	}

	return filteredWorkflows[start:end], total, nil
}

// Stub implementations for other methods
func (m *mockWorkflowRepository) CreateWorkflowVersion(ctx context.Context, version *WorkflowVersion) (int64, error) {
	return 1, nil
}
func (m *mockWorkflowRepository) GetWorkflowVersions(ctx context.Context, workflowId int64) ([]*WorkflowVersion, error) {
	return []*WorkflowVersion{}, nil
}
func (m *mockWorkflowRepository) GetWorkflowVersionById(ctx context.Context, versionId int64) (*WorkflowVersion, error) {
	return nil, nil
}
func (m *mockWorkflowRepository) RecordExecution(ctx context.Context, execution *WorkflowExecution) (int64, error) {
	return 1, nil
}
func (m *mockWorkflowRepository) ListExecutions(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowExecution, int, error) {
	return []*WorkflowExecution{}, 0, nil
}
func (m *mockWorkflowRepository) CreateLog(ctx context.Context, log *WorkflowLog) (int64, error) {
	return 1, nil
}
func (m *mockWorkflowRepository) ListLogs(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowLog, int, error) {
	return []*WorkflowLog{}, 0, nil
}
func (m *mockWorkflowRepository) UpdateWorkflowStatus(ctx context.Context, id int64, status int) error {
	if wf, exists := m.workflows[id]; exists {
		wf.Status = status
		wf.UpdatedAt = time.Now().Unix()
	}
	return nil
}

func TestWorkflowService_CreateWorkflow(t *testing.T) {
	mockRepo := newMockWorkflowRepository()

	// Replace the global repository with mock
	originalRepo := workflowRepoInstance
	workflowRepoInstance = mockRepo
	defer func() { workflowRepoInstance = originalRepo }()

	service := GetWorkflowService()

	tests := []struct {
		name     string
		workflow *Workflow
		wantErr  bool
	}{
		{
			name: "Create valid workflow",
			workflow: &Workflow{
				Name:        "Test Workflow",
				Description: "Test Description",
				Status:      1,
				Version:     1,
				Trigger:     "contact_created",
			},
			wantErr: false,
		},
		{
			name: "Create workflow with empty name",
			workflow: &Workflow{
				Name:        "",
				Description: "Test Description",
				Status:      1,
				Version:     1,
				Trigger:     "contact_created",
			},
			wantErr: true, // Service validates required fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := service.CreateWorkflow(context.Background(), tt.workflow)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id <= 0 {
				t.Errorf("CreateWorkflow() returned invalid id = %v", id)
			}
		})
	}
}

func TestWorkflowService_GetWorkflow(t *testing.T) {
	mockRepo := newMockWorkflowRepository()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Test Workflow",
		Description: "Test Description",
		Status:      1,
		Version:     1,
		Trigger:     "contact_created",
	}
	id, _ := mockRepo.CreateWorkflow(context.Background(), testWorkflow)

	// Replace the global repository with mock
	originalRepo := workflowRepoInstance
	workflowRepoInstance = mockRepo
	defer func() { workflowRepoInstance = originalRepo }()

	service := GetWorkflowService()

	tests := []struct {
		name       string
		workflowId int64
		wantNil    bool
		wantErr    bool
	}{
		{
			name:       "Get existing workflow",
			workflowId: id,
			wantNil:    false,
			wantErr:    false,
		},
		{
			name:       "Get non-existing workflow",
			workflowId: 999,
			wantNil:    true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow, err := service.GetWorkflow(context.Background(), tt.workflowId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (workflow == nil) != tt.wantNil {
				t.Errorf("GetWorkflow() returned nil = %v, wantNil %v", workflow == nil, tt.wantNil)
			}
		})
	}
}

func TestWorkflowService_ListWorkflows(t *testing.T) {
	mockRepo := newMockWorkflowRepository()

	// Create test workflows
	for i := 1; i <= 5; i++ {
		workflow := &Workflow{
			Name:        fmt.Sprintf("Workflow %d", i),
			Description: fmt.Sprintf("Description %d", i),
			Status:      i % 2, // Alternate between 0 and 1
			Version:     1,
			Trigger:     "contact_created",
		}
		mockRepo.CreateWorkflow(context.Background(), workflow)
	}

	// Replace the global repository with mock
	originalRepo := workflowRepoInstance
	workflowRepoInstance = mockRepo
	defer func() { workflowRepoInstance = originalRepo }()

	service := GetWorkflowService()

	t.Run("List all workflows", func(t *testing.T) {
		workflows, total, err := service.ListWorkflows(context.Background(), 1, 10, "", -1)
		if err != nil {
			t.Errorf("ListWorkflows() error = %v", err)
			return
		}
		if len(workflows) != 5 {
			t.Errorf("ListWorkflows() returned %d workflows, want 5", len(workflows))
		}
		if total != 5 {
			t.Errorf("ListWorkflows() returned total %d, want 5", total)
		}
	})

	t.Run("List active workflows", func(t *testing.T) {
		workflows, total, err := service.ListWorkflows(context.Background(), 1, 10, "", 1)
		if err != nil {
			t.Errorf("ListWorkflows() error = %v", err)
			return
		}
		// Should return workflows with status 1 (created 1, 3, 5)
		expectedCount := 3
		if len(workflows) != expectedCount {
			t.Errorf("ListWorkflows() returned %d workflows, want %d", len(workflows), expectedCount)
		}
		if total != expectedCount {
			t.Errorf("ListWorkflows() returned total %d, want %d", total, expectedCount)
		}
	})

	t.Run("List with pagination", func(t *testing.T) {
		workflows, total, err := service.ListWorkflows(context.Background(), 1, 2, "", -1)
		if err != nil {
			t.Errorf("ListWorkflows() error = %v", err)
			return
		}
		if len(workflows) != 2 {
			t.Errorf("ListWorkflows() returned %d workflows, want 2", len(workflows))
		}
		if total != 5 {
			t.Errorf("ListWorkflows() returned total %d, want 5", total)
		}
	})
}

func TestWorkflowService_UpdateWorkflow(t *testing.T) {
	mockRepo := newMockWorkflowRepository()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Original Name",
		Description: "Original Description",
		Status:      1,
		Version:     1,
		Trigger:     "contact_created",
	}
	id, _ := mockRepo.CreateWorkflow(context.Background(), testWorkflow)

	// Replace the global repository with mock
	originalRepo := workflowRepoInstance
	workflowRepoInstance = mockRepo
	defer func() { workflowRepoInstance = originalRepo }()

	service := GetWorkflowService()

	t.Run("Update existing workflow", func(t *testing.T) {
		updateWorkflow := &Workflow{
			Id:          id,
			Name:        "Updated Name",
			Description: "Updated Description",
			Status:      0,
			Version:     2,
			Trigger:     "contact_updated",
		}

		err := service.UpdateWorkflow(context.Background(), updateWorkflow)
		if err != nil {
			t.Errorf("UpdateWorkflow() error = %v", err)
			return
		}

		// Verify the update
		updated, err := service.GetWorkflow(context.Background(), id)
		if err != nil {
			t.Errorf("GetWorkflow() error = %v", err)
			return
		}
		if updated.Name != "Updated Name" {
			t.Errorf("UpdateWorkflow() name = %v, want Updated Name", updated.Name)
		}
		if updated.Status != 0 {
			t.Errorf("UpdateWorkflow() status = %v, want 0", updated.Status)
		}
	})

	t.Run("Update non-existing workflow", func(t *testing.T) {
		updateWorkflow := &Workflow{
			Id:          999,
			Name:        "Non-existing",
			Description: "Should fail",
		}

		err := service.UpdateWorkflow(context.Background(), updateWorkflow)
		if err == nil {
			t.Errorf("UpdateWorkflow() expected error for non-existing workflow")
		}
	})
}

func TestWorkflowService_DeleteWorkflow(t *testing.T) {
	mockRepo := newMockWorkflowRepository()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Test Workflow",
		Description: "Test Description",
		Status:      1,
		Version:     1,
		Trigger:     "contact_created",
	}
	id, _ := mockRepo.CreateWorkflow(context.Background(), testWorkflow)

	// Replace the global repository with mock
	originalRepo := workflowRepoInstance
	workflowRepoInstance = mockRepo
	defer func() { workflowRepoInstance = originalRepo }()

	service := GetWorkflowService()

	t.Run("Delete existing workflow", func(t *testing.T) {
		err := service.DeleteWorkflow(context.Background(), id)
		if err != nil {
			t.Errorf("DeleteWorkflow() error = %v", err)
			return
		}

		// Verify deletion
		workflow, err := service.GetWorkflow(context.Background(), id)
		if err != nil {
			t.Errorf("GetWorkflow() error = %v", err)
			return
		}
		if workflow != nil {
			t.Errorf("DeleteWorkflow() workflow still exists after deletion")
		}
	})

	t.Run("Delete non-existing workflow", func(t *testing.T) {
		err := service.DeleteWorkflow(context.Background(), 999)
		if err == nil {
			t.Errorf("DeleteWorkflow() expected error for non-existing workflow")
		}
	})
}

func TestWorkflowService_DuplicateWorkflow(t *testing.T) {
	mockRepo := newMockWorkflowRepository()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Original Workflow",
		Description: "Original Description",
		Status:      1,
		Version:     1,
		Trigger:     "contact_created",
	}
	id, _ := mockRepo.CreateWorkflow(context.Background(), testWorkflow)

	// Replace the global repository with mock
	originalRepo := workflowRepoInstance
	workflowRepoInstance = mockRepo
	defer func() { workflowRepoInstance = originalRepo }()

	service := GetWorkflowService()

	t.Run("Duplicate existing workflow", func(t *testing.T) {
		newWorkflow, err := service.DuplicateWorkflow(context.Background(), id)
		if err != nil {
			t.Errorf("DuplicateWorkflow() error = %v", err)
			return
		}
		if newWorkflow == nil {
			t.Errorf("DuplicateWorkflow() returned nil workflow")
		}

		// Verify the duplicate exists and has correct name
		if newWorkflow.Name != "Original Workflow Copy" {
			t.Errorf("DuplicateWorkflow() name = %v, want 'Original Workflow Copy'", newWorkflow.Name)
		}
	})

	t.Run("Duplicate non-existing workflow", func(t *testing.T) {
		_, err := service.DuplicateWorkflow(context.Background(), 999)
		if err == nil {
			t.Errorf("DuplicateWorkflow() expected error for non-existing workflow")
		}
	})
}

/*
func TestWorkflowService_GetWorkflowStatistics(t *testing.T) {
	mockRepo := newMockWorkflowRepository()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Test Workflow",
		Description: "Test Description",
		Status:      1,
		Version:     1,
		Trigger:     "contact_created",
	}
	id, _ := mockRepo.CreateWorkflow(context.Background(), testWorkflow)

	// Replace the global repository with mock
	originalRepo := workflowRepoInstance
	workflowRepoInstance = mockRepo
	defer func() { workflowRepoInstance = originalRepo }()

	service := GetWorkflowService()

	t.Run("Get statistics for existing workflow", func(t *testing.T) {
		stats, err := service.GetWorkflowStatistics(context.Background(), id)
		if err != nil {
			t.Errorf("GetWorkflowStatistics() error = %v", err)
			return
		}
		if stats == nil {
			t.Errorf("GetWorkflowStatistics() returned nil")
			return
		}
		if stats.WorkflowId != id {
			t.Errorf("GetWorkflowStatistics() workflowId = %v, want %v", stats.WorkflowId, id)
		}
		// Mock returns 0 for all stats
		if stats.TotalExecutions != 0 {
			t.Errorf("GetWorkflowStatistics() totalExecutions = %v, want 0", stats.TotalExecutions)
		}
	})

	t.Run("Get statistics for non-existing workflow", func(t *testing.T) {
		stats, err := service.GetWorkflowStatistics(context.Background(), 999)
		if err != nil {
			t.Errorf("GetWorkflowStatistics() error = %v", err)
			return
		}
		if stats == nil {
			t.Errorf("GetWorkflowStatistics() returned nil")
			return
		}
		if stats.WorkflowId != 999 {
			t.Errorf("GetWorkflowStatistics() workflowId = %v, want 999", stats.WorkflowId)
		}
	})
}
*/
