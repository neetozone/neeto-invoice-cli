$ErrorActionPreference = "Stop"

$BaseUrl = "https://neeto-downloads.s3.amazonaws.com/cli/NeetoInvoice/latest"
$InstallDir = if ($env:NEETOINVOICE_INSTALL_DIR) { $env:NEETOINVOICE_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\neetoinvoice" }

$Arch = if ([Environment]::Is64BitOperatingSystem) {
  if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
  Write-Error "Unsupported architecture"
  exit 1
}

$Archive = "neetoinvoice_windows_${Arch}.zip"
$Url = "${BaseUrl}/${Archive}"

$TmpDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }

try {
  $ZipPath = Join-Path $TmpDir $Archive

  Write-Host "Downloading NeetoInvoice CLI for windows/${Arch}..."
  Invoke-WebRequest -Uri $Url -OutFile $ZipPath

  Write-Host "Verifying checksum..."
  $SumsPath = Join-Path $TmpDir "SHA256SUMS"
  Invoke-WebRequest -Uri "${BaseUrl}/SHA256SUMS" -OutFile $SumsPath

  $Expected = $null
  foreach ($Line in Get-Content $SumsPath) {
    $Parts = $Line.Trim() -split "\s+", 2
    if ($Parts.Count -eq 2 -and $Parts[1].TrimStart("*") -eq $Archive) { $Expected = $Parts[0] }
  }
  if (-not $Expected) {
    throw "No published checksum for ${Archive}. Aborting."
  }

  $Actual = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash
  if ($Expected -ne $Actual) {
    throw "Checksum mismatch for ${Archive}. Aborting.`n  expected: ${Expected}`n  actual:   ${Actual}"
  }

  Write-Host "Extracting..."
  Expand-Archive -Path $ZipPath -DestinationPath $TmpDir -Force

  Write-Host "Installing to ${InstallDir}..."
  New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
  Copy-Item (Join-Path $TmpDir "neetoinvoice.exe") -Destination (Join-Path $InstallDir "neetoinvoice.exe") -Force

  $Key = [Microsoft.Win32.Registry]::CurrentUser.CreateSubKey("Environment")
  try {
    $Kind = try { $Key.GetValueKind("Path") } catch { [Microsoft.Win32.RegistryValueKind]::ExpandString }
    $UserPath = [string]$Key.GetValue("Path", "", [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
    $Entries = @($UserPath -split ";" | Where-Object { $_ -ne "" })
    if ($Entries -notcontains $InstallDir) {
      $Key.SetValue("Path", (($Entries + $InstallDir) -join ";"), $Kind)
      Write-Host "Added ${InstallDir} to user PATH."
    }
  } finally {
    $Key.Dispose()
  }
} finally {
  Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}

Write-Host "NeetoInvoice CLI installed successfully. Restart your terminal and run 'neetoinvoice --help' to get started."
