package cdp

import (
	"github.com/ecwid/cdp/pkg/devtool"
	"github.com/ecwid/cdp/pkg/mobile"
)

// SetDeviceMetricsOverride ...
func (session Emulation) SetDeviceMetricsOverride(metrics *devtool.DeviceMetrics) error {
	return session.call("Emulation.setDeviceMetricsOverride", metrics, nil)
}

// SetUserAgent set user agent
func (session Emulation) SetUserAgent(userAgent string) error {
	return session.SetUserAgentOverride(userAgent, nil, nil)
}

// SetUserAgentOverride ...
func (session Emulation) SetUserAgentOverride(userAgent string, acceptLanguage, platform *string) error {
	p := Map{"userAgent": userAgent}
	if acceptLanguage != nil {
		p["acceptLanguage"] = acceptLanguage
	}
	if platform != nil {
		p["platform"] = platform
	}
	return session.call("Emulation.setUserAgentOverride", p, nil)
}

// ClearDeviceMetricsOverride ...
func (session Emulation) ClearDeviceMetricsOverride() error {
	return session.call("Emulation.clearDeviceMetricsOverride", nil, nil)
}

// SetScrollbarsHidden ...
func (session Emulation) SetScrollbarsHidden(hidden bool) error {
	return session.call("Emulation.setScrollbarsHidden", Map{"hidden": hidden}, nil)
}

// SetCPUThrottlingRate https://chromedevtools.github.io/devtools-protocol/tot/Emulation#method-setCPUThrottlingRate
func (session Emulation) SetCPUThrottlingRate(rate int) error {
	return session.call("Emulation.setCPUThrottlingRate", Map{"rate": rate}, nil)
}

// SetDocumentCookieDisabled https://chromedevtools.github.io/devtools-protocol/tot/Emulation/#method-setDocumentCookieDisabled
func (session Emulation) SetDocumentCookieDisabled(disabled bool) error {
	return session.call("Emulation.setDocumentCookieDisabled", Map{"disabled": disabled}, nil)
}

// Emulate emulate predefined device
func (session Emulation) Emulate(device *mobile.Device) error {
	f := true
	device.Metrics.DontSetVisibleSize = &f
	if err := session.SetDeviceMetricsOverride(device.Metrics); err != nil {
		return err
	}
	return session.SetUserAgent(device.UserAgent)
}
