# Workflow System - Complete Setup with Port 25432

**Your Configuration:**
- Database: billionmail
- Database User: postgres
- Database Password: postgres
- Database Port: 25432
- Database Host: 127.0.0.1

This guide is customized for your .env settings.

---

## 📋 Step 1: Start PostgreSQL on Port 25432

### Option A: PostgreSQL in Docker (Recommended - Easiest)

If you have Docker Desktop installed:

```powershell
# Run PostgreSQL container on port 25432
docker run -d `
  --name billionmail-db `
  -e POSTGRES_USER=postgres `
  -e POSTGRES_PASSWORD=postgres `
  -e POSTGRES_DB=billionmail `
  -p 127.0.0.1:25432:5432 `
  postgres:16-alpine

# Verify container is running
docker ps | findstr billionmail-db
```

**Expected output:**
```
billionmail-db    postgres:16-alpine    Up X minutes    127.0.0.1:25432->5432/tcp
```

### Option B: Local PostgreSQL Installation (Port 25432)

If you have PostgreSQL installed locally, configure it for port 25432:

1. **Stop default PostgreSQL service:**
   ```powershell
   # Stop the default service
   Stop-Service postgresql-x64-16 -ErrorAction SilentlyContinue
   ```

2. **Edit PostgreSQL config:**
   - Find: `C:\Program Files\PostgreSQL\16\data\postgresql.conf`
   - Search for: `port = 5432`
   - Change to: `port = 25432`
   - Save and close

3. **Restart PostgreSQL:**
   ```powershell
   Start-Service postgresql-x64-16
   ```

---

## ✅ Step 2: Verify PostgreSQL Connection

Run these commands in PowerShell to verify PostgreSQL is accessible on port 25432:

### Test 1: Check if port is open
```powershell
# Check if port 25432 is listening
netstat -ano | findstr :25432
```

**Expected output shows port 25432 in LISTENING state**

### Test 2: Connect with psql

If you have PostgreSQL client tools installed:
```powershell
# Try to connect to PostgreSQL on port 25432
psql -U postgres -h 127.0.0.1 -p 25432 -d postgres -c "SELECT version();"

# When prompted for password, enter: postgres
```

**Expected output:**
```
PostgreSQL 16.0 on ... (SUCCESS)
```

### Test 3: Test from PowerShell without psql (Universal method)

```powershell
# Test connection using PowerShell TCP test
$TcpClient = New-Object System.Net.Sockets.TcpClient
try {
    $TcpClient.Connect("127.0.0.1", 25432)
    if ($TcpClient.Connected) {
        Write-Host "✓ PostgreSQL is accessible on port 25432" -ForegroundColor Green
    }
    $TcpClient.Close()
}
catch {
    Write-Host "✗ Cannot connect to PostgreSQL on port 25432" -ForegroundColor Red
    Write-Host "Error: $_"
}
```

**If connection fails:**
- Verify PostgreSQL is running (Docker or service)
- Check that port 25432 is not blocked by firewall
- Try the next section for troubleshooting

---

## 📦 Step 3: Verify billionmail Database Exists

Create the database and user (if not already created):

### Using psql (if available):

```powershell
# Connect to PostgreSQL and create database
psql -U postgres -h 127.0.0.1 -p 25432 -c "CREATE DATABASE billionmail OWNER postgres;" -v ON_ERROR_STOP=0

# Check if billionmail database exists
psql -U postgres -h 127.0.0.1 -p 25432 -l | findstr billionmail
```

**Expected output:**
```
 billionmail  | postgres | UTF8     | en_US.UTF-8 | en_US.UTF-8 |
```

### Using Docker (if running container):

```powershell
# Database is automatically created, just verify
docker exec billionmail-db psql -U postgres -d billionmail -c "SELECT 1;"
```

---

## 🚀 Step 4: Apply Database Migrations

Navigate to project directory and apply the migration:

```powershell
# Change to project directory
cd "d:\mechmat\computer networks\BillionMail"

# Apply migration using psql (method 1 - if psql is available)
psql -U postgres -h 127.0.0.1 -p 25432 -d billionmail -f migrations/20250427_create_workflow_tables.sql

# If psql is not available, use this PowerShell method (method 2)
# Skip to next section if method 1 worked
```

