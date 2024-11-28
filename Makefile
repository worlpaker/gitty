test:
	go test -v ./...
race:
	go test -race ./...
cover:
	go test -coverprofile="cover.out" ./...
	go tool cover -html="cover.out"
covtotal:
	go test ./... -coverprofile cover.out
	go tool cover -func cover.out
covatomic:
	go test -race -coverprofile="cover.out" -covermode=atomic ./...
	go tool cover -html="cover.out"
deps:
	go get -u all
	go mod tidy
# Ensure https://github.com/icholy/gomajor is installed.
major:
	gomajor get all
	go mod tidy