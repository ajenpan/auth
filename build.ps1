go env -w GOOS="linux"
go build -o auth ./cmd

go env -w GOOS="windows"
go build -o auth.exe ./cmd
