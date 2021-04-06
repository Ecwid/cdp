package devtool

// ColorFormat https://chromedevtools.github.io/devtools-protocol/tot/Overlay/#type-ColorFormat
type ColorFormat string

// ContrastAlgorithm https://chromedevtools.github.io/devtools-protocol/tot/Overlay/#type-ContrastAlgorithm
type ContrastAlgorithm string

const (
	ColorFormatRGB ColorFormat = "rgb"
	ColorFormatHSL ColorFormat = "hsl"
	ColorFormatHEX ColorFormat = "hex"
)

// LineStyle https://chromedevtools.github.io/devtools-protocol/tot/Overlay/#type-LineStyle
type LineStyle struct {
	Color   *RGBA  `json:"color,omitempty"`
	Pattern string `json:"pattern,omitempty"`
}

// BoxStyle https://chromedevtools.github.io/devtools-protocol/tot/Overlay/#type-BoxStyle
type BoxStyle struct {
	FillColor  *RGBA `json:"fillColor,omitempty"`
	HatchColor *RGBA `json:"hatchColor,omitempty"`
}

// GridHighlightConfig https://chromedevtools.github.io/devtools-protocol/tot/Overlay/#type-GridHighlightConfig
type GridHighlightConfig struct {
	ShowGridExtensionLines  bool  `json:"showGridExtensionLines,omitempty"`  // Whether the extension lines from grid cells to the rulers should be shown (default: false).
	ShowPositiveLineNumbers bool  `json:"showPositiveLineNumbers,omitempty"` // Show Positive line number labels (default: false).
	ShowNegativeLineNumbers bool  `json:"showNegativeLineNumbers,omitempty"` // Show Negative line number labels (default: false).
	ShowAreaNames           bool  `json:"showAreaNames,omitempty"`           // Show area name labels (default: false).
	ShowLineNames           bool  `json:"showLineNames,omitempty"`           // Show line name labels (default: false).
	ShowTrackSizes          bool  `json:"showTrackSizes,omitempty"`          // Show track size labels (default: false).
	GridBorderColor         *RGBA `json:"gridBorderColor,omitempty"`         // The grid container border highlight color (default: transparent).
	RowLineColor            *RGBA `json:"rowLineColor,omitempty"`            // The row line color (default: transparent).
	ColumnLineColor         *RGBA `json:"columnLineColor,omitempty"`         // The column line color (default: transparent).
	GridBorderDash          bool  `json:"gridBorderDash,omitempty"`          // Whether the grid border is dashed (default: false).
	RowLineDash             bool  `json:"rowLineDash,omitempty"`             // Whether row lines are dashed (default: false).
	ColumnLineDash          bool  `json:"columnLineDash,omitempty"`          // Whether column lines are dashed (default: false).
	RowGapColor             *RGBA `json:"rowGapColor,omitempty"`             // The row gap highlight fill color (default: transparent).
	RowHatchColor           *RGBA `json:"rowHatchColor,omitempty"`           // The row gap hatching fill color (default: transparent).
	ColumnGapColor          *RGBA `json:"columnGapColor,omitempty"`          // The column gap highlight fill color (default: transparent).
	ColumnHatchColor        *RGBA `json:"columnHatchColor,omitempty"`        // The column gap hatching fill color (default: transparent).
	AreaBorderColor         *RGBA `json:"areaBorderColor,omitempty"`         // The named grid areas border color (Default: transparent).
	GridBackgroundColor     *RGBA `json:"gridBackgroundColor,omitempty"`     // The grid container background color (Default: transparent).
}

