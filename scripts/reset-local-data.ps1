param([switch]$ConfirmDataLoss, [string]$Project = "backend-interview-lab")
$ErrorActionPreference = "Stop"
if (-not $ConfirmDataLoss) { throw "Pass -ConfirmDataLoss to remove only this project's named volumes." }
docker compose -p $Project down --volumes --remove-orphans
