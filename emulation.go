package cdp

func (session *Session) setDeviceMetricsOverride(width, height int64, deviceScaleFactor float64) error {
	_, err := session.blockingSend("Emulation.setDeviceMetricsOverride", &Params{
		"width":             width,
		"height":            height,
		"deviceScaleFactor": deviceScaleFactor,
		"mobile":            false,
	})
	return err
}

func (session *Session) clearDeviceMetricsOverride() error {
	_, err := session.blockingSend("Emulation.clearDeviceMetricsOverride", &Params{})
	return err
}
