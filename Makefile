APPNAME=dev.flowingspdg.vmix.sdPlugin

MAKEFILE_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BUILDDIR = $(MAKEFILE_DIR)build

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

.DEFAULT_GOAL := build

test:
	cd $(MAKEFILE_DIR)$(APPNAME)/code && go run $(GOFLAGS) main.go handlers.go settings.go vmix.go -port 12345 -pluginUUID 213 -registerEvent test -info "{\"application\":{\"language\":\"en\",\"platform\":\"mac\",\"version\":\"4.1.0\"},\"plugin\":{\"version\":\"1.1\"},\"devicePixelRatio\":2,\"devices\":[{\"id\":\"55F16B35884A859CCE4FFA1FC8D3DE5B\",\"name\":\"Device Name\",\"size\":{\"columns\":5,\"rows\":3},\"type\":0},{\"id\":\"B8F04425B95855CF417199BCB97CD2BB\",\"name\":\"Another Device\",\"size\":{\"columns\":3,\"rows\":2},\"type\":1}]}"

prepare:
	@$(MKDIR) $(BUILDDIR)
	@$(RM) $(BUILDDIR)/*

build: prepare
	cd $(MAKEFILE_DIR)$(APPNAME)/code && GOOS=windows GOARCH=amd64 go build -o $(BUILDDIR)/vmix_go.exe .
	$(CP) $(MAKEFILE_DIR)$(APPNAME)/*.json $(BUILDDIR)
	$(CP) $(MAKEFILE_DIR)$(APPNAME)/inspector $(BUILDDIR)
	$(CP) $(MAKEFILE_DIR)$(APPNAME)/images $(BUILDDIR)