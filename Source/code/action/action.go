// Package action provides the action interface and its implementations.
//
// Each Action interface does not belong to a specific "button".
// Instead, it is a usecase(logics) that can own any button.
// Action interface owns each StreamDeck context internally.

package action

import "context"

type Action interface {
	Tally(ctx context.Context, host string, input int) error
	Activator(ctx context.Context) error // TODO...
}
