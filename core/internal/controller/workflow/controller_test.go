package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"billionmail-core/api/workflow/v1"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// Mock workflow service for testing
type mockWorkflowService struct {
	workflows map[int64]*Workflow
	nextID    int64
}

func newMockWorkflowService() *mockWorkflowService {
	return &mockWorkflowService{
		workflows: make(map[int64]*Workflow),
		nextID:    1,
	}
}

func (m *mockWorkflowService) CreateWorkflow(ctx context.Context, workflow *Workflow) (int64, error) {
	workflow.Id = m.nextID
	m.workflows[m.nextID] = workflow
	m.nextID++
	return workflow.Id, nil
}

func (m *mockWorkflowService) GetWorkflow(ctx context.Context, workflowId int64) (*Workflow, error) {
	workflow, exists := m.workflows[workflowId]
	if !exists {
		return nil, nil
	}
	return workflow, nil
}

func (m *mockWorkflowService) ListWorkflows(ctx context.Context, page, pageSize int, keyword string, status int) ([]*Workflow, int, error) {
	var workflows []*Workflow
	total := 0

	for _, workflow := range m.workflows {
		if status != -1 && workflow.Status != status {
			continue
		}
		if keyword != "" && !strings.Contains(workflow.Name, keyword) && !strings.Contains(workflow.Description, keyword) {
			continue
		}
		total++
		workflows = append(workflows, workflow)
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(workflows) {
		return []*Workflow{}, total, nil
	}
	if end > len(workflows) {
		end = len(workflows)
	}

	return workflows[start:end], total, nil
}

func (m *mockWorkflowService) UpdateWorkflow(ctx context.Context, workflow *Workflow) error {
	if _, exists := m.workflows[workflow.Id]; !exists {
		return gerror.New("workflow not found")
	}
	m.workflows[workflow.Id] = workflow
	return nil
}

func (m *mockWorkflowService) DeleteWorkflow(ctx context.Context, workflowId int64) error {
	if _, exists := m.workflows[workflowId]; !exists {
		return gerror.New("workflow not found")
	}
	delete(m.workflows, workflowId)
	return nil
}

func (m *mockWorkflowService) DuplicateWorkflow(ctx context.Context, workflowId int64) (int64, error) {
	original, exists := m.workflows[workflowId]
	if !exists {
		return 0, gerror.New("workflow not found")
	}

	duplicate := &Workflow{
		Name:        original.Name + " (Copy)",
		Description: original.Description,
		Status:      0, // Disabled by default
		Version:     1,
		Trigger:     original.Trigger,
		Nodes:       original.Nodes,
		Connections: original.Connections,
		Metadata:    original.Metadata,
	}

	return m.CreateWorkflow(ctx, duplicate)
}

func (m *mockWorkflowService) ToggleWorkflow(ctx context.Context, workflowId int64) error {
	workflow, exists := m.workflows[workflowId]
	if !exists {
		return gerror.New("workflow not found")
	}
	workflow.Status = 1 - workflow.Status // Toggle between 0 and 1
	return nil
}

func (m *mockWorkflowService) GetWorkflowStatistics(ctx context.Context, workflowId int64) (*WorkflowStatistics, error) {
	return &WorkflowStatistics{
		WorkflowId:      workflowId,
		TotalExecutions: 10,
		SuccessCount:    8,
		FailureCount:    2,
		AverageDuration: 1500,
		LastRunAt:       1623456789,
	}, nil
}

func TestControllerV1_Create(t *testing.T) {
	mockService := newMockWorkflowService()

	// Replace the service with mock
	originalService := service.WorkflowService()
	defer func() {
		// Restore original service (this is a simplified approach)
	}()

	// Create test server
	s := g.Server()
	s.BindHandler("/api/v1/workflow", func(r *ghttp.Request) {
		req := &v1.CreateWorkflowReq{}
		if err := r.Parse(req); err != nil {
			r.Response.WriteJsonExit(g.Map{"error": err.Error()})
			return
		}

		workflow := &Workflow{
			Name:        req.Name,
			Description: req.Description,
			Status:      gconv.Int(req.Status),
			Trigger:     req.Trigger,
		}

		id, err := mockService.CreateWorkflow(r.Context(), workflow)
		if err != nil {
			r.Response.WriteJsonExit(g.Map{"error": err.Error()})
			return
		}

		r.Response.WriteJsonExit(g.Map{"id": id})
	})

	s.SetPort(0) // Use random port
	s.Start()

	defer s.Shutdown()

	tests := []struct {
		name         string
		requestBody  interface{}
		expectedCode int
		expectError  bool
	}{
		{
			name: "Create valid workflow",
			requestBody: v1.CreateWorkflowReq{
				Name:        "Test Workflow",
				Description: "Test Description",
				Status:      1,
				Trigger:     "contact_created",
			},
			expectedCode: 200,
			expectError:  false,
		},
		{
			name: "Create workflow with empty name",
			requestBody: v1.CreateWorkflowReq{
				Name:        "",
				Description: "Test Description",
				Status:      1,
				Trigger:     "contact_created",
			},
			expectedCode: 200,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/workflow", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create a simple handler for testing
			handler := func(w http.ResponseWriter, r *http.Request) {
				var req v1.CreateWorkflowReq
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				workflow := &Workflow{
					Name:        req.Name,
					Description: req.Description,
					Status:      gconv.Int(req.Status),
					Trigger:     req.Trigger,
				}

				id, err := mockService.CreateWorkflow(context.Background(), workflow)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(g.Map{"id": id})
			}

			handler(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			var response g.Map
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Errorf("Failed to decode response: %v", err)
				return
			}

			if !tt.expectError {
				if _, exists := response["id"]; !exists {
					t.Errorf("Expected 'id' in response, got: %v", response)
				}
			}
		})
	}
}

