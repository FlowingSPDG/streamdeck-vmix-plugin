package solver

import (
	"slices"

	"github.com/puzpuzpuz/xsync/v3"
)

// Solver solves the mapping between vMix host and StreamDeck contexts.
type Solver interface {
	SolveByContext(context string) (host string, ok bool)
	SolveByHost(host string) (contexts []string, ok bool)

	AddHost(host string, context string)

	RemoveContext(context string) (hostRemoved bool)
}

type solver struct {
	// host -> contexts
	hostContexts *xsync.MapOf[string, []string]
}

// NewSolver creates a new Solver.
func NewSolver() Solver {
	return &solver{
		hostContexts: xsync.NewMapOf[string, []string](),
	}
}

func (s *solver) AddHost(host string, context string) {
	actal, loaded := s.hostContexts.LoadOrStore(host, []string{context})
	if loaded {
		actal = append(actal, context)
		s.hostContexts.Store(host, actal)
	}
}

func (s *solver) RemoveContext(context string) bool {
	removed := false
	s.hostContexts.Range(func(host string, contexts []string) bool {
		contexts = slices.DeleteFunc(contexts, func(c string) bool {
			return c == context
		})

		if len(contexts) == 0 {
			s.hostContexts.Delete(host)
			removed = true
			return false
		}

		s.hostContexts.Store(host, contexts)
		return true
	})
	return removed
}

func (s *solver) SolveByContext(context string) (host string, ok bool) {
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

func (s *solver) SolveByHost(host string) (contexts []string, ok bool) {
	return s.hostContexts.Load(host)
}
