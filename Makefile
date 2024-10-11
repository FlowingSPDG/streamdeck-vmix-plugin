APPNAME=dev.flowingspdg.vmix.sdPlugin

GOFLAGS=--race

MAKEFILE_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BUILDDIR = $(MAKEFILE_DIR)$(APPNAME)
SRCDIR = $(MAKEFILE_DIR)Source
PIDIR = $(MAKEFILE_DIR)Source/pi
RELEASEDIR = Release

# Replacing "RM" command for Windows PowerShell.
RM = rm -rf
ifeq ($(OS),Windows_NT)
    RM = Remove-Item -Recurse -Force
endif

# Replacing "MKDIR" command for Windows PowerShell.
MKDIR = mkdir -p
ifeq ($(OS),Windows_NT)
    MKDIR = New-Item -Force -ItemType Directory
endif

# Replacing "CP" command for Windows PowerShell.
CP = cp -R
ifeq ($(OS),Windows_NT)
	CP = powershell -Command Copy-Item -Recurse -Force
endif

# Replacing Distribute command for Windows PowerShell.
DISTRIBUTION_TOOL = ./DistributionTool.exe
ifeq  ($(shell uname),Darwin)
	DISTRIBUTION_TOOL = ./DistributionTool
endif

.DEFAULT_GOAL := build

test:
	cd $(SRCDIR)/code && go run $(GOFLAGS) main.go handlers.go pi.go vmix.go -port 12345 -pluginUUID 213 -registerEvent test -info "{\"application\":{\"language\":\"en\",\"platform\":\"mac\",\"version\":\"4.1.0\"},\"plugin\":{\"version\":\"1.1\"},\"devicePixelRatio\":2,\"devices\":[{\"id\":\"55F16B35884A859CCE4FFA1FC8D3DE5B\",\"name\":\"Device Name\",\"size\":{\"columns\":5,\"rows\":3},\"type\":0},{\"id\":\"B8F04425B95855CF417199BCB97CD2BB\",\"name\":\"Another Device\",\"size\":{\"columns\":3,\"rows\":2},\"type\":1}]}"

vet:
	cd $(SRCDIR)/code && go vet

prepare:
	@$(MKDIR) $(BUILDDIR)
	@$(RM) $(BUILDDIR)/*
	@$(RM) ./$(RELEASEDIR)/*

build-pi:
	cd $(SRCDIR)/pi && npm run build

build-server:
	cd $(SRCDIR)/code/cmd && GOOS=windows GOARCH=amd64 go build -o $(BUILDDIR)/vmix_go.exe .
	cd $(SRCDIR)/code/cmd && GOOS=darwin GOARCH=amd64 go build -o $(BUILDDIR)/vmix_go .

build: prepare build-pi build-server
	$(CP) $(PIDIR)/dist $(BUILDDIR)/inspector
	$(CP) $(SRCDIR)/manifest.json $(BUILDDIR)
	$(CP) $(SRCDIR)/images $(BUILDDIR)

distribute: build
	@$(MKDIR) $(RELEASEDIR)
	$(DISTRIBUTION_TOOL) -b -i $(APPNAME) -o $(RELEASEDIR)