$ErrorActionPreference = 'Stop'

$packageName = 'fe'
$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$url = 'https://github.com/flushwhy/fe/releases/download/__TAG__/fe-windows-amd64.exe'

Get-ChocolateyWebFile -PackageName $packageName -FileFullPath (Join-Path $toolsDir 'fe.exe') -Url $url -Checksum '__SHA256__' -ChecksumType 'sha256'
Install-BinFile -Name 'fe' -Path (Join-Path $toolsDir 'fe.exe')
