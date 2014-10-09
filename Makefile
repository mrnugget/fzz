DESTDIR    ?= /usr/local
MAN_DIR    := ./man
MAN_PAGE   := $(MAN_DIR)/fzz.1
EXECUTABLE := fzz

.PHONY: all

all: man $(EXECUTABLE)

clean:
	rm $(EXECUTABLE)

install:
	install -d $(DESTDIR)/bin
	install -d $(DESTDIR)/share/man/man1
	install $(EXECUTABLE) $(DESTDIR)/bin
	cp -R $(MAN_PAGE) $(DESTDIR)/share/man/man1

#
# fzz executable
#

$(EXECUTABLE): *.go
	go fmt
	go build -o $(EXECUTABLE)

#
# tests
#

test:
	go test -v

#
# man page
#

$(MAN_DIR)/%.1: $(MAN_DIR)/%.markdown
	@which md2man-roff >/dev/null || (echo "md2man missing: gem install md2man"; exit 1)
	md2man-roff $< > $@

man: $(MAN_PAGE)
