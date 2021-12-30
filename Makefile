export MAKEFLAGS="-j 8"

build_all: \
	cacheitems \
	createalbum \
	drive2photos \
	findallmissinglocal \
	findallmissingphotos \
	labelphotos \
	spacesaver

clean:
	find bin/ -type f -not -name .keep -delete
	rm -f out

cacheitems:
	go build -o bin/$@ cmd/$@/main.go

createalbum:
	go build -o bin/$@ cmd/$@/main.go

drive2photos:
	go build -o bin/$@ cmd/$@/main.go

findallmissinglocal:
	go build -o bin/$@ cmd/$@/main.go

findallmissingphotos:
	go build -o bin/$@ cmd/$@/main.go

labelphotos:
	go build -o bin/$@ cmd/$@/main.go

spacesaver:
	go build -o bin/$@ cmd/$@/main.go

check: findallmissingphotos findallmissinglocal
	rm -f out && touch out
	./bin/findallmissinglocal >> out
	./bin/findallmissingphotos ~/Pictures/All\ Pictures/Parents\ Grad\ Trip\ 2019/ >> out
	./bin/findallmissingphotos ~/Pictures/All\ Pictures/San\ Francisco\ 2018/ >> out
	./bin/findallmissingphotos ~/Pictures/All\ Pictures/Seattle\ 2018/ >> out
	./bin/findallmissingphotos ~/Pictures/All\ Pictures/Seattle\ 2019/ >> out
	./bin/findallmissingphotos ~/Pictures/All\ Pictures/Seattle\ 2020/ >> out
	./bin/findallmissingphotos ~/Pictures/All\ Pictures/Seattle\ 2021/ >> out
	./bin/findallmissingphotos ~/Pictures/All\ Pictures/Winter\ 2019\ Term/ >> out
	cat out
	
