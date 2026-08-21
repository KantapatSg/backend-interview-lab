param([string]$Project = "backend-interview-lab")
$ErrorActionPreference = "Stop"
docker compose -p $Project ps -a
if ($LASTEXITCODE -ne 0) {
  throw "docker compose ps failed with exit code $LASTEXITCODE"
}
