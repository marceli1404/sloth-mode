$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
New-Item -ItemType Directory -Force (Join-Path $root 'build'), (Join-Path $root 'installer/payload') | Out-Null
Push-Location (Join-Path $root 'connector')
try {
  $env:CGO_ENABLED='0'; $env:GOOS='windows'; $env:GOARCH='amd64'
  go test ./...
  if ($LASTEXITCODE -ne 0) { throw 'Connector tests failed' }
  go build -trimpath -ldflags='-H windowsgui -s -w' -o '../installer/payload/SlothMode-Connector.exe' .
  if ($LASTEXITCODE -ne 0) { throw 'Connector build failed' }
} finally { Pop-Location }
Push-Location (Join-Path $root 'installer')
try {
  go test ./...
  if ($LASTEXITCODE -ne 0) { throw 'Installer tests failed' }
  go build -trimpath -ldflags='-H windowsgui -s -w' -o '../build/SlothMode-Setup.exe' .
  if ($LASTEXITCODE -ne 0) { throw 'Installer build failed' }
} finally { Pop-Location }
Write-Host 'Built build/SlothMode-Setup.exe (unsigned)'