func TestControllerV1_Get(t *testing.T) {
	mockService := newMockWorkflowService()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Test Workflow",
		Description: "Test Description",
		Status:      1,
		Trigger:     "contact_created",
	}
	id, _ := mockService.CreateWorkflow(context.Background(), testWorkflow)

	tests := []struct {
		name         string
		workflowId   string
		expectedCode int
		expectError  bool
	}{
		{
			name:         "Get existing workflow",
			workflowId:   gconv.String(id),
			expectedCode: 200,
			expectError:  false,
		},
		{
			name:         "Get non-existing workflow",
			workflowId:   "999",
			expectedCode: 200,
			expectError:  true,
		},
		{
			name:         "Get workflow with invalid ID",
			workflowId:   "invalid",
			expectedCode: 400,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest("GET", "/api/v1/workflow/"+tt.workflowId, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create handler
			handler := func(w http.ResponseWriter, r *http.Request) {
				// Simple path parsing
				path := strings.TrimPrefix(r.URL.Path, "/api/v1/workflow/")
				workflowId := gconv.Int64(path)
				if workflowId == 0 && path != "0" {
					http.Error(w, "Invalid workflow ID", http.StatusBadRequest)
					return
				}

				workflow, err := mockService.GetWorkflow(context.Background(), workflowId)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if workflow == nil {
					http.Error(w, "Workflow not found", http.StatusNotFound)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(workflow)
			}

			handler(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}

func TestControllerV1_List(t *testing.T) {
	mockService := newMockWorkflowService()

	// Create test workflows
	for i := 1; i <= 3; i++ {
		workflow := &Workflow{
			Name:        gconv.Stringf("Workflow %d", i),
			Description: gconv.Stringf("Description %d", i),
			Status:      1,
			Trigger:     "contact_created",
		}
		mockService.CreateWorkflow(context.Background(), workflow)
	}

	t.Run("List all workflows", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/workflow", nil)
		w := httptest.NewRecorder()

		handler := func(w http.ResponseWriter, r *http.Request) {
			workflows, total, err := mockService.ListWorkflows(context.Background(), 1, 10, "", -1)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := g.Map{
				"workflows": workflows,
				"total":     total,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}

		handler(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		var response g.Map
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode response: %v", err)
			return
		}

		workflows := response["workflows"].([]interface{})
		if len(workflows) != 3 {
			t.Errorf("Expected 3 workflows, got %d", len(workflows))
		}

		total := gconv.Int(response["total"])
		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}
	})
}

func TestControllerV1_Update(t *testing.T) {
	mockService := newMockWorkflowService()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Original Name",
		Description: "Original Description",
		Status:      1,
		Trigger:     "contact_created",
	}
	id, _ := mockService.CreateWorkflow(context.Background(), testWorkflow)

	t.Run("Update existing workflow", func(t *testing.T) {
		updateReq := v1.UpdateWorkflowReq{
			Id:          gconv.String(id),
			Name:        "Updated Name",
			Description: "Updated Description",
			Status:      0,
			Trigger:     "contact_updated",
		}

		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", "/api/v1/workflow/"+gconv.String(id), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler := func(w http.ResponseWriter, r *http.Request) {
			var req v1.UpdateWorkflowReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			workflowId := gconv.Int64(req.Id)
			existing, err := mockService.GetWorkflow(context.Background(), workflowId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if existing == nil {
				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}

			existing.Name = req.Name
			existing.Description = req.Description
			existing.Status = gconv.Int(req.Status)
			existing.Trigger = req.Trigger

			err = mockService.UpdateWorkflow(context.Background(), existing)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		}

		handler(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
	})
}

func TestControllerV1_Delete(t *testing.T) {
	mockService := newMockWorkflowService()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Test Workflow",
		Description: "Test Description",
		Status:      1,
		Trigger:     "contact_created",
	}
	id, _ := mockService.CreateWorkflow(context.Background(), testWorkflow)

	t.Run("Delete existing workflow", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/workflow/"+gconv.String(id), nil)
		w := httptest.NewRecorder()

		handler := func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/api/v1/workflow/")
			workflowId := gconv.Int64(path)

			err := mockService.DeleteWorkflow(context.Background(), workflowId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		}

		handler(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
	})
}

func TestControllerV1_Duplicate(t *testing.T) {
	mockService := newMockWorkflowService()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Original Workflow",
		Description: "Original Description",
		Status:      1,
		Trigger:     "contact_created",
	}
	id, _ := mockService.CreateWorkflow(context.Background(), testWorkflow)

	t.Run("Duplicate existing workflow", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/workflow/"+gconv.String(id)+"/duplicate", nil)
		w := httptest.NewRecorder()

		handler := func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/api/v1/workflow/")
			path = strings.TrimSuffix(path, "/duplicate")
			workflowId := gconv.Int64(path)

			newId, err := mockService.DuplicateWorkflow(context.Background(), workflowId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(g.Map{"id": newId})
		}

		handler(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		var response g.Map
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode response: %v", err)
			return
		}

		if _, exists := response["id"]; !exists {
			t.Errorf("Expected 'id' in response")
		}
	})
}

