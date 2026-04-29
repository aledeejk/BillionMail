-- Workflow tables migration
-- Created: 2025-04-27
-- Purpose: Create workflow management system tables

-- Workflow table: stores workflow definitions
CREATE TABLE IF NOT EXISTS workflow (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status SMALLINT NOT NULL DEFAULT 0,           -- 0: inactive, 1: active
    version INT NOT NULL DEFAULT 1,
    created_at INT NOT NULL DEFAULT 0,            -- Unix timestamp
    updated_at INT NOT NULL DEFAULT 0,            -- Unix timestamp
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    UNIQUE(name)
);

-- Workflow version table: stores workflow version history
CREATE TABLE IF NOT EXISTS workflow_version (
    id SERIAL PRIMARY KEY,
    workflow_id INT NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    version INT NOT NULL,
    content TEXT,                                 -- JSON definition of workflow
    created_at INT NOT NULL DEFAULT 0,
    created_by VARCHAR(255),
    UNIQUE(workflow_id, version)
);

-- Workflow execution table: stores workflow execution history
CREATE TABLE IF NOT EXISTS workflow_execution (
    id SERIAL PRIMARY KEY,
    workflow_id INT NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,                  -- pending, running, completed, failed
    started_at INT NOT NULL DEFAULT 0,
    completed_at INT,
    error_message TEXT,
    FOREIGN KEY(workflow_id) REFERENCES workflow(id) ON DELETE CASCADE
);

-- Workflow execution log table: stores detailed logs of execution steps
CREATE TABLE IF NOT EXISTS workflow_execution_log (
    id SERIAL PRIMARY KEY,
    execution_id INT NOT NULL REFERENCES workflow_execution(id) ON DELETE CASCADE,
    workflow_id INT NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    step_name VARCHAR(255),
    message TEXT,
    level VARCHAR(20),                            -- info, warning, error
    timestamp INT NOT NULL DEFAULT 0,
    FOREIGN KEY(execution_id) REFERENCES workflow_execution(id) ON DELETE CASCADE,
    FOREIGN KEY(workflow_id) REFERENCES workflow(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX idx_workflow_status ON workflow(status);
CREATE INDEX idx_workflow_created_at ON workflow(created_at);
CREATE INDEX idx_workflow_version_workflow_id ON workflow_version(workflow_id);
CREATE INDEX idx_workflow_execution_workflow_id ON workflow_execution(workflow_id);
CREATE INDEX idx_workflow_execution_status ON workflow_execution(status);
CREATE INDEX idx_workflow_execution_log_execution_id ON workflow_execution_log(execution_id);
CREATE INDEX idx_workflow_execution_log_workflow_id ON workflow_execution_log(workflow_id);
