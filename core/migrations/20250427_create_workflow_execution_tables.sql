CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS workflow_execution (
    id BIGSERIAL PRIMARY KEY,
    workflow_id BIGINT NOT NULL,
    contact_id UUID,
    status VARCHAR(20) DEFAULT 'running',
    started_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW()),
    finished_at BIGINT,
    current_node_id VARCHAR(128),
    version INTEGER DEFAULT 1,
    trigger VARCHAR(255) DEFAULT '',
    completed_at BIGINT DEFAULT 0,
    duration BIGINT DEFAULT 0,
    result TEXT DEFAULT '',
    error_message TEXT DEFAULT '',
    created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW()),
    updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())
);

CREATE TABLE IF NOT EXISTS workflow_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id VARCHAR(64) NOT NULL,
    node_id VARCHAR(128),
    node_type VARCHAR(64) DEFAULT '',
    action VARCHAR(255),
    status VARCHAR(20),
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workflow_execution_workflow_id ON workflow_execution(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_execution_status ON workflow_execution(status);
CREATE INDEX IF NOT EXISTS idx_workflow_log_execution_id ON workflow_log(execution_id);
CREATE INDEX IF NOT EXISTS idx_workflow_log_node_id ON workflow_log(node_id);
