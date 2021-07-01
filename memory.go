package cdp

import "github.com/ecwid/cdp/pkg/devtool"

// GetAllTimeSamplingProfile ...
func (session Memory) GetAllTimeSamplingProfile() (*devtool.SamplingProfile, error) {
	result := new(devtool.SamplingProfile)
	if err := session.call("Memory.getAllTimeSamplingProfile", nil, result); err != nil {
		return nil, err
	}
	return result, nil
}
