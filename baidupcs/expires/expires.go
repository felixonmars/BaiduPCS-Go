package expires

import (
	"fmt"
	"time"
	_ "unsafe" // for go:linkname
)

type (
	Expires interface {
		IsExpires() bool
		GetExpires() time.Time
		SetExpires(e bool)
		fmt.Stringer
	}

	expires struct {
		expiresAt time.Time
		abort     bool
	}
)

// StripMono strip monotonic clocks
//go:linkname StripMono time.(*Time).stripMono
func StripMono(t *time.Time)

func NewExpires(dur time.Duration) Expires {
	t := time.Now().Add(dur)
	StripMono(&t)
	return &expires{
		expiresAt: t,
	}
}

func NewExpiresAt(at time.Time) Expires {
	StripMono(&at)
	return &expires{
		expiresAt: at,
	}
}

func (ep *expires) GetExpires() time.Time {
	return ep.expiresAt
}

func (ep *expires) SetExpires(e bool) {
	ep.abort = e
}

func (ep *expires) IsExpires() bool {
	return ep.abort || time.Now().After(ep.expiresAt)
}

func (ep *expires) String() string {
	return fmt.Sprintf("expires at: %s, abort: %t", ep.expiresAt, ep.abort)
}