func TestControllerV1_Toggle(t *testing.T) {
	mockService := newMockWorkflowService()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Test Workflow",
		Description: "Test Description",
		Status:      1, // Active
		Trigger:     "contact_created",
	}
	id, _ := mockService.CreateWorkflow(context.Background(), testWorkflow)

	t.Run("Toggle workflow status", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/workflow/"+gconv.String(id)+"/toggle", nil)
		w := httptest.NewRecorder()

		handler := func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/api/v1/workflow/")
			path = strings.TrimSuffix(path, "/toggle")
			workflowId := gconv.Int64(path)

			err := mockService.ToggleWorkflow(context.Background(), workflowId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		}

		handler(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
	})
}

func TestControllerV1_GetStats(t *testing.T) {
	mockService := newMockWorkflowService()

	// Create a test workflow
	testWorkflow := &Workflow{
		Name:        "Test Workflow",
		Description: "Test Description",
		Status:      1,
		Trigger:     "contact_created",
	}
	id, _ := mockService.CreateWorkflow(context.Background(), testWorkflow)

	t.Run("Get workflow statistics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/workflow/"+gconv.String(id)+"/stats", nil)
		w := httptest.NewRecorder()

		handler := func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/api/v1/workflow/")
			path = strings.TrimSuffix(path, "/stats")
			workflowId := gconv.Int64(path)

			stats, err := mockService.GetWorkflowStatistics(context.Background(), workflowId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(stats)
		}

		handler(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		var stats WorkflowStatistics
		if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
			t.Errorf("Failed to decode response: %v", err)
			return
		}

		if stats.WorkflowId != id {
			t.Errorf("Expected workflow ID %d, got %d", id, stats.WorkflowId)
		}
	})
}