package expires

import (
	"time"
)

type (
	Expires interface {
		IsExpires() bool
		SetExpires(e bool)
	}

	expires struct {
		expiresAt time.Time
		abort     bool
	}
)

func NewExpires(t time.Duration) Expires {
	return &expires{
		expiresAt: time.Now().Add(t),
	}
}

func (ep *expires) SetExpires(e bool) {
	ep.abort = !e
}

func (ep *expires) IsExpires() bool {
	return ep.abort || time.Now().Sub(ep.expiresAt) > 0
}
