package config

import (
	"testing"
	"time"
)

func TestAttachOptionsDefauleValue(t *testing.T) {
	opts := &AttachOptions{}
	t.Run("path", func(t *testing.T) {
		if path := opts.Path(); opts.path == nil && path != "/engine.io" {
			t.Fatalf(`AttachOptions.Path() = %q, want match for %#q`, path, "/engine.io")
		}
	})

	t.Run("destroyUpgrade", func(t *testing.T) {
		if destroyUpgrade := opts.DestroyUpgrade(); opts.destroyUpgrade == nil && destroyUpgrade != true {
			t.Fatalf(`AttachOptions.DestroyUpgrade() = %t, want match for %t`, destroyUpgrade, true)
		}
	})

	t.Run("destroyUpgradeTimeout", func(t *testing.T) {
		if destroyUpgradeTimeout := opts.DestroyUpgradeTimeout(); opts.destroyUpgradeTimeout == nil && destroyUpgradeTimeout != 1000*time.Millisecond {
			t.Fatalf(`AttachOptions.DestroyUpgradeTimeout() = %d, want match for %d`, destroyUpgradeTimeout, 1000*time.Millisecond)
		}
	})
}

func TestServerOptionsDefauleValue(t *testing.T) {
	opts := &ServerOptions{}

	t.Run("pingTimeout", func(t *testing.T) {
		if pingTimeout := opts.PingTimeout(); opts.pingTimeout == nil && pingTimeout != 20000*time.Millisecond {
			t.Fatalf(`ServerOptions.PingTimeout() = %d, want match for %d`, pingTimeout, 20000*time.Millisecond)
		}
	})

	t.Run("pingInterval", func(t *testing.T) {
		if pingInterval := opts.PingInterval(); opts.pingInterval == nil && pingInterval != 25000*time.Millisecond {
			t.Fatalf(`ServerOptions.PingInterval() = %d, want match for %d`, pingInterval, 25000*time.Millisecond)
		}
	})

	t.Run("upgradeTimeout", func(t *testing.T) {
		if upgradeTimeout := opts.UpgradeTimeout(); opts.upgradeTimeout == nil && upgradeTimeout != 10000*time.Millisecond {
			t.Fatalf(`ServerOptions.UpgradeTimeout() = %d, want match for %d`, upgradeTimeout, 10000*time.Millisecond)
		}
	})

	t.Run("maxHttpBufferSize", func(t *testing.T) {
		if maxHttpBufferSize := opts.MaxHttpBufferSize(); opts.maxHttpBufferSize == nil && maxHttpBufferSize != 100000 {
			t.Fatalf(`ServerOptions.MaxHttpBufferSize() = %d, want match for %d`, maxHttpBufferSize, 100000)
		}
	})

	t.Run("allowRequest", func(t *testing.T) {
		if allowRequest := opts.AllowRequest(); opts.allowRequest == nil && allowRequest != nil {
			t.Fatalf(`ServerOptions.AllowRequest() = %v, want match for nil`, allowRequest)
		}
	})

	t.Run("transports", func(t *testing.T) {
		if transports := opts.Transports(); opts.transports == nil && transports != nil && !(transports.Has("polling") && transports.Has("websocket")) {
			t.Fatalf(`ServerOptions.Transports() = %s, want match for ["polling", "websocket")]`, transports.Keys())
		}
	})

	t.Run("allowUpgrades", func(t *testing.T) {
		if allowUpgrades := opts.AllowUpgrades(); opts.allowUpgrades == nil && allowUpgrades != true {
			t.Fatalf(`ServerOptions.AllowUpgrades() = %t, want match for %t`, allowUpgrades, true)
		}
	})

	t.Run("perMessageDeflate", func(t *testing.T) {
		if perMessageDeflate := opts.PerMessageDeflate(); opts.perMessageDeflate == nil && perMessageDeflate != nil {
			t.Fatalf(`ServerOptions.PerMessageDeflate() = %v, want match for nil`, perMessageDeflate)
		}
	})

	t.Run("httpCompression/threshold", func(t *testing.T) {
		if httpCompression := opts.HttpCompression(); opts.httpCompression == nil && httpCompression != nil && httpCompression.Threshold != 1024 {
			t.Fatalf(`ServerOptions.HttpCompression().Threshold = %d, want match for %d`, httpCompression.Threshold, 1024)
		}
	})

	t.Run("initialPacket", func(t *testing.T) {
		if initialPacket := opts.InitialPacket(); opts.initialPacket == nil && initialPacket != nil {
			t.Fatalf(`ServerOptions.InitialPacket() = %v, want match for nil`, initialPacket)
		}
	})

	t.Run("cookie", func(t *testing.T) {
		if cookie := opts.Cookie(); opts.cookie == nil && cookie != nil {
			t.Fatalf(`ServerOptions.Cookie() = %v, want match for nil`, cookie)
		}
	})

	t.Run("cors", func(t *testing.T) {
		if cors := opts.Cors(); opts.cors == nil && cors != nil {
			t.Fatalf(`ServerOptions.Cors() = %v, want match for nil`, cors)
		}
	})

	t.Run("allowEIO3", func(t *testing.T) {
		if allowEIO3 := opts.AllowEIO3(); opts.allowEIO3 == nil && allowEIO3 != false {
			t.Fatalf(`ServerOptions.AllowEIO3() = %t, want match for %t`, allowEIO3, false)
		}
	})
}
