<#
.SYNOPSIS
  Start the local infrastructure for the interview labs.

.EXAMPLE
  .\scripts\infra-up.ps1 -Profile foundation
  .\scripts\infra-up.ps1 -Profile eventing
  .\scripts\infra-up.ps1 -Profile full -Build
#>
param(
  [ValidateSet("foundation", "eventing", "cache", "full")]
  [string]$Profile = "foundation",
  [string]$Project = "backend-interview-lab",
  [switch]$Build
)

$ErrorActionPreference = "Stop"

function Invoke-Compose([string[]]$Arguments) {
  & docker compose -p $Project @Arguments
  if ($LASTEXITCODE -ne 0) {
    throw "docker compose failed with exit code $LASTEXITCODE"
  }
}

switch ($Profile) {
  "foundation" {
    # PostgreSQL is the durable base. Migration is a one-shot service and
    # exits 0 after applying platform/migrations/001_init.sql.
    Invoke-Compose @("up", "-d", "postgres", "migrate")
  }
  "eventing" {
    Invoke-Compose @("--profile", "eventing", "up", "-d")
  }
  "cache" {
    Invoke-Compose @("--profile", "cache", "up", "-d")
  }
  "full" {
    $composeArgs = @("--profile", "full", "up")
    if ($Build) { $composeArgs += "--build" }
    $composeArgs += "-d"
    Invoke-Compose $composeArgs
  }
}

Write-Host "Infrastructure profile '$Profile' is started for project '$Project'."
Write-Host "Run .\scripts\infra-status.ps1 to inspect health and ports."
