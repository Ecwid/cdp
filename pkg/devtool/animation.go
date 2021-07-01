package devtool

type KeyframeStyle struct {
	Offset string `json:"offset"`
	Easing string `json:"easing"`
}

type KeyframesRule struct {
	Name      string          `json:"name"`
	Keyframes []KeyframeStyle `json:"keyframes"`
}

type AnimationEffect struct {
	Delay          float64        `json:"delay"`
	EndDelay       float64        `json:"endDelay"`
	IterationStart float64        `json:"iterationStart"`
	Iterations     float64        `json:"iterations"`
	Duration       float64        `json:"duration"`
	Direction      string         `json:"direction"`
	Fill           string         `json:"fill"`
	BackendNodeId  BackendNodeID  `json:"backendNodeId"`
	KeyframesRule  *KeyframesRule `json:"keyframesRule"`
	Easing         string         `json:"easing"`
}
type Animation struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	PausedState  bool             `json:"pausedState"`
	PlayState    string           `json:"playState"`
	PlaybackRate float64          `json:"playbackRate"`
	StartTime    float64          `json:"startTime"`
	CurrentTime  float64          `json:"currentTime"`
	Type         float64          `json:"type"`
	Source       *AnimationEffect `json:"source"`
	CssID        string           `json:"cssId"`
}
