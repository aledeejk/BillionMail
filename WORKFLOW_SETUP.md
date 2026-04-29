# Workflow System Setup and Testing Guide

This guide provides step-by-step instructions to get the complete Workflow CRUD and automation system running in BillionMail.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Database Setup](#database-setup)
3. [Database Migration](#database-migration)
4. [Backend Setup](#backend-setup)
5. [Frontend Setup](#frontend-setup)
6. [Testing the Complete Flow](#testing-the-complete-flow)
7. [Fallback Options](#fallback-options)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

**System Requirements:**
- Windows 11 (or Windows 10)
- Go 1.26.2
- Node.js v24.15.0
- PostgreSQL 12+ (local installation or Docker)
- Terminal: PowerShell, CMD, or Git Bash

**Project Location:**
```
d:\mechmat\computer networks\BillionMail
```

---

## Database Setup

### Option 1: Local PostgreSQL Installation (Recommended)

#### Step 1: Install PostgreSQL

1. Download PostgreSQL from: https://www.postgresql.org/download/windows/
2. Run the installer and follow the wizard:
   - Choose installation directory (e.g., `C:\Program Files\PostgreSQL\16`)
   - Set superuser password (remember this!)
   - Port: 5432
   - Locale: [your preference]
3. Complete installation and add PostgreSQL to PATH

#### Step 2: Start PostgreSQL Service

PowerShell (as Administrator):
```powershell
Start-Service postgresql-x64-16
# Verify service is running
Get-Service postgresql-x64-16
```

Or manually through Services:
- Press `Win + R`
- Type `services.msc`
- Find "postgresql-x64-16" and set to "Running"

#### Step 3: Create Database and User

Open PostgreSQL command line:
```powershell
# Connect to PostgreSQL as superuser
psql -U postgres -h localhost
```

In psql terminal, execute:
```sql
-- Create billionmail user
CREATE USER billionmail WITH PASSWORD 'NauF7ysRYyt9HTOiOn4JjIAL3QcRZnzj';

-- Create billionmail database
CREATE DATABASE billionmail OWNER billionmail;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE billionmail TO billionmail;

-- Connect to billionmail database
\c billionmail

-- Grant schema privileges
GRANT ALL ON SCHEMA public TO billionmail;

-- Exit psql
\q
```

**Verify connection:**
```powershell
psql -U billionmail -d billionmail -h localhost -c "SELECT version();"
# Password: NauF7ysRYyt9HTOiOn4JjIAL3QcRZnzj
```

### Option 2: Docker PostgreSQL

If you have Docker Desktop installed:

```powershell
# Run PostgreSQL container
docker run -d `
  --name billionmail-postgres `
  -e POSTGRES_USER=billionmail `
  -e POSTGRES_PASSWORD=NauF7ysRYyt9HTOiOn4JjIAL3QcRZnzj `
  -e POSTGRES_DB=billionmail `
  -p 5432:127.0.0.1:5432 `
  postgres:16-alpine

# Verify it's running
docker ps | findstr billionmail-postgres
```

---

## Database Migration

### Step 1: Create .env File

Create file: `d:\mechmat\computer networks\BillionMail\.env`

```ini
# Database Configuration
DBNAME=billionmail
DBUSER=billionmail
DBPASS=NauF7ysRYyt9HTOiOn4JjIAL3QcRZnzj

# Database Host (adjust if not local)
DB_HOST=127.0.0.1
DB_PORT=5432

# System Configuration
ADMIN_USERNAME=billion
ADMIN_PASSWORD=billion
SafePath=billion

# Redis Configuration
REDISPASS=zKLnZQr3riFpcS2lEy3MOtfncztaCGKp
REDIS_PORT=127.0.0.1:26379

# Mail Configuration
BILLIONMAIL_HOSTNAME=mail.example.com

# Time Zone
TZ=Etc/UTC
```

### Step 2: Apply Database Migrations

The workflow tables are defined in:
```
migrations/20250427_create_workflow_tables.sql
```

Apply the migration using psql:

```powershell
# Navigate to project directory
cd "d:\mechmat\computer networks\BillionMail"

# Apply the migration
psql -U billionmail -d billionmail -h 127.0.0.1 -f migrations/20250427_create_workflow_tables.sql

# Verify tables were created
psql -U billionmail -d billionmail -h 127.0.0.1 -c "\dt workflow*"
```

Expected output should show:
```
              List of relations
 Schema |            Name            | Type  | Owner
--------+----------------------------+-------+-------
 public | workflow                   | table | billionmail
 public | workflow_execution         | table | billionmail
 public | workflow_execution_log     | table | billionmail
 public | workflow_version           | table | billionmail
```

### Step 3: Verify Tables and Schemas

```powershell
# Check workflow table structure
psql -U billionmail -d billionmail -h 127.0.0.1 -c "\d workflow"

# Check all created tables
psql -U billionmail -d billionmail -h 127.0.0.1 -c "
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name LIKE 'workflow%';"
```

---

## Backend Setup

### Step 1: Verify Workflow Route Registration

The workflow controller is already registered in:
- File: `core/internal/cmd/cmd.go`
- Import: `"billionmail-core/internal/controller/workflow"`
- Binding: Added to `group.Bind()` at line 265

Routes available at:
- `POST /api/workflow` - Create workflow
- `GET /api/workflow` - List workflows
- `GET /api/workflow/{id}` - Get workflow
- `PUT /api/workflow/{id}` - Update workflow
- `DELETE /api/workflow/{id}` - Delete workflow
- `POST /api/workflow/{id}/duplicate` - Duplicate workflow
- `POST /api/workflow/{id}/toggle` - Toggle active status
- `GET /api/workflow/{id}/stats` - Get workflow statistics

### Step 2: Install Go Dependencies

```powershell
cd "d:\mechmat\computer networks\BillionMail\core"

# Download dependencies
go mod download

# Tidy up dependencies
go mod tidy
```

### Step 3: Run the Backend

```powershell
# From core directory
cd "d:\mechmat\computer networks\BillionMail\core"

# Run the backend server
go run main.go

# Expected output:
# [INFO] 2025/04/28 12:34:56 starting server...
# [INFO] 2025/04/28 12:34:56 server started, listening on :80
```

**Backend is ready when you see:**
```
listening on :80 (http)
listening on :443 (https)
```

Press `Ctrl+C` to stop. Keep it running in a separate terminal for testing.

---

## Frontend Setup

### Step 1: Install Node Dependencies

```powershell
cd "d:\mechmat\computer networks\BillionMail\core\frontend"

# Install dependencies
npm install
```

### Step 2: Run Frontend Development Server

```powershell
# From frontend directory
cd "d:\mechmat\computer networks\BillionMail\core\frontend"

# Start development server
npm run dev

# Expected output:
#   VITE v5.0.0  ready in X ms
#   ➜ Local: http://localhost:5173/
```

**Frontend is ready when you see:**
```
➜ Local: http://localhost:5173/
```

Keep this terminal running in a separate window.

---

## Testing the Complete Flow

### Prerequisites for Testing
- ✅ PostgreSQL is running
- ✅ Backend is running on http://localhost/api (or https://localhost/api with self-signed cert)
- ✅ Frontend is running on http://localhost:5173

### Test Flow

#### Test 1: Open the Application

1. Open browser and navigate to:
   ```
   http://localhost:5173/automation
   ```

2. You should see the Workflow automation page with an empty list

#### Test 2: Create a Workflow

1. Click **"New Workflow"** button
2. Fill in:
   - **Name:** "Test Newsletter Campaign"
   - **Description:** "Send newsletter to subscribers every Monday"
3. Click **"Create"**
4. Verify:
   - Workflow appears in the list
   - Status shows as "Inactive"
   - Created timestamp is current time

#### Test 3: View Workflow Details

1. Click on the created workflow
2. Verify details panel shows:
   - Name, description, version
   - Created/Updated timestamps
   - Current status

#### Test 4: Duplicate Workflow

1. Click **"Duplicate"** button on a workflow
2. Verify:
   - New workflow created with name "Test Newsletter Campaign Copy"
   - Same description as original
   - New ID and timestamp
   - Status is "Inactive"

#### Test 5: Toggle Workflow Status

1. Click **"Toggle"** button on a workflow
2. Verify:
   - Status changes from "Inactive" → "Active" or vice versa
   - UI updates immediately

#### Test 6: View Statistics

1. Click **"Stats"** on an active workflow
2. Verify:
   - Statistics panel displays (may show 0 if no executions)
   - Metrics shown: Total Executions, Successful, Failed, Last Run

#### Test 7: Edit Workflow

1. Click **"Edit"** button
2. Change:
   - Name to "Updated Newsletter"
   - Description to "Updated description"
3. Click **"Save"**
4. Verify changes are reflected in the list

#### Test 8: Delete Workflow

1. Click **"Delete"** button
2. Confirm deletion
3. Verify:
   - Workflow removed from list
   - No errors in console

#### Test 9: API Response Verification

Open browser DevTools (F12):
1. Go to Network tab
2. Perform workflow operations (Create, Update, Delete)
3. Verify:
   - Requests go to `/api/workflow` endpoints
   - Responses contain proper JSON data
   - Status codes: 200 for success, appropriate error codes for failures

---

## Fallback Options

### Option 1: Mock Frontend Mode (No Backend Required)

If the backend is not available, revert the frontend API to mock mode:

**File:** `core/frontend/src/api/workflow.ts`

Replace the entire file with the mock implementation:
```typescript
// Restore from git history or use the mock version that includes:
// - MOCK_WORKFLOWS array with sample data
// - Mock API methods returning sample data
// - setTimeout delays to simulate network latency
```

Then restart frontend:
```powershell
npm run dev
```

The workflow UI will work with local mock data.

### Option 2: Simple UI Demonstration

For instructor demonstration without a running system:

1. **Screenshots:**
   - Take screenshots of the workflow interface
   - Document the UI components and interactions

2. **Static HTML:**
   - Build a static HTML prototype
   - Show workflow list, create form, details view

3. **Swagger API Documentation:**
   ```
   http://localhost/swagger
   ```
   Shows all API endpoints and models

---

## Troubleshooting

### Database Connection Issues

**Error:** `connection refused` or `could not connect to server`

Solutions:
```powershell
# Check PostgreSQL is running
Get-Service postgresql-x64-16

# Verify connection details
psql -U billionmail -d billionmail -h 127.0.0.1 -W

# Check database exists
psql -U postgres -h 127.0.0.1 -l | findstr billionmail

# Recreate database if needed
psql -U postgres -h 127.0.0.1 -f path/to/init.sql
```

### Backend Won't Start

**Error:** `port already in use` or `connection to database failed`

Solutions:
```powershell
# Kill process using port 80 (Windows)
netstat -ano | findstr :80
taskkill /PID <PID> /F

# Or use different port - modify config.yaml
# Set server.address to ":8080" and test

# Check logs
cat logs/error-*.log
```

### Frontend API Errors

**Error:** `404 Not Found` or `CORS error`

Solutions:
```powershell
# Ensure backend is running
curl -i http://localhost/api/workflow

# Check CORS headers in browser console
# Verify API client uses correct prefix: '/api'

# Clear browser cache
# Ctrl + Shift + Delete in most browsers
```

### Workflow Tables Not Created

**Error:** `relation "workflow" does not exist`

Solutions:
```powershell
# Re-apply migration
psql -U billionmail -d billionmail -h 127.0.0.1 -f migrations/20250427_create_workflow_tables.sql

# Check if tables exist
psql -U billionmail -d billionmail -h 127.0.0.1 -c "SELECT * FROM workflow LIMIT 1;"
```

### Go Module Issues

**Error:** `module not found` or `package not found`

Solutions:
```powershell
cd core

# Clear Go cache
go clean -cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy

# Re-run
go run main.go
```

---

## Summary of Key Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/workflow` | Create new workflow |
| GET | `/api/workflow` | List all workflows |
| GET | `/api/workflow/{id}` | Get workflow details |
| PUT | `/api/workflow/{id}` | Update workflow |
| DELETE | `/api/workflow/{id}` | Delete workflow |
| POST | `/api/workflow/{id}/duplicate` | Duplicate workflow |
| POST | `/api/workflow/{id}/toggle` | Toggle active status |
| GET | `/api/workflow/{id}/stats` | Get workflow statistics |

---

## Files Changed/Created

### New Files
- `migrations/20250427_create_workflow_tables.sql` - Database schema
- `.env` - Environment configuration

### Modified Files
- `core/internal/cmd/cmd.go` - Added workflow controller registration

### Existing Files (Already Complete)
- `core/internal/controller/workflow/controller.go` - Workflow endpoints
- `core/internal/service/workflow/` - Business logic
- `core/frontend/src/api/workflow.ts` - Frontend API client
- `core/frontend/src/types/workflow.ts` - TypeScript types
- `core/frontend/src/views/automation/` - UI components

---

## Next Steps

1. ✅ Complete this setup guide
2. ✅ Test all workflow operations
3. Consider implementing:
   - Workflow versioning UI
   - Execution history viewing
   - Advanced filtering and search
   - Workflow templates
   - Email campaign integration

---

## Questions or Issues?

Refer to:
- Backend logs: `core/logs/`
- Frontend console: Browser DevTools (F12)
- Database logs: PostgreSQL logs directory
- API documentation: Auto-generated at `/swagger`

