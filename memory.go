package cdp

import "github.com/ecwid/cdp/pkg/devtool"

// GetAllTimeSamplingProfile ...
func (mem Memory) GetAllTimeSamplingProfile() (*devtool.SamplingProfile, error) {
	result := new(devtool.SamplingProfile)
	if err := mem.call("Memory.getAllTimeSamplingProfile", nil, result); err != nil {
		return nil, err
	}
	return result, nil
}
