package setting

import "github.com/puzpuzpuz/xsync/v3"

// Setting defines the common interface for all settings
type Setting interface {
	IsDefault() bool
	Initialize()
	GetVMixAddress() string
	GetTallyMode() TallyMode
	GetMix() int
	GetInput() *int
}

type TallyMode int

const (
	_ TallyMode = iota
	TallyModeTALLY
	TallyModeACTS
	TallyModeDisabled
)

type SettingStore[T any] interface {
	Load(key string) (value T, ok bool)
	LoadOrStore(key string, value T) (actual T, ok bool)
	Store(key string, setting T)
	Delete(key string)
	Range(f func(key string, value T) bool)
}

func NewSettingStore[T any]() SettingStore[T] {
	return &settingStore[T]{
		m: xsync.NewMapOf[string, T](),
	}
}

type settingStore[T any] struct {
	m *xsync.MapOf[string, T]
}

func (s *settingStore[T]) Store(key string, setting T) {
	s.m.Store(key, setting)
}

func (s *settingStore[T]) Load(key string) (value T, ok bool) {
	return s.m.Load(key)
}

func (s *settingStore[T]) LoadOrStore(key string, value T) (actual T, ok bool) {
	return s.m.LoadOrStore(key, value)
}

func (s *settingStore[T]) Delete(key string) {
	s.m.Delete(key)
}

func (s *settingStore[T]) Range(f func(key string, value T) bool) {
	s.m.Range(func(key string, value T) bool {
		return f(key, value)
	})
}
