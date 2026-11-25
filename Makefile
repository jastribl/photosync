export MAKEFLAGS="-j 16"

build_all: \
	cacheitems \
	createalbum \
	findallmissingphotos \
	labelphotos \
	sortlocal \
	spacesaver

clean:
	find bin/ -type f -not -name .keep -delete
	rm -f out

cacheitems:
	go build -o bin/$@ cmd/$@/main.go

createalbum:
	go build -o bin/$@ cmd/$@/main.go

findallmissingphotos:
	go build -o bin/$@ cmd/$@/main.go

labelphotos:
	go build -o bin/$@ cmd/$@/main.go

sortlocal:
	go build -o bin/$@ cmd/$@/main.go

spacesaver:
	go build -o bin/$@ cmd/$@/main.go

check: build_all
	go vet ./...