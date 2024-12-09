package solver

import (
	"context"
	"slices"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger/loggers"
	"github.com/puzpuzpuz/xsync/v3"
)

// Solver solves the mapping between vMix host and StreamDeck contexts.
type Solver interface {
	SolveByContext(ctx context.Context, context string) (host string, ok bool)
	SolveByHost(ctx context.Context, host string) (contexts []string, ok bool)

	AddHost(ctx context.Context, host string, context string)

	RemoveContext(ctx context.Context, context string) (hostRemoved bool)
}

type solver struct {
	logger       loggers.Logger
	hostContexts *xsync.MapOf[string, []string]
}

// NewSolver creates a new Solver.
func NewSolver(logger loggers.Logger) Solver {
	return &solver{
		logger:       logger,
		hostContexts: xsync.NewMapOf[string, []string](),
	}
}

func (s *solver) AddHost(ctx context.Context, host string, context string) {
	s.logger.LogMessage(ctx, "adding host: %s, context: %s", host, context)
	defer func() {
		value, _ := s.hostContexts.Load(host)
		s.logger.LogMessage(ctx, "AddHost for %s: Result: %v", host, value)
		s.hostContexts.Range(func(key string, value []string) bool {
			s.logger.LogMessage(ctx, "hostContexts [%s/%s]", key, value)
			return true
		})
	}()

	// contextが別のhostに紐づいていた場合、削除する
	s.hostContexts.Range(func(h string, contexts []string) bool {
		if h == host {
			return true
		}
		for _, c := range contexts {
			if c == context {
				s.logger.LogMessage(ctx, "context %s is already used by host %s. removing.", context, h)
				s.RemoveContext(ctx, context)
				return false
			}
		}
		return true
	})

	loaded, ok := s.hostContexts.Load(host)
	if ok {
		s.logger.LogMessage(ctx, "found contexts for host %s. appending.", host)
		loaded = append(loaded, context)
		s.hostContexts.Store(host, loaded)
		return
	}
	s.hostContexts.Store(host, []string{context})
}

func (s *solver) RemoveContext(ctx context.Context, context string) bool {
	removed := false
	s.logger.LogMessage(ctx, "removing context: %s", context)
	s.hostContexts.Range(func(host string, contexts []string) bool {
		contexts = slices.DeleteFunc(contexts, func(c string) bool {
			return c == context
		})

		if len(contexts) == 0 {
			s.logger.LogMessage(ctx, "destination %s is no longer used. delete!", host)
			s.hostContexts.Delete(host)
			removed = true
			return false
		}

		s.hostContexts.Store(host, contexts)
		return true
	})
	return removed
}

func (s *solver) SolveByContext(ctx context.Context, context string) (host string, ok bool) {
	ok = false
	s.hostContexts.Range(func(h string, contexts []string) bool {
		for _, c := range contexts {
			if c == context {
				host = h
				ok = true
				return false
			}
		}
		return true
	})
	return
}

func (s *solver) SolveByHost(ctx context.Context, host string) (contexts []string, ok bool) {
	s.logger.LogMessage(ctx, "solving contexts for host: %s", host)
	value, ok := s.hostContexts.Load(host)
	s.logger.LogMessage(ctx, "solved:%v(%v)", value, ok)
	return value, ok
}
