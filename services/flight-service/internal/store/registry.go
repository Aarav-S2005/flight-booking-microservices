package store

import "sync/atomic"

type Registry struct {
	snap atomic.Pointer[FlightsSnapshot]
}
