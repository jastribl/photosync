export MAKEFLAGS="-j 16"

build_all: \
	createalbum \
	labelphotos \
	sortlocal \
	suggestnames

clean:
	find bin/ -type f -not -name .keep -delete
	rm -f out

createalbum:
	go build -o bin/$@ cmd/$@/main.go

labelphotos:
	go build -o bin/$@ cmd/$@/main.go

sortlocal:
	go build -o bin/$@ cmd/$@/main.go

suggestnames:
	go build -o bin/$@ cmd/$@/main.go

check: build_all
	go vet ./...