If you see:
```
CREATE TABLE
CREATE INDEX
```

✅ **Migration was successful!**

---

## 🔍 Step 5: Verify Workflow Tables Were Created

```powershell
# Method 1: Using psql (if available)
psql -U postgres -h 127.0.0.1 -p 25432 -d billionmail -c "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_name LIKE 'workflow%' ORDER BY table_name;"

# Method 2: Using Docker container
docker exec billionmail-db psql -U postgres -d billionmail -c "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_name LIKE 'workflow%' ORDER BY table_name;"
```

**Expected output shows 4 tables:**
```
              table_name              
--------------------------------------
 workflow
 workflow_execution
 workflow_execution_log
 workflow_version
(4 rows)
```

✅ **If you see these 4 tables, your database is ready!**

---

## 🔧 Step 6: Update Backend Configuration (Important!)

Your backend needs to know about the non-standard database port. Check the backend config:

```powershell
# Open config file
cd "d:\mechmat\computer networks\BillionMail\core"

# View current database configuration
cat manifest/config/config.yaml | findstr -A 5 "database:"
```

The config should read the .env values. Verify your `.env` file is in the correct location:

```powershell
# Verify .env exists in project root
Test-Path "d:\mechmat\computer networks\BillionMail\.env"

# If it doesn't exist or needs updating, copy from example
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
}

# Show current .env settings
Get-Content ".env" | findstr "DBNAME|DBUSER|DBPASS|SQL_PORT"
```

**Your .env should have:**
```ini
DBNAME=billionmail
DBUSER=postgres
DBPASS=postgres
SQL_PORT=127.0.0.1:25432
```

---

## 🎬 Step 7: Start the Backend

Open a **new PowerShell terminal** and run:

```powershell
# Navigate to core directory
cd "d:\mechmat\computer networks\BillionMail\core"

# Ensure dependencies are ready
go mod tidy

# Start the backend server
go run main.go
```

**Wait for this output:**
```
[INFO] 2025/04/28 ... listening on :80 (http)
[INFO] 2025/04/28 ... listening on :443 (https)
```

**Backend is listening on:**
- HTTP: `http://localhost:80` (or `http://localhost`)
- HTTPS: `https://localhost:443`

✅ **Keep this terminal open!** The backend is now running.

---

## 🎨 Step 8: Start the Frontend

Open **another new PowerShell terminal** and run:

```powershell
# Navigate to frontend directory
cd "d:\mechmat\computer networks\BillionMail\core\frontend"

# Install dependencies (if not already done)
npm install

# Start development server
npm run dev
```

**Wait for this output:**
```
VITE v5.0.0  ready in XXX ms

➜ Local: http://localhost:5173/
➜ press h to show help
```

**Frontend is accessible at:**
- `http://localhost:5173/automation`

✅ **Keep this terminal open!** The frontend is now running.

---

## 🌐 Step 9: Open the Application

Open your web browser and navigate to:

```
http://localhost:5173/automation
```

You should see:
- Empty workflow list
- "New Workflow" button
- Ready to create, edit, delete workflows

---

## ✨ Step 10: Test the Workflow System

### Test 1: Create a Workflow
1. Click **"New Workflow"** button
2. Enter:
   - Name: `Test Workflow`
   - Description: `This is a test`
3. Click **"Create"**
4. ✅ Workflow appears in the list

### Test 2: Edit Workflow
1. Click the workflow in the list
2. Click **"Edit"**
3. Change name to `Updated Workflow`
4. Click **"Save"**
5. ✅ Name updates in the list

### Test 3: API Call Verification
1. Open browser DevTools: Press `F12`
2. Go to **Network** tab
3. Create a new workflow
4. ✅ See `POST /api/workflow` request with 200 status

### Test 4: Check Database
```powershell
# Verify workflow was saved to database
psql -U postgres -h 127.0.0.1 -p 25432 -d billionmail -c "SELECT id, name, status, created_at FROM workflow ORDER BY id DESC LIMIT 1;"
```

**Should show your created workflow**

---

## 🐛 Troubleshooting

### Problem: "Connection refused" to port 25432

**Check if PostgreSQL is running:**
```powershell
# If using Docker
docker ps | findstr billionmail-db

# If using local PostgreSQL service
Get-Service postgresql-x64-16

# Check if port is actually open
netstat -ano | findstr 25432
```

