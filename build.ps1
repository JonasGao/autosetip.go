$pass = $args[0]
$file = $args[1]
$name = $args[2]
Write-Output "Using" $file
$file = (Get-Item $file).FullName
Push-Location .\encrypt
go build main.go
./main.exe $pass $file
Remove-Item main.exe
Pop-Location
Move-Item encrypt\config.yaml.enc one\static\
Push-Location .\one
go build
Compress-Archive one.exe one.zip
Remove-Item static\config.yaml.enc
Remove-Item one.exe
Move-Item one.zip ..\$name.zip
Pop-Location
