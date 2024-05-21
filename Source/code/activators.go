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
	contextKeys *xsync.MapOf[activatorKey, []activatorContext]
}

type activatorContext struct {
	ctxStr         string
	activatorColor activatorColor
}

type activatorKey struct {
	input         int
	activatorName string
}

func newActivatorContexts() *activatorContexts {
	return &activatorContexts{
		contextKeys: xsync.NewMapOf[activatorKey, []activatorContext](),
	}
}

func (ac *activatorContexts) Store(key activatorKey, ctx activatorContext) {
	contexts, _ := ac.contextKeys.LoadOrStore(key, []activatorContext{})
	ac.contextKeys.Store(key, append(contexts, ctx))
}

func (ac *activatorContexts) Delete(key activatorKey, ctxStr string) {
	tallies, ok := ac.contextKeys.Load(key)
	if !ok {
		return
	}
	newTallies := make([]activatorContext, 0, len(tallies)-1)
	for _, c := range tallies {
		if c.ctxStr == ctxStr {
			continue
		}
		newTallies = append(newTallies, c)
	}
	ac.contextKeys.Store(key, newTallies)
}

func (ac *activatorContexts) DeleteByContext(ctxStr string) {
	ac.contextKeys.Range(func(key activatorKey, tallies []activatorContext) bool {
		for _, tally := range tallies {
			if tally.ctxStr == ctxStr {
				ac.Delete(key, ctxStr)
				return false
			}
		}
		return true
	})
}
