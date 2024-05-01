test:
	go test -v ./...
cover:
	go test -coverprofile="cover.out" ./...
	go tool cover -html="cover.out"
covtotal:
	go test ./... -coverprofile cover.out
	go tool cover -func cover.out