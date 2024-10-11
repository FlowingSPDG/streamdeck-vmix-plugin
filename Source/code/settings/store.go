package settings

import "github.com/puzpuzpuz/xsync/v3"

// StreamDeckのcontext(string)から実際のPIデータを型情報込みで引き抜く必要がある

type SettingStore[T any] interface {
	Load(key string) (value *T, ok bool)
	LoadOrStore(key string, value *T) (actual *T, ok bool)
	Store(key string, setting *T)
	Delete(key string)
}

func NewSettingStore[T any]() SettingStore[T] {
	return &settingStore[T]{
		m: xsync.NewMapOf[string, *T](),
	}
}

type settingStore[T any] struct {
	m *xsync.MapOf[string, *T]
}

func (s *settingStore[T]) Store(key string, setting *T) {
	s.m.Store(key, setting)
}

func (s *settingStore[T]) Load(key string) (value *T, ok bool) {
	return s.m.Load(key)
}

func (s *settingStore[T]) LoadOrStore(key string, value *T) (actual *T, ok bool) {
	return s.m.LoadOrStore(key, value)
}

func (s *settingStore[T]) Delete(key string) {
	s.m.Delete(key)
}
