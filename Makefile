#!/usr/bin/env gmake -f

GOOPTS=CGO_ENABLED=0
BUILDOPTS=-ldflags="-s -w" -a -gcflags=all=-l -trimpath -buildvcs=false
MYNAME=aleesa-misc-go
BINARY=${MYNAME}
UNIX_BINARY=${MYNAME}
WINDOWS_BINARY=${MYNAME}.exe
RMCMD=rm -rf

# Default target
.PHONY: help
help:
	@echo "aleesa-misc-go build system"
	@echo ""
	@echo "Available targets:"
	@echo "  all       Clean and build (default)"
	@echo "  build     Build binary only"
	@echo "  clean     Remove built artifacts"
	@echo "  test      Run unit tests"
	@echo "  upgrade   Update dependencies (get, tidy, vendor)"
	@echo ""

# На windows имя бинарника может зависеть не только от платформы, но и от выбранной цели, для linux-а суффикс .exe
# не нужен
ifeq ($(OS),Windows_NT)
ifdef GOOS
ifeq ($(GOOS),windows)
BINARY=${WINDOWS_BINARY}
else  # not ifeq ($(GOOS),windows)
BINARY=${MYNAME}
endif # ifeq ($(GOOS),windows)
else  # not ifdef GOOS
BINARY=${WINDOWS_BINARY}
endif # ifdef GOOS
ifeq ($(SHELL), sh.exe)
RMCMD=DEL /Q /F
endif
endif

# Явно определяем символ новой строки, чтобы избежать неоднозначности на windows
define IFS

endef

.PHONY: all build clean test upgrade help

all: clean build


build:
ifeq ($(OS),Windows_NT)
# Looks like on windows gnu make explicitly set SHELL to sh.exe, if it was not set.
ifeq ($(SHELL), sh.exe)
#       # Vanilla cmd.exe / powershell.
	SET "CGO_ENABLED=0"
	go build ${BUILDOPTS} -o ${BINARY} ./cmd/${MYNAME}
else ifeq (,$(findstring(Git/usr/bin/sh.exe, $(SHELL))))
#       # git-bash
	CGO_ENABLED=0 go build ${BUILDOPTS} -o ${BINARY} ./cmd/${MYNAME}
else  # not ifeq (,$(findstring(Git/usr/bin/sh.exe, $(SHELL))))
#       # Some other shell.
#       # TODO: handle it.
	$(info "-- Dunno how to handle this shell: ${SHELL}")
endif # ifeq (,$(findstring(Git/usr/bin/sh.exe, $(SHELL))))
else  # not  ($(OS),Windows_NT)
	CGO_ENABLED=0 go build ${BUILDOPTS} -o ${BINARY} ./cmd/${MYNAME}
endif # ifeq ($(OS),Windows_NT)


clean:
ifeq ($(OS),Windows_NT)
ifeq ($(SHELL),sh.exe)
#	# Vanilla cmd.exe / powershell.
	if exist ${WINDOWS_BINARY} ${RMCMD} ${WINDOWS_BINARY}
	if exist ${UNIX_BINARY} ${RMCMD} ${UNIX_BINARY}
else  # not ifeq ($(SHELL),sh.exe)
	${RMCMD} ./${WINDOWS_BINARY}
	${RMCMD} ./${UNIX_BINARY}
endif # ifeq ($(SHELL),sh.exe)
else  # not ifeq ($(OS),Windows_NT)
	${RMCMD} ./${BINARY}
endif


test:
	go test ./...


upgrade:
	go get -u ./...
	go mod tidy
	go mod vendor

# vim: set ft=make noet ai ts=4 sw=4 sts=4:
