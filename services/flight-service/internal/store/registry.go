package store

import "sync/atomic"

type Registry struct {
	Snap atomic.Pointer[FlightsSnapshot]
}
