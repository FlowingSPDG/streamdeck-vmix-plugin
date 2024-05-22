package stdvmix

import (
	"github.com/puzpuzpuz/xsync/v3"
)

type activatorColor int

const (
	activatorColorInvalid activatorColor = iota
	activatorColorRed
	activatorColorGreen
)

type activatorContexts struct {
	// Activatorsの設定
	// StreamDeck側としてはevent.Context を使って参照したい → 設定を変更した際にdestinationだと負えなくなるため
	// vMix側としてはdestination, input, activatorName で参照したい
	contextKeys *xsync.MapOf[string, activatorContext] // key:context value:activatorContext
}

type activatorContext struct {
	destination    string
	input          int
	activatorName  string
	activatorColor activatorColor
}

func newActivatorContexts() *activatorContexts {
	return &activatorContexts{
		contextKeys: xsync.NewMapOf[string, activatorContext](),
	}
}

func (ac *activatorContexts) Store(key string, ctx activatorContext) {
	ac.contextKeys.Store(key, ctx)
}

func (ac *activatorContexts) Delete(key string) {
	ac.contextKeys.Delete(key)
}
