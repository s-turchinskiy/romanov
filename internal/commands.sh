#staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
export PATH=$PATH:$(go env GOPATH)/bin && staticcheck ./...
https://staticcheck.dev/docs/checks/#S1030 или staticcheck -explain SA4009