// FlexContainerHighlightConfig https://chromedevtools.github.io/devtools-protocol/tot/Overlay/#type-FlexContainerHighlightConfig
type FlexContainerHighlightConfig struct {
	ContainerBorder       *LineStyle `json:"containerBorder,omitempty"`       // The style of the container border
	LineSeparator         *LineStyle `json:"lineSeparator,omitempty"`         // The style of the separator between lines
	ItemSeparator         *LineStyle `json:"itemSeparator,omitempty"`         // The style of the separator between items
	MainDistributedSpace  *BoxStyle  `json:"mainDistributedSpace,omitempty"`  // Style of content-distribution space on the main axis (justify-content).
	CrossDistributedSpace *BoxStyle  `json:"crossDistributedSpace,omitempty"` // Style of content-distribution space on the cross axis (align-content).
	RowGapSpace           *BoxStyle  `json:"rowGapSpace,omitempty"`           // Style of empty space caused by row gaps (gap/row-gap).
	ColumnGapSpace        *BoxStyle  `json:"columnGapSpace,omitempty"`        // Style of empty space caused by columns gaps (gap/column-gap).
	CrossAlignment        *LineStyle `json:"crossAlignment,omitempty"`        // Style of the self-alignment line (align-items).
}

// FlexItemHighlightConfig https://chromedevtools.github.io/devtools-protocol/tot/Overlay/#type-FlexItemHighlightConfig
type FlexItemHighlightConfig struct {
	BaseSizeBox      *BoxStyle  `json:"baseSizeBox,omitempty"`      // Style of the box representing the item's base size
	BaseSizeBorder   *LineStyle `json:"baseSizeBorder,omitempty"`   // Style of the border around the box representing the item's base size
	FlexibilityArrow *LineStyle `json:"flexibilityArrow,omitempty"` // Style of the arrow representing if the item grew or shrank
}

// HighlightConfig https://chromedevtools.github.io/devtools-protocol/tot/Overlay/#type-HighlightConfig
type HighlightConfig struct {
	ShowInfo                     bool                          `json:"showInfo,omitempty"`                     // Whether the node info tooltip should be shown (default: false).
	ShowStyles                   bool                          `json:"showStyles,omitempty"`                   // Whether the node styles in the tooltip (default: false).
	ShowRulers                   bool                          `json:"showRulers,omitempty"`                   // Whether the rulers should be shown (default: false).
	ShowAccessibilityInfo        bool                          `json:"showAccessibilityInfo,omitempty"`        // Whether the a11y info should be shown (default: true).
	ShowExtensionLines           bool                          `json:"showExtensionLines,omitempty"`           // Whether the extension lines from node to the rulers should be shown (default: false).
	ContentColor                 *RGBA                         `json:"contentColor,omitempty"`                 // The content box highlight fill color (default: transparent).
	PaddingColor                 *RGBA                         `json:"paddingColor,omitempty"`                 // The padding highlight fill color (default: transparent).
	BorderColor                  *RGBA                         `json:"borderColor,omitempty"`                  // The border highlight fill color (default: transparent).
	MarginColor                  *RGBA                         `json:"marginColor,omitempty"`                  // The margin highlight fill color (default: transparent).
	EventTargetColor             *RGBA                         `json:"eventTargetColor,omitempty"`             // The event target element highlight fill color (default: transparent).
	ShapeColor                   *RGBA                         `json:"shapeColor,omitempty"`                   // The shape outside fill color (default: transparent).
	ShapeMarginColor             *RGBA                         `json:"shapeMarginColor,omitempty"`             // The shape margin fill color (default: transparent).
	CSSGridColor                 *RGBA                         `json:"cssGridColor,omitempty"`                 // The grid layout color (default: transparent).
	ColorFormat                  ColorFormat                   `json:"colorFormat,omitempty"`                  // The color format used to format color styles (default: hex).
	GridHighlightConfig          *GridHighlightConfig          `json:"gridHighlightConfig,omitempty"`          // The grid layout highlight configuration (default: all transparent).
	FlexContainerHighlightConfig *FlexContainerHighlightConfig `json:"flexContainerHighlightConfig,omitempty"` // The flex container highlight configuration (default: all transparent).
	FlexItemHighlightConfig      *FlexItemHighlightConfig      `json:"flexItemHighlightConfig,omitempty"`      // The flex item highlight configuration (default: all transparent).
	ContrastAlgorithm            ContrastAlgorithm             `json:"contrastAlgorithm,omitempty"`            // The contrast algorithm to use for the contrast ratio (default: aa).
}
