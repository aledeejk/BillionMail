# BillionMail Workflow System - Quick Reference

This document provides a quick reference for the Workflow CRUD and automation system implementation.

## 📋 What Was Implemented

✅ **Backend Workflow Service**
- Complete CRUD operations (Create, Read, Update, Delete)
- Workflow versioning system
- Workflow execution tracking
- Statistics and analytics
- Duplicate workflow functionality
- Toggle workflow active/inactive status

✅ **Frontend Workflow Management**
- Vue 3 component for workflow list
- Create workflow form
- Edit workflow dialog
- Workflow details panel
- Statistics view
- Real HTTP client integration with backend API

✅ **Database Schema**
- `workflow` - Main workflow definitions
- `workflow_version` - Version history
- `workflow_execution` - Execution records
- `workflow_execution_log` - Detailed execution logs

✅ **API Endpoints**
```
POST   /api/workflow              - Create workflow
GET    /api/workflow              - List workflows
GET    /api/workflow/{id}         - Get workflow details
PUT    /api/workflow/{id}         - Update workflow
DELETE /api/workflow/{id}         - Delete workflow
POST   /api/workflow/{id}/duplicate - Duplicate workflow
POST   /api/workflow/{id}/toggle  - Toggle active status
GET    /api/workflow/{id}/stats   - Get statistics
```

## 🚀 Quick Start

### Prerequisites
- Go 1.26.2
- Node.js v24.15.0
- PostgreSQL 12+

### Setup (1-2 minutes)

1. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Create database:**
   ```bash
   psql -U postgres -h 127.0.0.1 -c "CREATE USER billionmail WITH PASSWORD 'NauF7ysRYyt9HTOiOn4JjIAL3QcRZnzj';"
   psql -U postgres -h 127.0.0.1 -c "CREATE DATABASE billionmail OWNER billionmail;"
   ```

3. **Apply migrations:**
   ```bash
   psql -U billionmail -d billionmail -h 127.0.0.1 -f migrations/20250427_create_workflow_tables.sql
   ```

### Start Services (separate terminals)

**Terminal 1 - Backend:**
```bash
cd core
go mod tidy
go run main.go
# Waits for: listening on :80 and :443
```

**Terminal 2 - Frontend:**
```bash
cd core/frontend
npm install
npm run dev
# Waits for: ➜ Local: http://localhost:5173/
```

### Access the Application

- **Frontend:** http://localhost:5173/automation
- **API Base:** http://localhost/api/workflow
- **Swagger Docs:** http://localhost/swagger

## 📁 Project Structure

```
core/
├── api/workflow/                 # API definitions and routes
│   ├── v1/workflow.go           # Request/response types
│   └── routes.go                # Route registration
│
├── internal/
│   ├── controller/workflow/      # HTTP handlers
│   │   ├── controller.go
│   │   └── controller_test.go
│   ├── service/workflow/         # Business logic
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── service_stats.go
│   │   └── service_test.go
│   └── cmd/cmd.go              # Main entry point (routes registration)
│
├── frontend/src/
│   ├── api/workflow.ts         # Frontend HTTP client
│   ├── types/workflow.ts       # TypeScript interfaces
│   └── views/automation/       # UI components
│
└── migrations/
    └── 20250427_create_workflow_tables.sql  # Database schema
```

## 🧪 Testing

### Manual Testing

1. Open http://localhost:5173/automation
2. Create a workflow
3. Verify it appears in the list
4. Test CRUD operations:
   - Edit workflow
   - Duplicate workflow
   - Toggle status
   - Delete workflow
5. Check browser console (F12) for API calls

### API Testing with curl

```bash
# Create workflow
curl -X POST http://localhost/api/workflow \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"name":"Test","description":"Test workflow","is_active":true}'

# List workflows
curl http://localhost/api/workflow \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get specific workflow
curl http://localhost/api/workflow/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🔧 Configuration Files

### Backend Configuration
- `core/manifest/config/config.yaml` - Server settings
- `.env` - Environment variables (create from `.env.example`)

### Frontend Configuration
- `core/frontend/.env.development` - Dev environment
- `core/frontend/.env.production` - Production build
- `core/frontend/rsbuild.config.ts` - Build configuration

## 📝 Database Details

### Workflow Table
```sql
id           - Primary key
name         - Workflow name
description  - Workflow description
status       - 0: inactive, 1: active
version      - Current version number
created_at   - Unix timestamp
updated_at   - Unix timestamp
```

### Workflow Version Table
```sql
id              - Primary key
workflow_id     - Reference to workflow
version         - Version number
content         - JSON definition
created_at      - Creation timestamp
created_by      - Creator username
```

### Workflow Execution Table
```sql
id              - Primary key
workflow_id     - Reference to workflow
status          - pending, running, completed, failed
started_at      - Execution start time
completed_at    - Execution end time
error_message   - Error details if failed
```

## 🐛 Troubleshooting

### Backend won't start
```bash
# Check PostgreSQL is running
Get-Service postgresql-x64-16

# Check logs
cat core/logs/error-*.log

# Check port is available
netstat -ano | findstr :80
```

### Frontend won't connect to backend
```bash
# Verify backend is running
curl http://localhost/api/workflow

# Check browser console for CORS errors
# Check API client URL in core/frontend/src/api/index.ts
```

### Database migration failed
```bash
# Check database exists
psql -U postgres -l | findstr billionmail

# Re-apply migration
psql -U billionmail -d billionmail -f migrations/20250427_create_workflow_tables.sql
```

## 📚 Documentation

- **Detailed Setup:** See [WORKFLOW_SETUP.md](WORKFLOW_SETUP.md)
- **API Examples:** See [docs/examples/workflow_examples.md](docs/examples/workflow_examples.md)
- **Backend Tests:** See [core/internal/service/workflow/service_test.go](core/internal/service/workflow/service_test.go)

## 🔗 API Response Format

All API responses follow this format:

```json
{
  "code": 0,
  "msg": "success",
  "success": true,
  "data": {
    // Response data here
  }
}
```

### Create/Update Response
```json
{
  "data": {
    "id": "1",
    "name": "Workflow Name",
    "description": "Description",
    "is_active": true,
    "version": 1,
    "created_at": "2025-04-27T12:34:56Z",
    "updated_at": "2025-04-27T12:34:56Z"
  }
}
```

### List Response
```json
{
  "data": {
    "list": [
      // Array of workflows
    ],
    "total": 10
  }
}
```

## 📦 Dependencies

### Backend
- GoFrame v2 - Web framework
- PostgreSQL driver - Database access
- Redis - Caching

### Frontend
- Vue 3 - UI framework
- TypeScript - Type safety
- Axios - HTTP client
- UnoCSS - Styling

## ✨ Features

- **CRUD Operations** - Create, Read, Update, Delete workflows
- **Versioning** - Track workflow changes across versions
- **Execution History** - Monitor workflow executions
- **Statistics** - Track execution metrics
- **Duplication** - Clone workflows with one click
- **Status Toggle** - Activate/deactivate workflows
- **Real-time UI** - Frontend connected to live backend API

## 🎯 Next Steps

1. Complete the quick start above
2. Run manual tests against the API
3. Build out workflow definition editor
4. Add execution schedule configuration
5. Implement email trigger integration
6. Add advanced filtering and search

## 📞 Support

For issues or questions:
1. Check [WORKFLOW_SETUP.md](WORKFLOW_SETUP.md) troubleshooting section
2. Review error logs in `core/logs/`
3. Check browser DevTools (F12) for frontend errors
4. Verify database schema with `psql` commands

---

**Status:** ✅ Ready for development and testing
**Last Updated:** 2025-04-28
