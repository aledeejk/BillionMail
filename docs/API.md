# Workflow Automation API

Base path: `/api`  
All endpoints return `Content-Type: application/json`.  
Authentication: Bearer JWT in `Authorization` header (where required by server config).

---

## Workflows

### List workflows
```
GET /api/workflow
```
Query params: `page`, `page_size`, `keyword`, `status` (-1 = all, 0 = inactive, 1 = active)

Response:
```json
{"code":0,"data":{"total":5,"list":[{"id":"1","name":"Welcome","status":1,"version":3}]}}
```

---

### Get workflow
```
GET /api/workflow/{id}
```

---

### Create workflow
```
POST /api/workflow
```
Body:
```json
{"name":"Welcome series","description":"...","is_active":false}
```

---

### Update workflow
```
PUT /api/workflow/{id}
```
Body: same as create.

---

### Delete workflow
```
DELETE /api/workflow/{id}
```

---

### Duplicate workflow
```
POST /api/workflow/{id}/duplicate
```
Returns the new workflow with name `"<original> Copy"`.

---

### Toggle active/inactive
```
POST /api/workflow/{id}/toggle
```
Flips `status` between `0` and `1`.

---

## Execution

### Execute workflow (async when Redis is available)
```
POST /api/workflow/{id}/execute
```
Body:
```json
{
  "contact_email": "user@example.com",
  "contact_id": 42,
  "trigger": "manual",
  "idempotency_key": "order-123",
  "input_data": {"plan": "pro"}
}
```
Response when queued:
```json
{"code":0,"data":{"status":0}}
```
Response when synchronous:
```json
{"code":0,"data":{"id":"17","workflow_id":"1","status":2,"started_at":1716000000}}
```

---

### Execution history
```
GET /api/workflow/{id}/executions?page=1&page_size=20
```

---

### Execution log (detailed per-node)
```
GET /api/workflow/{id}/execution-log?contact=user@example.com&status=failed
```
Query params: `contact` (email filter), `status` (`success`|`failed`)

---

### Export execution log as CSV
```
GET /api/workflow/{id}/execution-log/export
GET /api/workflow/{id}/export
```
Returns `text/csv` file.

---

## Visual Editor

### Get editor data (nodes + connections)
```
GET /api/workflow/{id}/editor
```
Response:
```json
{
  "workflow": {"id":"1","name":"Welcome","version":3},
  "nodes": [{"id":"uuid","type":"trigger","config":{},"position_x":100,"position_y":200}],
  "connections": [{"id":"uuid","source":"node-a","target":"node-b","condition":""}]
}
```

---

### Save editor data
```
PUT /api/workflow/{id}/editor
```
Body:
```json
{
  "nodes": [...],
  "connections": [...]
}
```
Saves a version snapshot atomically.

---

## Version Management

### List versions
```
GET /api/workflow/{id}/versions
```
Response: array of `{version, created_at, status}` (status=1 means current).

---

### Rollback to version
```
POST /api/workflow/{id}/rollback/{version}
```
Deletes current nodes/connections and restores them from the snapshot.  
Returns fresh editor data.

---

### Delete version
```
DELETE /api/workflow/{id}/versions/{version}
```
Cannot delete the currently active version.

---

## Analytics

### Statistics summary
```
GET /api/workflow/{id}/stats
```
Response:
```json
{"total_executions":100,"success_count":95,"failure_count":5,"average_duration":1.2,"last_run_at":1716000000}
```

---

### Per-node statistics
```
GET /api/workflow/{id}/node-stats
```
Response: array of `{node_id, node_type, entered_count, completed_count, conversion_rate, avg_duration_ms}`.

---

### Email engagement report
```
GET /api/workflow/{id}/report
```
Response:
```json
{"sent":200,"unique_opens":80,"unique_clicks":30,"unsubscribes":2}
```

---

## Webhook (external trigger)

### Trigger workflow from external system
```
POST /api/workflow/trigger/{id}
POST /api/webhook/trigger/{workflow_id}
```
Body: arbitrary JSON payload — all keys are available in node `inputData`.  
```json
{"contact_email":"user@example.com","contact_id":"42","plan":"pro"}
```
Response (queued):
```json
{"status":"queued"}
```
Response (synchronous):
```json
{"status":"started","execution_id":"17"}
```

---

## Node types and configuration

| Node type | Required config keys | Notes |
|---|---|---|
| `trigger` | — | Entry point, always first |
| `send-email` | `templateId` (int), `sender` (email, optional), `subject` (optional) | Sends real email via SMTP |
| `delay` | `duration` (int), `unit` (`minutes`\|`hours`\|`days`) | Logged only; actual sleep not yet implemented |
| `condition` | `logic` (`AND`\|`OR`), `conditions[]` | Branches: `true` edge / `false` edge |
| `action` | `actionType` (`add-tag`\|`remove-tag`\|`move-to-group`), `value`, `groupId` | Modifies contact data |
| `split` | `branchA` (%), `branchB` (%) | Deterministic per-contact A/B split |

---

## Split node — determinism

The `split` node distributes contacts deterministically:  
`hash(contact_email + node_id) mod 10000 / 100` vs `branchA` threshold.  
The same contact always goes to the same branch for a given node.

## Idempotency

Pass `idempotency_key` in the execute request body.  
If an execution with the same `(workflow_id, idempotency_key)` already exists,  
the existing record is returned without re-executing.
