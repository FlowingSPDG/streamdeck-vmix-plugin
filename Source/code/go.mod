module github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code

go 1.23

toolchain go1.23.0

require (
	github.com/FlowingSPDG/streamdeck v0.0.0-20250312080211-6e0c0c0223d6
	github.com/FlowingSPDG/vmix-go v0.2.4-0.20250311203920-96f5d7585454
	github.com/puzpuzpuz/xsync/v3 v3.4.0
	github.com/samber/lo v1.49.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.10.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)

replace github.com/FlowingSPDG/vmix-go => ../../../vmix-go