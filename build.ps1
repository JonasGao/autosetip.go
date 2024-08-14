$pass = $args[0]
$file = $args[1]
$name = $args[2]
Write-Output "Using" $file
$file = (Get-Item $file).FullName
Push-Location .\encrypt
Write-Output "Building encrypt"
go build main.go
Write-Output "Encrypting"
./main.exe $pass $file
Remove-Item main.exe
Pop-Location
Write-Output "Config one"
Move-Item encrypt\config.yaml.enc one\static\
Push-Location .\one
Write-Output "Building one"
go build
Write-Output "Compressing"
Compress-Archive one.exe one.zip
Write-Output "Clear"
Remove-Item static\config.yaml.enc
New-Item -ItemType Directory -Force -Path ..\$name
Move-Item one.exe ..\$name\one.exe
Move-Item one.zip ..\$name\one.zip
Pop-Location
Write-Output "Finish"
