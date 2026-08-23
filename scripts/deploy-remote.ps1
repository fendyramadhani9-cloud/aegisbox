<#
.SYNOPSIS
    AegisBox Operator Remote Deployment Script for Windows.
.DESCRIPTION
    Transfers the release artifact to the Ubuntu VMware host (192.168.1.9)
    over SSH and triggers the atomic deployment & health validation process.
.EXAMPLE
    .\scripts\deploy-remote.ps1 -Host "192.168.1.9" -User "ubuntu"
#>

[CmdletBinding()]
param (
    [Parameter()]
    [string]$TargetHost = "192.168.1.9",

    [Parameter()]
    [string]$TargetUser = "ubuntu",

    [Parameter()]
    [string]$ArtifactPath = "dist/aegisbox-linux-amd64.tar.gz",

    [Parameter()]
    [int]$Port = 8080
)

$ErrorActionPreference = "Stop"

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  AegisBox Windows Operator Remote Deployment" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan

# 1. Locate release artifact
if (-not (Test-Path $ArtifactPath)) {
    if (Test-Path "aegisbox-linux-amd64.tar.gz") {
        $ArtifactPath = "aegisbox-linux-amd64.tar.gz"
    } else {
        Write-Host "Error: Release artifact '$ArtifactPath' not found." -ForegroundColor Red
        Write-Host "Please download the approved artifact from GitHub Actions or run 'make package' first." -ForegroundColor Yellow
        exit 1
    }
}

Write-Host "==> Release artifact: $ArtifactPath" -ForegroundColor Green
Write-Host "==> Target Host:      ${TargetUser}@${TargetHost}" -ForegroundColor Green

# 2. Transfer artifact via SCP
Write-Host "`n[Step 1/3] Uploading release package to ${TargetHost}:/tmp/aegisbox-linux-amd64.tar.gz..." -ForegroundColor Yellow
scp $ArtifactPath "${TargetUser}@${TargetHost}:/tmp/aegisbox-linux-amd64.tar.gz"
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: SCP file transfer failed." -ForegroundColor Red
    exit 1
}

# 3. Trigger remote deployment script
Write-Host "`n[Step 2/3] Executing atomic deployment on remote Ubuntu VM..." -ForegroundColor Yellow
$remoteCommand = "sudo mkdir -p /opt/aegisbox/releases /tmp/aegisbox-bootstrap && " +
                 "sudo tar -xzf /tmp/aegisbox-linux-amd64.tar.gz -C /tmp/aegisbox-bootstrap && " +
                 "sudo chmod +x /tmp/aegisbox-bootstrap/scripts/*.sh && " +
                 "sudo /tmp/aegisbox-bootstrap/scripts/deploy.sh /tmp/aegisbox-linux-amd64.tar.gz $Port && " +
                 "sudo rm -rf /tmp/aegisbox-bootstrap /tmp/aegisbox-linux-amd64.tar.gz"

ssh "${TargetUser}@${TargetHost}" $remoteCommand
if ($LASTEXITCODE -ne 0) {
    Write-Host "`nError: Remote deployment script reported a failure." -ForegroundColor Red
    exit 1
}

# 4. Verify remote health endpoint from Windows
Write-Host "`n[Step 3/3] Verifying remote health endpoint from Windows workstation..." -ForegroundColor Yellow
$healthUrl = "http://${TargetHost}:${Port}/health"

try {
    $response = Invoke-RestMethod -Uri $healthUrl -Method Get -TimeoutSec 5
    Write-Host "`n==================================================" -ForegroundColor Green
    Write-Host "  [SUCCESS] Deployment Live & Verified Healthy!" -ForegroundColor Green
    Write-Host "==================================================" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3 | Write-Host -ForegroundColor Cyan
} catch {
    Write-Host "Warning: Could not connect to $healthUrl directly from Windows: $_" -ForegroundColor Yellow
    Write-Host "Please ensure port $Port is allowed through the VM firewall if testing remotely." -ForegroundColor Yellow
}
