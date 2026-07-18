# agentrule installer for Windows (PowerShell).
#
# irm https://raw.githubusercontent.com/mewisme/agentrule/main/install.ps1 | iex
#
# Environment:
#   AGENTRULE_VERSION      release tag (default: latest)
#   AGENTRULE_INSTALL_DIR  install location (default: %LOCALAPPDATA%\agentrule)

$ErrorActionPreference = 'Stop'
$repo = 'mewisme/agentrule'
$installDir = if ($env:AGENTRULE_INSTALL_DIR) { $env:AGENTRULE_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'agentrule' }

$arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
  'Arm64' { 'arm64' }
  'X64' { 'amd64' }
  'X86' { '386' }
  default { throw "agentrule: unsupported architecture '$_'." }
}

$version = $env:AGENTRULE_VERSION
if (-not $version) {
  $version = (Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest").tag_name
}
if (-not $version) { throw "agentrule: could not resolve latest version; set AGENTRULE_VERSION." }
if ($version -notmatch '^v') { $version = "v$version" }
$ver = $version.TrimStart('v')

$url = "https://github.com/$repo/releases/download/$version/agentrule_${ver}_windows_${arch}.zip"
Write-Host "Installing agentrule $version (windows/$arch)..."

$tmp = Join-Path $env:TEMP ("agentrule-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Force -Path $tmp | Out-Null
$zip = Join-Path $tmp 'agentrule.zip'
Invoke-WebRequest -Uri $url -OutFile $zip

$dest = Join-Path $installDir 'current'
if (Test-Path $dest) { Remove-Item -Recurse -Force $dest }
New-Item -ItemType Directory -Force -Path $dest | Out-Null
Expand-Archive -Path $zip -DestinationPath $dest -Force
Remove-Item -Recurse -Force $tmp

$exe = Join-Path $dest 'agentrule.exe'
if (-not (Test-Path $exe)) { throw "agentrule: agentrule.exe missing from archive." }

$defaultInstall = Join-Path $env:LOCALAPPDATA 'agentrule'
# Only mutate user PATH for the default install location.
if ($installDir -eq $defaultInstall) {
  $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
  if (($userPath -split ';') -notcontains $dest) {
    [Environment]::SetEnvironmentVariable('Path', "$dest;$userPath", 'User')
    $env:Path = "$dest;$env:Path"
    Write-Host "Added $dest to your PATH (restart your terminal if needed)."
  }
}

Write-Host "Installed to $dest"
Write-Host "Run: agentrule --help"
