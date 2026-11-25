export MAKEFLAGS="-j 16"

build_all: \
	cacheitems \
	createalbum \
	labelphotos \
	sortlocal

clean:
	find bin/ -type f -not -name .keep -delete
	rm -f out

cacheitems:
	go build -o bin/$@ cmd/$@/main.go

createalbum:
	go build -o bin/$@ cmd/$@/main.go

labelphotos:
	go build -o bin/$@ cmd/$@/main.go

sortlocal:
	go build -o bin/$@ cmd/$@/main.go

check: build_all
	go vet ./...