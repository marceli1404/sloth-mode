$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$installer = Join-Path $root 'build/SlothMode-Setup.exe'
if (-not (Test-Path -LiteralPath $installer)) { throw 'Build the Windows installer first with scripts/build-windows.ps1' }
$output = Join-Path $root 'build/preview'
New-Item -ItemType Directory -Force -Path $output | Out-Null
Copy-Item -LiteralPath $installer -Destination (Join-Path $output 'SlothMode-Setup.exe')
Copy-Item -LiteralPath (Join-Path $root 'LICENSE') -Destination $output
@'
Sloth Mode Windows preview

This installer is unsigned. Models and llama.cpp are downloaded separately only
after approval in the app. An account is optional; browser pairing is required.

Run SlothMode-Setup.exe, choose Install & open, then use Setup & Pairing.
Source, known limitations and setup instructions:
https://github.com/marceli1404/sloth-mode

Automated builds do not establish hardware validation or real Gemma inference.
See docs/RELEASE_VALIDATION.md in the source repository before claiming support.
'@ | Set-Content -LiteralPath (Join-Path $output 'README.txt') -Encoding utf8
$hash = (Get-FileHash -LiteralPath $installer -Algorithm SHA256).Hash.ToLowerInvariant()
"$hash  SlothMode-Setup.exe" | Set-Content -LiteralPath (Join-Path $output 'SHA256SUMS.txt') -Encoding ascii
Write-Host "Prepared unsigned preview and checksum at $output"
