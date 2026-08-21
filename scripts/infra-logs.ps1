param(
  [string]$Project = "backend-interview-lab",
  [string]$Service,
  [int]$Tail = 100,
  [switch]$Follow
)

$ErrorActionPreference = "Stop"
$arguments = @("-p", $Project, "logs", "--tail=$Tail")
if ($Follow) { $arguments += "--follow" }
if ($Service) { $arguments += $Service }

& docker compose @arguments
if ($LASTEXITCODE -ne 0) {
  throw "docker compose logs failed with exit code $LASTEXITCODE"
}
