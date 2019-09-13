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

// SetCPUThrottlingRate https://chromedevtools.github.io/devtools-protocol/tot/Emulation#method-setCPUThrottlingRate
func (session *Session) SetCPUThrottlingRate(rate int) error {
	_, err := session.blockingSend("Emulation.setCPUThrottlingRate", &Params{
		"rate": rate,
	})
	return err
}

func (session *Session) clearDeviceMetricsOverride() error {
	_, err := session.blockingSend("Emulation.clearDeviceMetricsOverride", &Params{})
	return err
}

// Metric https://chromedevtools.github.io/devtools-protocol/tot/Performance#type-Metric
type Metric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// PerformanceEnable Enable collecting and reporting metrics
func (session *Session) PerformanceEnable() error {
	_, err := session.blockingSend("Performance.enable", &Params{})
	return err
}

// PerformanceDisable Disable collecting and reporting metrics
func (session *Session) PerformanceDisable() error {
	_, err := session.blockingSend("Performance.disable", &Params{})
	return err
}

// SetPageScaleFactor https://chromedevtools.github.io/devtools-protocol/tot/Emulation#method-setPageScaleFactor
func (session *Session) SetPageScaleFactor(pageScaleFactor int) error {
	_, err := session.blockingSend("Emulation.setPageScaleFactor", &Params{
		"pageScaleFactor": pageScaleFactor,
	})
	return err
}

// PerformanceGetMetrics Retrieve current values of run-time metrics
func (session *Session) PerformanceGetMetrics() ([]Metric, error) {
	msg, err := session.blockingSend("Performance.getMetrics", &Params{})
	if err != nil {
		return nil, err
	}
	m := make([]Metric, 0)
	unmarshal(msg["metrics"], &m)
	return m, nil
}
