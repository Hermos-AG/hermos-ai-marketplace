<#
.SYNOPSIS
  Copies the current gpu-mcp release state from the source repo (HER-gpu-mcp) into the
  catalog's release copy plugins/gpu-mcp, syncs the versions and opens a pull request.

.EXAMPLE
  pwsh scripts/sync-gpu-mcp.ps1
  pwsh scripts/sync-gpu-mcp.ps1 -Source D:\DEV\HER\HER-MCP\gpu-mcp -CatalogVersion 1.5.0
#>
param(
  [string]$Source = (Join-Path (Split-Path $PSScriptRoot -Parent) '..\gpu-mcp'),
  [string]$CatalogVersion = ''
)

$ErrorActionPreference = 'Stop'
$catalogRoot = Split-Path $PSScriptRoot -Parent
$source = (Resolve-Path $Source).Path
$target = Join-Path $catalogRoot 'plugins\gpu-mcp'

Write-Host "source: $source"
Write-Host "target: $target"

$files = @(
  'main.go', 'proc_other.go', 'proc_windows.go', 'go.mod',
  'README.md', 'README_de.md', 'CHANGELOG.md', 'CHANGELOG_de.md',
  'RELEASE_NOTES.md', 'RELEASE_NOTES_de.md',
  'claude_desktop_config.example.json',
  '.claude-plugin\plugin.json'
)
foreach ($file in $files) {
  $from = Join-Path $source $file
  if (Test-Path $from) { Copy-Item $from (Join-Path $target $file) -Force; Write-Host "  copied $file" }
  else { Write-Warning "missing in source: $file" }
}
Copy-Item (Join-Path $source 'test') (Join-Path $target 'test') -Recurse -Force

foreach ($binary in @('gpu-mcp.exe', 'gpu-mcp-linux')) {
  $from = Join-Path $source $binary
  if (Test-Path $from) { Copy-Item $from (Join-Path $target $binary) -Force; Write-Host "  copied $binary" }
  else { Write-Warning "$binary not in the source working copy — build it, or let the workflow refresh-gpu-binaries do it" }
}

Push-Location $catalogRoot
try {
  python scripts\bump_versions.py $CatalogVersion
  python scripts\validate_catalog.py .
  if ($LASTEXITCODE -ne 0) { throw 'validate_catalog.py reported errors' }

  $version = (Get-Content 'plugins\gpu-mcp\.claude-plugin\plugin.json' -Raw | ConvertFrom-Json).version
  $catalog = (Get-Content '.claude-plugin\marketplace.json' -Raw | ConvertFrom-Json).version
  $branch = "sync/gpu-mcp-$version"

  git checkout -b $branch
  git add -A
  git commit -m "chore(gpu-mcp): Release-Kopie auf $version, Katalog $catalog"
  git push -u origin $branch
  gh pr create --fill --base main --head $branch
  Write-Host "pull request opened — merging it triggers the sync to Claude"
}
finally { Pop-Location }