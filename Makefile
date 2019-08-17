serve:
	CGO_ENABLED=0 GOPATH=$(PWD) go run firefight.go

mac:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 GOPATH=$(PWD) go build -o firefight_darwin firefight.go

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOPATH=$(PWD) go build -o firefight_linux firefight.go

windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOPATH=$(PWD) go build -o firefight_windows firefight.go
