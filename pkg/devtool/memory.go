package devtool

// PressureLevel memory pressure level.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Memory#type-PressureLevel
type PressureLevel string

// String returns the PressureLevel as string value.
func (t PressureLevel) String() string {
	return string(t)
}

// PressureLevel values.
const (
	PressureLevelModerate PressureLevel = "moderate"
	PressureLevelCritical PressureLevel = "critical"
)

// SamplingProfileNode heap profile sample.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Memory#type-SamplingProfileNode
type SamplingProfileNode struct {
	Size  float64  `json:"size"`  // Size of the sampled allocation.
	Total float64  `json:"total"` // Total bytes attributed to this sample.
	Stack []string `json:"stack"` // Execution stack at the point of allocation.
}

// SamplingProfile array of heap profile samples.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Memory#type-SamplingProfile
type SamplingProfile struct {
	Samples []*SamplingProfileNode `json:"samples"`
	Modules []*Module              `json:"modules"`
}

// Module executable module information.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Memory#type-Module
type Module struct {
	Name        string  `json:"name"`        // Name of the module.
	UUID        string  `json:"uuid"`        // UUID of the module.
	BaseAddress string  `json:"baseAddress"` // Base address where the module is loaded into memory. Encoded as a decimal or hexadecimal (0x prefixed) string.
	Size        float64 `json:"size"`        // Size of the module in bytes.
}
