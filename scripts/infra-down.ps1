<#
.SYNOPSIS
  Stop local interview-lab containers without deleting data by default.
#>
param(
  [string]$Project = "backend-interview-lab",
  [switch]$Volumes
)

$ErrorActionPreference = "Stop"
$arguments = @("down", "--remove-orphans")
if ($Volumes) {
  Write-Warning "Removing the '$Project' named volumes deletes local PostgreSQL data."
  $arguments += "--volumes"
}

& docker compose -p $Project @arguments
if ($LASTEXITCODE -ne 0) {
  throw "docker compose down failed with exit code $LASTEXITCODE"
}
