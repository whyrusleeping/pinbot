include $(GOROOT)/src/Make.inc

TARG=github.com/mudler/sendfd
GOFILES=\
	sendfd.go\

GOFILES_linux=\
	cmsg_$(GOOS)_$(GOARCH).go\

GOFILES_arm=\
	cmsg_$(GOOS)_$(GOARCH).go\

GOFILES_darwin=\
	cmsg_$(GOOS).go\

GOFILES_freebsd=\
	cmsg_$(GOOS).go\

GOFILES+=$(GOFILES_$(GOOS))

include $(GOROOT)/src/Make.pkg
