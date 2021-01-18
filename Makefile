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

prepare:
	@$(MKDIR) $(BUILDDIR)
	@$(RM) $(BUILDDIR)/*

build: prepare
	$(CP) $(MAKEFILE_DIR)/manifest.json $(BUILDDIR)
	$(CP) $(MAKEFILE_DIR)/inspector $(BUILDDIR)