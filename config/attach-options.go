package config

import (
	"time"
)

type AttachOptionsInterface interface {
	SetPath(string)
	GetRawPath() *string
	Path() string

	SetDestroyUpgrade(bool)
	GetRawDestroyUpgrade() *bool
	DestroyUpgrade() bool

	SetDestroyUpgradeTimeout(time.Duration)
	GetRawDestroyUpgradeTimeout() *time.Duration
	DestroyUpgradeTimeout() time.Duration
}

type AttachOptions struct {
	// name of the path to capture
	path *string `json:"path,omitempty"`

	// destroy unhandled upgrade requests
	destroyUpgrade *bool `json:"destroyUpgrade,omitempty"`

	//  milliseconds after which unhandled requests are ended
	destroyUpgradeTimeout *time.Duration `json:"destroyUpgradeTimeout,omitempty"`
}

func DefaultAttachOptions() *AttachOptions {
	a := &AttachOptions{}
	// a.SetPath("/engine.io")
	// a.SetDestroyUpgradeTimeout(time.Duration(1000 * time.Millisecond))
	// a.SetDestroyUpgrade(true)
	return a
}

func (a *AttachOptions) Assign(data AttachOptionsInterface) AttachOptionsInterface {
	if data == nil {
		return a
	}

	if a.GetRawPath() == nil {
		a.SetPath(data.Path())
	}

	if a.GetRawDestroyUpgradeTimeout() == nil {
		a.SetDestroyUpgradeTimeout(data.DestroyUpgradeTimeout())
	}

	if a.GetRawDestroyUpgrade() == nil {
		a.SetDestroyUpgrade(data.DestroyUpgrade())
	}

	return a
}

// name of the path to capture
// @default "/engine.io"
func (a *AttachOptions) SetPath(path string) {
	a.path = &path
}
func (a *AttachOptions) GetRawPath() *string {
	return a.path
}
func (a *AttachOptions) Path() string {
	if a.path == nil {
		return "/engine.io"
	}

	return *a.path
}

// destroy unhandled upgrade requests
// @default true
func (a *AttachOptions) SetDestroyUpgrade(destroyUpgrade bool) {
	a.destroyUpgrade = &destroyUpgrade
}
func (a *AttachOptions) GetRawDestroyUpgrade() *bool {
	return a.destroyUpgrade
}
func (a *AttachOptions) DestroyUpgrade() bool {
	if a.destroyUpgrade == nil {
		return true
	}

	return *a.destroyUpgrade
}

// milliseconds after which unhandled requests are ended
// @default 1000
func (a *AttachOptions) SetDestroyUpgradeTimeout(destroyUpgradeTimeout time.Duration) {
	a.destroyUpgradeTimeout = &destroyUpgradeTimeout
}
func (a *AttachOptions) GetRawDestroyUpgradeTimeout() *time.Duration {
	return a.destroyUpgradeTimeout
}
func (a *AttachOptions) DestroyUpgradeTimeout() time.Duration {
	if a.destroyUpgradeTimeout == nil {
		return time.Duration(1000 * time.Millisecond)
	}

	return *a.destroyUpgradeTimeout
}
