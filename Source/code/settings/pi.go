package settings

type PropertyInspectorSettings interface {
	IsDefault() bool
	Initialize()
}
