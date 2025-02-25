#
# Copyright (c) 2020 HyTrust, Inc. All Rights Reserved.
#
THISDIR = $(shell pwd)

GOCMD = /usr/local/go/bin/go

PASMCLI_SRCDIR = $(THISDIR)
PARENTDIR = $(THISDIR)/..

BUILDDIR = pasmcli-build
WORKSPACE = $(PARENTDIR)/$(BUILDDIR)
WORKSPACE_SRCDIR = $(WORKSPACE)/src
WORKSPACE_BINDIR = $(WORKSPACE)/bin

PASMCLI_MAINGO = main.go

PASMCLI_LINUX = pasmcli
PASMCLI_WINDOWS = pasmcli.exe

all:
	@/usr/bin/echo "Clearing old Workspace if any at $(WORKSPACE).."
	@/usr/bin/rm -rf $(WORKSPACE)
	@/usr/bin/echo "Creating new pasmcli Workspace at $(WORKSPACE).."
	@/usr/bin/mkdir $(WORKSPACE)
	@/usr/bin/mkdir $(WORKSPACE_SRCDIR)
	@/usr/bin/cp -r $(PASMCLI_SRCDIR)/. $(WORKSPACE_SRCDIR)
	@/usr/bin/cp -r $(PARENTDIR)/getpasswd/ $(WORKSPACE_SRCDIR)/
	@/usr/bin/cp -r $(PARENTDIR)/cmd/. $(WORKSPACE_SRCDIR)/cmd/
	@cd $(WORKSPACE_SRCDIR) && $(GOCMD) mod init cli
	@cd $(WORKSPACE_SRCDIR) && $(GOCMD) mod tidy
	@/usr/bin/echo "Compiling pasmcli for Linux, Windows & Mac..."
	@cd $(WORKSPACE_SRCDIR) && (env GOOS=linux GOARCH=amd64 $(GOCMD) build -o $(WORKSPACE_BINDIR)/$(PASMCLI_LINUX) $(PASMCLI_MAINGO))
	@cd $(WORKSPACE_SRCDIR) && (env GOOS=windows GOARCH=amd64 $(GOCMD) build -o $(WORKSPACE_BINDIR)/$(PASMCLI_WINDOWS) $(PASMCLI_MAINGO))
	@/usr/bin/echo "Please find respective Linux & Windows pasmcli binaries, $(PASMCLI_LINUX) & $(PASMCLI_WINDOWS) at $(WORKSPACE_BINDIR)"

install:

clean:
	@/usr/bin/echo "Clearing old Workspace if any at $(WORKSPACE).."
	@/usr/bin/rm -rf $(WORKSPACE)
	@/usr/bin/echo "Clean up complete..."
