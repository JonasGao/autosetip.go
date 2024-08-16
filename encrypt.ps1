$pass = $args[0]
$file = $args[1]
Write-Output "Using" $file
$file = (Get-Item $file).FullName
Push-Location .\encrypt
Write-Output "Building encrypt"
go build main.go
Write-Output "Encrypting"
./main.exe $pass $file
Remove-Item main.exe
Pop-Location
Write-Output "Finish"
