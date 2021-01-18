APPNAME=dev.flowingspdg.streamdeck.vmix.sdPlugin

BIN_NAME = streamdeck-vmix-plugin
BUILDDIR = ./build
INSTALLDIR = $(APPDATA)\Elgato\StreamDeck\Plugins\$(APPNAME)

# Replacing "INSTALLDIR" directory for Mac.
ifeq  ($(shell uname),Darwin)
    INSTALLDIR = ~/Library/Application\ Support/com.elgato.StreamDeck/Plugins/$(APPNAME)/
endif

# Replacing "INSTALLDIR" directory for Windows.
ifeq ($(OS),Windows_NT)
    INSTALLDIR = $(APPDATA)\\Elgato\\StreamDeck\\Plugins\\$(APPNAME)
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
	@-$(RM) $(INSTALLDIR)
	@-$(MKDIR) $(INSTALLDIR)
	@-$(MKDIR) $(BUILDDIR)
	@-$(RM) $(BUILDDIR)/*

build: prepare
	$(CP) ./manifest.json $(BUILDDIR)
	$(CP) ./inspector $(BUILDDIR)

install: build
	cp *.json $(INSTALLDIR)
	cp *.html $(INSTALLDIR)
	cp *.css $(INSTALLDIR)

logs:
	tail -f $(TMP)/streamdeck-vmix.log*