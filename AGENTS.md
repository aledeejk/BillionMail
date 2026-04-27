# AGENTS.md - Instructions for AI agents

## Development environment
- OS: Windows 11
- Go version: go1.26.2
- Node.js version: v24.15.0
- Database: PostgreSQL
- Package Manager: go mod / npm

## Build and launch a project
```bash
# Backend
cd core
go mod download
go run main.go

# Frontend
cd core/frontend
npm install
npm run dev

# Complete build (without Makefile)
cd core
go build -o bin/billionmail.exe ./...
cd core/frontend
npm run build
```