**Solution:**
```powershell
# Start Docker container if not running
docker start billionmail-db

# OR start local PostgreSQL service
Start-Service postgresql-x64-16
```

### Problem: "Migration failed" or "psql not found"

**Install PostgreSQL client tools:**

Option A - Minimal approach (just use Docker):
```powershell
# Connect to database via Docker instead
docker exec billionmail-db psql -U postgres -d billionmail -f migrations/20250427_create_workflow_tables.sql

# Copy migration file into container
docker cp migrations/20250427_create_workflow_tables.sql billionmail-db:/tmp/
docker exec billionmail-db psql -U postgres -d billionmail -f /tmp/20250427_create_workflow_tables.sql
```

Option B - Install PostgreSQL client tools:
- Download from: https://www.postgresql.org/download/windows/
- Run installer and select only "Command Line Tools"
- Add to PATH: `C:\Program Files\PostgreSQL\16\bin`
- Restart PowerShell

### Problem: Backend won't start - "Cannot connect to database"

**Check database connection settings:**
```powershell
# View your .env file
Get-Content ".env" | Select-String "DB|SQL_PORT"

# Ensure these match your PostgreSQL setup:
# DBNAME=billionmail
# DBUSER=postgres
# DBPASS=postgres
# SQL_PORT=127.0.0.1:25432
```

**Verify connection manually:**
```powershell
# Try to connect
psql -U postgres -h 127.0.0.1 -p 25432 -d billionmail -c "SELECT 1;"
# Should respond with: (1 row)
```

### Problem: Frontend shows "Cannot connect to API"

**Check browser console (F12):**
- Look for CORS errors
- Check Network tab for API request status
- Verify backend is running: `http://localhost/api/workflow`

**Verify backend is accessible:**
```powershell
# From PowerShell
curl -SkI https://localhost/api/workflow

# Should get response (ignore SSL warning on localhost)
```

---

## 🔄 Fallback: Mock Mode for Demonstration

If you cannot get PostgreSQL working, switch to mock API for demonstration:

### Step 1: Backup Current API File
```powershell
cd "d:\mechmat\computer networks\BillionMail\core\frontend\src\api"

# Backup current file
Copy-Item workflow.ts workflow.ts.backup
```

### Step 2: Restore Mock Implementation

Option A - From git history:
```powershell
# Check git history for mock version
git log --oneline -- src/api/workflow.ts | head -5

# Restore previous version with mock data
git show <commit-hash>:src/api/workflow.ts > src/api/workflow.ts
```

Option B - Use inline mock directly:

Create a mock version by replacing `workflow.ts` with:

