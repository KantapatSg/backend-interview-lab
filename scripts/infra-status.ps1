param([string]$Project = "backend-interview-lab")
$ErrorActionPreference = "Stop"
docker compose -p $Project ps
