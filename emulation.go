package cdp

import (
	"github.com/ecwid/cdp/pkg/devtool"
	"github.com/ecwid/cdp/pkg/mobile"
)

// SetDeviceMetricsOverride ...
func (emu Emulation) SetDeviceMetricsOverride(metrics *devtool.DeviceMetrics) error {
	return emu.call("Emulation.setDeviceMetricsOverride", metrics, nil)
}

// SetUserAgent set user agent
func (emu Emulation) SetUserAgent(userAgent string) error {
	return emu.SetUserAgentOverride(userAgent, nil, nil)
}

// SetUserAgentOverride ...
func (emu Emulation) SetUserAgentOverride(userAgent string, acceptLanguage, platform *string) error {
	p := Map{"userAgent": userAgent}
	if acceptLanguage != nil {
		p["acceptLanguage"] = acceptLanguage
	}
	if platform != nil {
		p["platform"] = platform
	}
	return emu.call("Emulation.setUserAgentOverride", p, nil)
}

// ClearDeviceMetricsOverride ...
func (emu Emulation) ClearDeviceMetricsOverride() error {
	return emu.call("Emulation.clearDeviceMetricsOverride", nil, nil)
}

// SetScrollbarsHidden ...
func (emu Emulation) SetScrollbarsHidden(hidden bool) error {
	return emu.call("Emulation.setScrollbarsHidden", Map{"hidden": hidden}, nil)
}

// SetCPUThrottlingRate https://chromedevtools.github.io/devtools-protocol/tot/Emulation#method-setCPUThrottlingRate
func (emu Emulation) SetCPUThrottlingRate(rate int) error {
	return emu.call("Emulation.setCPUThrottlingRate", Map{"rate": rate}, nil)
}

// Emulate emulate predefined device
func (emu Emulation) Emulate(device *mobile.Device) error {
	f := true
	device.Metrics.DontSetVisibleSize = &f
	if err := emu.SetDeviceMetricsOverride(device.Metrics); err != nil {
		return err
	}
	return emu.SetUserAgent(device.UserAgent)
}
