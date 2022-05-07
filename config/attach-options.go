package config

import (
	"time"
)

type AttachOptions struct {
	InternalPath                  *string        `json:"path,omitempty"`
	InternalDestroyUpgrade        *bool          `json:"destroyUpgrade,omitempty"`
	InternalDestroyUpgradeTimeout *time.Duration `json:"destroyUpgradeTimeout,omitempty"`
}

func DefaultAttachOptions() *AttachOptions {
	a := &AttachOptions{}
	a.SetPath("/engine.io")
	a.SetDestroyUpgradeTimeout(time.Duration(1000 * time.Millisecond))
	a.SetDestroyUpgrade(true)
	return a
}

// name of the path to capture
// @default "/engine.io"
func (a *AttachOptions) SetPath(path string) {
	a.InternalPath = &path
}
func (a *AttachOptions) Path() string {
	if a.InternalPath == nil {
		return "/engine.io"
	}

	return *a.InternalPath
}

// destroy unhandled upgrade requests
// @default true
func (a *AttachOptions) SetDestroyUpgrade(destroyUpgrade bool) {
	a.InternalDestroyUpgrade = &destroyUpgrade
}
func (a *AttachOptions) DestroyUpgrade() bool {
	if a.InternalDestroyUpgrade == nil {
		return true
	}

	return *a.InternalDestroyUpgrade
}

// milliseconds after which unhandled requests are ended
// @default 1000
func (a *AttachOptions) SetDestroyUpgradeTimeout(destroyUpgradeTimeout time.Duration) {
	a.InternalDestroyUpgradeTimeout = &destroyUpgradeTimeout
}
func (a *AttachOptions) DestroyUpgradeTimeout() time.Duration {
	if a.InternalDestroyUpgradeTimeout == nil {
		return time.Duration(1000 * time.Millisecond)
	}

	return *a.InternalDestroyUpgradeTimeout
}