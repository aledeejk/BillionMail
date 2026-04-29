# BillionMail Workflow System - Quick Start Script
# Purpose: Automate setup and launch of workflow system
# Usage: .\workflow-quickstart.ps1
# Run as Administrator for best results

param(
    [ValidateSet('setup', 'start', 'stop', 'clean')]
    [string]$Action = 'start'
)

# Colors for output
$Colors = @{
    Success = "Green"
    Error   = "Red"
    Warning = "Yellow"
    Info    = "Cyan"
}

function Write-Log {
    param([string]$Message, [string]$Color = "White")
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] $Message" -ForegroundColor $Color
}

function Test-PostgreSQL {
    Write-Log "Checking PostgreSQL connectivity..." "Info"
    try {
        $connection = New-Object System.Data.Odbc.OdbcConnection
        $connection.ConnectionString = "Driver={PostgreSQL Unicode(x64)};Server=127.0.0.1;Port=5432;Database=billionmail;Uid=billionmail;Pwd=NauF7ysRYyt9HTOiOn4JjIAL3QcRZnzj;"
        $connection.Open()
        $connection.Close()
        Write-Log "✓ PostgreSQL connection successful" "Success"
        return $true
    }
    catch {
        Write-Log "✗ PostgreSQL connection failed: $_" "Error"
        return $false
    }
}

function Setup-Database {
    Write-Log "Setting up database..." "Info"
    
    if (-not (Test-PostgreSQL)) {
        Write-Log "PostgreSQL is not accessible. Please ensure it's installed and running." "Error"
        Write-Log "Visit: https://www.postgresql.org/download/windows/" "Warning"
        return $false
    }

    # Apply migrations
    Write-Log "Applying database migrations..." "Info"
    $migrationFile = "migrations/20250427_create_workflow_tables.sql"
    
    if (Test-Path $migrationFile) {
        & psql -U billionmail -d billionmail -h 127.0.0.1 -f $migrationFile
        if ($LASTEXITCODE -eq 0) {
            Write-Log "✓ Migrations applied successfully" "Success"
            return $true
        }
        else {
            Write-Log "✗ Migration failed" "Error"
            return $false
        }
    }
    else {
        Write-Log "✗ Migration file not found: $migrationFile" "Error"
        return $false
    }
}

function Start-Services {
    Write-Log "Starting services..." "Info"

    # Check if directories exist
    if (-not (Test-Path "core")) {
        Write-Log "✗ core directory not found. Please run from project root." "Error"
        return $false
    }

    # Start backend
    Write-Log "Starting backend server..." "Info"
    $backendProcess = Start-Process -FilePath "powershell" -ArgumentList "-NoExit", "-Command", "cd core; go run main.go" -PassThru -WindowStyle Normal
    $backendPID = $backendProcess.Id
    Write-Log "✓ Backend started (PID: $backendPID)" "Success"

    Start-Sleep -Seconds 3

    # Start frontend
    Write-Log "Starting frontend development server..." "Info"
    $frontendProcess = Start-Process -FilePath "powershell" -ArgumentList "-NoExit", "-Command", "cd 'core/frontend'; npm run dev" -PassThru -WindowStyle Normal
    $frontendPID = $frontendProcess.Id
    Write-Log "✓ Frontend started (PID: $frontendPID)" "Success"

    Write-Log "" "Info"
    Write-Log "============================================" "Cyan"
    Write-Log "Services Started!" "Success"
    Write-Log "============================================" "Cyan"
    Write-Log "Backend:  https://localhost/api/workflow" "Info"
    Write-Log "Frontend: http://localhost:5173/automation" "Info"
    Write-Log "============================================" "Cyan"
    Write-Log "" "Info"
    Write-Log "Press any key to stop services..."
    [void][System.Console]::ReadKey($true)

    # Stop services
    Stop-Process -Id $backendPID -Force -ErrorAction SilentlyContinue
    Stop-Process -Id $frontendPID -Force -ErrorAction SilentlyContinue
    Write-Log "Services stopped" "Info"
}

function Verify-Setup {
    Write-Log "Verifying setup..." "Info"
    
    $checks = @{
        "Go"           = { go version | Out-Null }
        "Node.js"      = { node --version | Out-Null }
        "npm"          = { npm --version | Out-Null }
        "PostgreSQL"   = { psql --version | Out-Null }
    }

    foreach ($tool in $checks.Keys) {
        try {
            & $checks[$tool]
            Write-Log "✓ $tool is installed" "Success"
        }
        catch {
            Write-Log "✗ $tool is not installed or not in PATH" "Error"
        }
    }

    Write-Log "" "Info"
    if (Test-PostgreSQL) {
        Write-Log "✓ PostgreSQL database is accessible" "Success"
    }
    else {
        Write-Log "⚠ PostgreSQL database is not accessible" "Warning"
    }
}

function Clean-Cache {
    Write-Log "Cleaning cache files..." "Info"
    
    # Clean Go cache
    if (Test-Path "core/go.mod") {
        Write-Log "Cleaning Go cache..." "Info"
        Push-Location "core"
        go clean -cache
        go clean -modcache
        go mod tidy
        Pop-Location
        Write-Log "✓ Go cache cleaned" "Success"
    }

    # Clean npm cache
    if (Test-Path "core/frontend/package.json") {
        Write-Log "Cleaning npm cache..." "Info"
        Push-Location "core/frontend"
        npm cache clean --force
        if (Test-Path "node_modules") {
            Remove-Item -Path "node_modules" -Recurse -Force
        }
        if (Test-Path "package-lock.json") {
            Remove-Item -Path "package-lock.json" -Force
        }
        Pop-Location
        Write-Log "✓ npm cache cleaned" "Success"
    }

    Write-Log "Clean complete" "Info"
}

# Main execution
Write-Log "BillionMail Workflow System - Quick Start" "Cyan"
Write-Log "========================================" "Cyan"
Write-Log "" "Info"

switch ($Action) {
    'setup' {
        Verify-Setup
        if (Setup-Database) {
            Write-Log "✓ Setup completed successfully!" "Success"
        }
        else {
            Write-Log "✗ Setup failed. Please see errors above." "Error"
            exit 1
        }
    }

    'start' {
        Verify-Setup
        Start-Services
    }

    'stop' {
        Write-Log "Stopping services..." "Info"
        Stop-Process -Name "go" -Force -ErrorAction SilentlyContinue
        Stop-Process -Name "node" -Force -ErrorAction SilentlyContinue
        Write-Log "✓ Services stopped" "Success"
    }

    'clean' {
        Clean-Cache
    }

    default {
        Write-Log "Unknown action: $Action" "Error"
        Write-Log "Valid actions: setup, start, stop, clean" "Info"
    }
}

Write-Log "Done!" "Success"
