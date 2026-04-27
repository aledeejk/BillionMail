-- Migration: create workflow tables

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Create workflow table
CREATE TABLE workflow (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    description text,
    is_active boolean NOT NULL DEFAULT true,
    version int NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

-- Create workflow_node table
CREATE TABLE workflow_node (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id uuid NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    node_type varchar(50) NOT NULL CHECK (node_type IN ('trigger', 'email', 'delay', 'condition', 'action', 'split')),
    config jsonb NOT NULL DEFAULT '{}'::jsonb,
    position_x double precision NOT NULL DEFAULT 0,
    position_y double precision NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Create workflow_connection table
CREATE TABLE workflow_connection (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id uuid NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    from_node_id uuid NOT NULL REFERENCES workflow_node(id) ON DELETE CASCADE,
    to_node_id uuid NOT NULL REFERENCES workflow_node(id) ON DELETE CASCADE,
    condition jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Create workflow_execution table
CREATE TABLE workflow_execution (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id uuid NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    contact_id uuid NOT NULL,
    current_node_id uuid REFERENCES workflow_node(id),
    status varchar(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'suspended', 'failed')),
    started_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Create workflow_log table
CREATE TABLE workflow_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id uuid NOT NULL REFERENCES workflow_execution(id) ON DELETE CASCADE,
    node_id uuid NOT NULL REFERENCES workflow_node(id),
    action varchar(255),
    status varchar(20) NOT NULL CHECK (status IN ('successful', 'failed', 'pending')),
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_workflow_deleted_at ON workflow(deleted_at);
CREATE INDEX idx_workflow_node_workflow_id ON workflow_node(workflow_id);
CREATE INDEX idx_workflow_connection_workflow_id ON workflow_connection(workflow_id);
CREATE INDEX idx_workflow_execution_workflow_id ON workflow_execution(workflow_id);
CREATE INDEX idx_workflow_execution_contact_id ON workflow_execution(contact_id);
CREATE INDEX idx_workflow_execution_status ON workflow_execution(status);
CREATE INDEX idx_workflow_log_execution_id ON workflow_log(execution_id);
CREATE INDEX idx_workflow_log_created_at ON workflow_log(created_at);
