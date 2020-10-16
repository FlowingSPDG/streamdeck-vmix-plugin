GO = go
APPNAME=dev.flowingspdg.streamdeck.vmix.sdPlugin

BIN_NAME = streamdeck-vmix-plugin
BUILDDIR = ./build
INSTALLDIR = $(APPDATA)\Elgato\StreamDeck\Plugins\$(APPNAME)
# Replacing "INSTALLDIR" directory for Mac.
ifeq  ($(shell uname),Darwin)
    INSTALLDIR = ~/Library/Application\ Support/com.elgato.StreamDeck/Plugins/$(APPNAME)/
endif

# Replacing "RM" command for Windows PowerShell.
RM = rm -rf
ifeq ($(OS),Windows_NT)
    RM = Remove-Item -Recurse -Force
endif

# Replacing "MKDIR" command for Windows PowerShell.
MKDIR = mkdir -p
ifeq ($(OS),Windows_NT)
    MKDIR = New-Item -ItemType Directory
endif

# Replacing "CP" command for Windows PowerShell.
CP = cp -R
ifeq ($(OS),Windows_NT)
	CP = powershell -Command Copy-Item -Recurse -Force
endif

# Replacing "TMP" directory for Mac.
ifeq  ($(shell uname),Darwin)
    TMP = /tmp
endif


.PHONY: test install build logs

prepare:
	@$(RM) $(INSTALLDIR)
	@$(MKDIR) $(INSTALLDIR)
	@$(MKDIR) $(BUILDDIR)
	@$(RM) $(BUILDDIR)/*

build: prepare
	gox --osarch "windows/amd64" --output $(BUILDDIR)/${BIN_NAME}_{{.OS}}_{{.Arch}}
	$(CP) manifest.json $(BUILDDIR)
	$(CP) inspector $(BUILDDIR)
	$(CP) sdpi.css $(BUILDDIR)

test:
	$(GO) run $(GOFLAGS) main.go -port 12345 -pluginUUID 213 -registerEvent test -info "{\"application\":{\"language\":\"en\",\"platform\":\"mac\",\"version\":\"4.1.0\"},\"plugin\":{\"version\":\"1.1\"},\"devicePixelRatio\":2,\"devices\":[{\"id\":\"55F16B35884A859CCE4FFA1FC8D3DE5B\",\"name\":\"Device Name\",\"size\":{\"columns\":5,\"rows\":3},\"type\":0},{\"id\":\"B8F04425B95855CF417199BCB97CD2BB\",\"name\":\"Another Device\",\"size\":{\"columns\":3,\"rows\":2},\"type\":1}]}"

install: build
	cp *.json $(INSTALLDIR)
	cp *.html $(INSTALLDIR)
	cp *.css $(INSTALLDIR)

logs:
	tail -f $(TMP)/streamdeck-vmix.log*