```typescript
// Mock implementation for demonstration
import type {
  Workflow,
  WorkflowExecution,
  WorkflowStatistics,
  WorkflowVersion,
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  GetWorkflowStatsRequest,
} from '@/types/workflow'

// Sample mock data
const MOCK_WORKFLOWS: Workflow[] = [
  {
    id: '1',
    name: 'Newsletter Campaign',
    description: 'Send weekly newsletter to subscribers',
    isActive: true,
    version: 3,
    createdAt: '2025-04-27T10:00:00Z',
    updatedAt: '2025-04-25T15:30:00Z',
  },
]

export const workflowApi = {
  async getWorkflows() {
    await new Promise(resolve => setTimeout(resolve, 500))
    return MOCK_WORKFLOWS
  },

  async getWorkflow(id: string) {
    await new Promise(resolve => setTimeout(resolve, 300))
    return MOCK_WORKFLOWS.find(w => w.id === id) || null
  },

  async createWorkflow(data: CreateWorkflowRequest) {
    await new Promise(resolve => setTimeout(resolve, 800))
    const newWorkflow: Workflow = {
      id: String(MOCK_WORKFLOWS.length + 1),
      name: data.name,
      description: data.description,
      isActive: false,
      version: 1,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    MOCK_WORKFLOWS.push(newWorkflow)
    return newWorkflow
  },

  async updateWorkflow(data: UpdateWorkflowRequest) {
    await new Promise(resolve => setTimeout(resolve, 800))
    const workflow = MOCK_WORKFLOWS.find(w => w.id === data.id)
    if (!workflow) throw new Error('Workflow not found')
    if (data.name) workflow.name = data.name
    if (data.description) workflow.description = data.description
    workflow.updatedAt = new Date().toISOString()
    return workflow
  },

  async deleteWorkflow(id: string) {
    await new Promise(resolve => setTimeout(resolve, 500))
    const index = MOCK_WORKFLOWS.findIndex(w => w.id === id)
    if (index !== -1) MOCK_WORKFLOWS.splice(index, 1)
  },

  async duplicateWorkflow(id: string) {
    await new Promise(resolve => setTimeout(resolve, 800))
    const original = MOCK_WORKFLOWS.find(w => w.id === id)
    if (!original) throw new Error('Workflow not found')
    const duplicate: Workflow = {
      ...original,
      id: String(MOCK_WORKFLOWS.length + 1),
      name: `${original.name} (Copy)`,
      isActive: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    MOCK_WORKFLOWS.push(duplicate)
    return duplicate
  },

  async toggleWorkflow(id: string, isActive: boolean) {
    await new Promise(resolve => setTimeout(resolve, 500))
    const workflow = MOCK_WORKFLOWS.find(w => w.id === id)
    if (!workflow) throw new Error('Workflow not found')
    workflow.isActive = isActive
    return workflow
  },

  async getWorkflowStats(request: GetWorkflowStatsRequest) {
    await new Promise(resolve => setTimeout(resolve, 400))
    return {
      totalExecutions: 42,
      successfulExecutions: 40,
      failedExecutions: 2,
      averageDuration: 15000,
    }
  },

  async getWorkflowVersions(id: string) {
    await new Promise(resolve => setTimeout(resolve, 400))
    return [
      { version: 3, createdAt: '2025-04-25T15:30:00Z', status: 'active' },
      { version: 2, createdAt: '2025-04-20T10:00:00Z', status: 'inactive' },
    ]
  },

  async getWorkflowExecutions(workflowId: string) {
    await new Promise(resolve => setTimeout(resolve, 400))
    return []
  },

  async executeWorkflow(id: string) {
    await new Promise(resolve => setTimeout(resolve, 600))
    return {
      id: `exec-${Date.now()}`,
      workflowId: id,
      status: 'running' as const,
      startedAt: new Date().toISOString(),
    }
  },
}
```

### Step 3: Restart Frontend

```powershell
# The frontend will auto-reload with mock data
# If not, manually stop and restart:

# Press Ctrl+C in frontend terminal to stop
# Then restart:
npm run dev
```

✅ **Frontend now works with mock data - no backend needed!**

---

## ✅ Quick Verification Checklist

- [ ] PostgreSQL running on port 25432
- [ ] Can connect: `psql -U postgres -h 127.0.0.1 -p 25432`
- [ ] Database `billionmail` exists
- [ ] Migration applied: 4 workflow tables exist
- [ ] Backend running: `go run main.go` (listening on :80)
- [ ] Frontend running: `npm run dev` (localhost:5173)
- [ ] Can access: `http://localhost:5173/automation`
- [ ] Can create workflows
- [ ] Browser DevTools shows API calls to `/api/workflow`

---

## 🎯 Summary of Exact Commands

```powershell
# === TERMINAL 1: PostgreSQL (Docker) ===
docker run -d --name billionmail-db `
  -e POSTGRES_USER=postgres `
  -e POSTGRES_PASSWORD=postgres `
  -e POSTGRES_DB=billionmail `
  -p 127.0.0.1:25432:5432 postgres:16-alpine

# === TERMINAL 2: Verify Connection ===
psql -U postgres -h 127.0.0.1 -p 25432 -d billionmail -c "SELECT 1;"

# === TERMINAL 2: Apply Migration ===
cd "d:\mechmat\computer networks\BillionMail"
psql -U postgres -h 127.0.0.1 -p 25432 -d billionmail -f migrations/20250427_create_workflow_tables.sql

# === TERMINAL 3: Backend ===
cd "d:\mechmat\computer networks\BillionMail\core"
go mod tidy
go run main.go

# === TERMINAL 4: Frontend ===
cd "d:\mechmat\computer networks\BillionMail\core\frontend"
npm install
npm run dev

# === BROWSER ===
http://localhost:5173/automation
```

---

**You're all set! Follow these steps in order and you'll have the complete workflow system running.** ✅
