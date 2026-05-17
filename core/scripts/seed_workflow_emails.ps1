$ErrorActionPreference = "Stop"

$DB_HOST = "127.0.0.1"
$DB_PORT = "25432"
$DB_USER = "billionmail"
$DB_NAME = "billionmail"
$DB_PASS = "billionmail123"
$SQL_FILE = "$PSScriptRoot\seed_workflow_emails.sql"

$env:PGPASSWORD = $DB_PASS

Write-Host "=== BillionMail Workflow Email Seeder ===" -ForegroundColor Cyan
Write-Host "Connecting to ${DB_HOST}:${DB_PORT}/${DB_NAME} as $DB_USER"

$psql = $null
try { $psql = (Get-Command psql -ErrorAction Stop).Source } catch {}

if ($psql) {
    Write-Host "Using local psql: $psql" -ForegroundColor Green
    & $psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $SQL_FILE
} else {
    $container = docker ps --format "{{.Names}}" 2>$null | Where-Object { $_ -match "postgres" } | Select-Object -First 1
    if (-not $container) {
        Write-Error "Neither psql nor a running postgres Docker container found."
        exit 1
    }
    Write-Host "Using Docker container: $container" -ForegroundColor Green
    Get-Content $SQL_FILE -Raw | docker exec -i $container psql -U $DB_USER -d $DB_NAME
}

Write-Host ""
Write-Host "=== Done ===" -ForegroundColor Green
Write-Host "Verify results at http://localhost:8025 after running a workflow."
