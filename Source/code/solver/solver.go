package solver

import (
	"context"

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

// NewSolver creates a new Solver instance
func NewSolver(logger loggers.Logger) Solver {
	return &solver{
		logger:       logger,
		hostContexts: xsync.NewMapOf[string, []string](),
	}
}

func (s *solver) AddHost(ctx context.Context, host string, context string) {
	s.logger.LogMessage(ctx, "adding host: %s, context: %s", host, context)

	// 既存のcontextを他のhostから削除
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

	// 新しいcontextを追加
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
		// contextを削除
		newContexts := make([]string, 0, len(contexts))
		for _, c := range contexts {
			if c != context {
				newContexts = append(newContexts, c)
			}
		}

		if len(newContexts) == 0 {
			s.logger.LogMessage(ctx, "host %s has no more contexts. deleting host.", host)
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
	if !ok {
		s.logger.LogMessage(ctx, "host %s not found", host)
		return nil, false
	}

	// 多分もうここには入らない
	if len(value) == 0 {
		s.logger.LogMessage(ctx, "host %s has empty contexts", host)
		s.hostContexts.Delete(host)
		return nil, false
	}

	// スライスのディープコピーを作成して返す
	// 元データを保護するため、新しいスライスを作成して返す
	contexts = make([]string, len(value))
	copy(contexts, value)

	s.logger.LogMessage(ctx, "solved contexts for %s: %v", host, contexts)
	return contexts, true
}
