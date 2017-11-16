test:
	go list ./... |grep -v vendor | xargs go test -race -v --cover

gometalinter:
	gometalinter.v1 --disable=gotype ./...