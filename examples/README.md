# Maylng API Examples

This directory contains example scripts and code for using the Maylng API.

## Files

- `api_example.sh` - Bash script demonstrating basic API usage
- `api_example.ps1` - PowerShell script for Windows users
- `postman_collection.json` - Postman collection with API examples

## Usage

### Bash Script (Linux/macOS)

```bash
chmod +x api_example.sh
./api_example.sh
```

### PowerShell Script (Windows)

```powershell
.\api_example.ps1
```

## Prerequisites

- `curl` command available
- `jq` for JSON parsing (optional but recommended)
- Running Maylng API server

## Environment Variables

You can set these environment variables to customize the examples:

- `MAYLNG_API_BASE` - API base URL (default: <http://localhost:8080/v1>)
- `MAYLNG_API_KEY` - Your API key (if you already have one)

## API Endpoints Covered

1. Account creation
2. Account details retrieval
3. Email address creation
4. Email address listing
5. Email sending
6. Email status checking
7. Email listing
