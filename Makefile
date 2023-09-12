# Define the output binary name
binary_name = rootfsbuilder

# Define where to install the binary
prefix = /usr/local
bindir = $(prefix)/bin

mandir = $(prefix)/share/man
man1dir = $(mandir)/man1
man_page_src = rootfsbuilder.1

all: build

build:
	go build -o $(binary_name)

test:
	go test ./... -v

install:
	mkdir -p $(bindir)
	install -m 755 $(binary_name) $(bindir)/$(binary_name)
	mkdir -p $(man1dir)
	install -m 644 $(man_page_src) $(man1dir)/$(man_page_src)

uninstall:
	rm -f $(bindir)/$(binary_name)
	rm -f $(man1dir)/$(man_page_src)

clean:
	rm -f $(binary_name)

.PHONY: all build test install uninstall clean
