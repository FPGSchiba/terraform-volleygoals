# Terraform external data source build script for Windows
# Reads JSON from stdin, builds Go binary, outputs JSON to stdout

$InputJson = [Console]::In.ReadToEnd() | ConvertFrom-Json

$SrcDir    = $InputJson.src_dir
$BinaryOut = $InputJson.binary_out
$SrcHash   = $InputJson.src_hash

$HashFile    = "$BinaryOut.srchash"
$CurrentHash = if (Test-Path $HashFile) { (Get-Content $HashFile -Raw).Trim() } else { "" }

if ($CurrentHash -ne $SrcHash -or -not (Test-Path $BinaryOut)) {
    $OutDir = Split-Path $BinaryOut -Parent
    New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

    $PrevLocation = Get-Location
    Set-Location -Path $SrcDir

    go mod tidy 2>&1 | ForEach-Object { [Console]::Error.WriteLine($_) }

    $env:GOOS       = "linux"
    $env:GOARCH     = "amd64"
    $env:CGO_ENABLED = "0"

    go build -o $BinaryOut -mod=readonly -trimpath -ldflags="-s -w" . 2>&1 |
        ForEach-Object { [Console]::Error.WriteLine($_) }

    Set-Location -Path $PrevLocation

    Set-Content -Path $HashFile -Value $SrcHash -NoNewline
}

Write-Output "{`"hash`":`"$SrcHash`"}"
