#staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
export PATH=$PATH:$(go env GOPATH)/bin && staticcheck ./...
https://staticcheck.dev/docs/checks/#S1030 или staticcheck -explain SA4009

#tests
cd /home/stanislav/go/romanov/ && go test ./... -v -coverpkg=./... -coverprofile=c.out && go tool cover -html=c.out -o cover.html && rm -f c.out && mv cover.html ./cmd/4_xml_search_http/

#
cd /home/stanislav/go/romanov/ && git checkout -b hw_4 && git add . && git commit -m 'homework 4' && git push

#
go install mvdan.cc/gofumpt@latest
gofumpt -l -w .

#
go generate ./data/data.go

