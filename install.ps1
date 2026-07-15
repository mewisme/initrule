# initrule installer for Windows (PowerShell).
#
# irm https://raw.githubusercontent.com/mewisme/initrule/main/install.ps1 | iex
#
# Environment:
#   INITRULE_VERSION      release tag (default: latest)
#   INITRULE_INSTALL_DIR  install location (default: %LOCALAPPDATA%\initrule)

$ErrorActionPreference = 'Stop'
$repo = 'mewisme/initrule'
$installDir = if ($env:INITRULE_INSTALL_DIR) { $env:INITRULE_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'initrule' }

$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }

$version = $env:INITRULE_VERSION
if (-not $version) {
  $version = (Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest").tag_name
}
if (-not $version) { throw "initrule: could not resolve latest version; set INITRULE_VERSION." }
if ($version -notmatch '^v') { $version = "v$version" }
$ver = $version.TrimStart('v')

$url = "https://github.com/$repo/releases/download/$version/initrule_${ver}_windows_${arch}.zip"
Write-Host "Installing initrule $version (windows/$arch)..."

$tmp = Join-Path $env:TEMP ("initrule-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Force -Path $tmp | Out-Null
$zip = Join-Path $tmp 'initrule.zip'
Invoke-WebRequest -Uri $url -OutFile $zip

$dest = Join-Path $installDir 'current'
if (Test-Path $dest) { Remove-Item -Recurse -Force $dest }
New-Item -ItemType Directory -Force -Path $dest | Out-Null
Expand-Archive -Path $zip -DestinationPath $dest -Force
Remove-Item -Recurse -Force $tmp

$exe = Join-Path $dest 'initrule.exe'
if (-not (Test-Path $exe)) { throw "initrule: initrule.exe missing from archive." }

$defaultInstall = Join-Path $env:LOCALAPPDATA 'initrule'
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
Write-Host "Run: initrule --help"
