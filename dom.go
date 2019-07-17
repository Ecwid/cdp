package cdp

// Node https://chromedevtools.github.io/devtools-protocol/tot/DOM#type-Node
type Node struct {
	NodeID           int64    `json:"nodeId"`
	ParentID         int64    `json:"parentId"`
	BackendNodeID    int64    `json:"backendNodeId"`
	NodeType         int64    `json:"nodeType"`
	NodeName         string   `json:"nodeName"`
	LocalName        string   `json:"localName"`
	NodeValue        string   `json:"nodeValue"`
	ChildNodeCount   int64    `json:"childNodeCount"`
	Children         []*Node  `json:"children"`
	Attributes       []string `json:"attributes"`
	DocumentURL      string   `json:"documentURL"`
	BaseURL          string   `json:"baseURL"`
	PublicID         string   `json:"publicId"`
	SystemID         string   `json:"systemId"`
	InternalSubset   string   `json:"internalSubset"`
	XMLVersion       string   `json:"xmlVersion"`
	Name             string   `json:"name"`
	Value            string   `json:"value"`
	PseudoType       string   `json:"pseudoType"`
	ShadowRootType   string   `json:"shadowRootType"`
	FrameID          string   `json:"frameId"`
	ContentDocument  *Node    `json:"contentDocument"`
	ShadowRoots      []*Node  `json:"shadowRoots"`
	TemplateContent  *Node    `json:"templateContent"`
	PseudoElements   []*Node  `json:"pseudoElements"`
	ImportedDocument *Node    `json:"importedDocument"`
	IsSVG            bool     `json:"isSVG"`
}

type quad []float64

type highlightConfig struct {
	ShowInfo           bool  `json:"showInfo,omitempty"`
	ShowStyles         bool  `json:"showStyles,omitempty"`
	ShowRulers         bool  `json:"showRulers,omitempty"`
	ShowExtensionLines bool  `json:"showExtensionLines,omitempty"`
	ContentColor       *rgba `json:"contentColor,omitempty"`
	PaddingColor       *rgba `json:"paddingColor,omitempty"`
	BorderColor        *rgba `json:"borderColor,omitempty"`
	MarginColor        *rgba `json:"marginColor,omitempty"`
	EventTargetColor   *rgba `json:"eventTargetColor,omitempty"`
	ShapeColor         *rgba `json:"shapeColor,omitempty"`
	ShapeMarginColor   *rgba `json:"shapeMarginColor,omitempty"`
	CSSGridColor       *rgba `json:"cssGridColor,omitempty"`
}

// Rect https://chromedevtools.github.io/devtools-protocol/tot/DOM#type-Rect
type Rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type rgba struct {
	R int64   `json:"r"`
	G int64   `json:"g"`
	B int64   `json:"b"`
	A float64 `json:"a,omitempty"`
}

func (q quad) middle() (float64, float64) {
	x := 0.0
	y := 0.0
	for i := 0; i < 8; i += 2 {
		x += q[i]
		y += q[i+1]
	}
	return x / 4, y / 4
}

func (session *Session) focus(objectID string) error {
	_, err := session.blockingSend("DOM.focus", &Params{"objectId": objectID})
	return err
}

func (session *Session) getFrameOwner(frameID string) (int64, error) {
	msg, err := session.blockingSend("DOM.getFrameOwner", &Params{"frameId": frameID})
	if err != nil {
		return 0, err
	}
	nbf := msg["backendNodeId"].(float64)
	return int64(nbf), nil
}

func (session *Session) getNodeForLocation(x, y float64) (int64, error) {
	msg, err := session.blockingSend("DOM.getNodeForLocation", &Params{
		"x":                         x,
		"y":                         y,
		"includeUserAgentShadowDOM": true,
	})
	if err != nil {
		return 0, err
	}
	nbf := msg["backendNodeId"].(float64)
	return int64(nbf), nil
}

func (session *Session) describeNode(objectID string) (*Node, error) {
	raw, err := session.blockingSend("DOM.describeNode", &Params{
		"objectId": objectID,
		"depth":    1,
	})
	if err != nil {
		return nil, err
	}
	node := &Node{}
	unmarshal(raw["node"], node)
	return node, nil
}

func (session *Session) requestNode(objectID string) (int64, error) {
	msg, err := session.blockingSend("DOM.requestNode", &Params{
		"objectId": objectID,
	})
	if err != nil {
		return -1, err
	}
	return int64(msg["nodeId"].(float64)), nil
}

func (session *Session) resolveNode(backendNodeID int64) (*RemoteObject, error) {
	raw, err := session.blockingSend("DOM.resolveNode", &Params{
		"backendNodeId":      backendNodeID,
		"executionContextId": session.contextID,
	})
	if err != nil {
		return nil, err
	}
	ro := &RemoteObject{}
	unmarshal(raw["object"], ro)
	return ro, nil
}

func (session *Session) highlightNode(objectID string) error {
	_, err := session.blockingSend("Overlay.highlightNode", &Params{
		"objectId": objectID,
		"highlightConfig": &highlightConfig{
			ShowRulers:  true,
			BorderColor: &rgba{R: 255, G: 1, B: 1},
		},
	})
	return err
}
func (session *Session) highlightQuad(q quad, color *rgba) error {
	_, err := session.blockingSend("Overlay.highlightQuad", &Params{
		"quad":         q,
		"outlineColor": color,
	})
	return err
}

func (session *Session) getContentQuads(backendNodeID int64, objectID string) (quad, error) {
	params := &Params{}
	if backendNodeID > 0 {
		(*params)["backendNodeId"] = backendNodeID
	}
	if objectID != "" {
		(*params)["objectId"] = objectID
	}
	obj, err := session.blockingSend("DOM.getContentQuads", params)
	if err != nil {
		return nil, err
	}
	qs := make([]quad, 0)
	if len(qs) > 1 {
		panic("todo few quads support")
	}
	unmarshal(obj["quads"], &qs)
	return qs[0], nil
}

func (session *Session) setFileInputFiles(files []string, objectID string) error {
	_, err := session.blockingSend("DOM.setFileInputFiles", &Params{
		"files":    files,
		"objectId": objectID,
	})
	return err
}

func (session *Session) getAttributes(nodeID int64) ([]string, error) {
	msg, err := session.blockingSend("DOM.getAttributes", &Params{
		"nodeId": nodeID,
	})
	if err != nil {
		return nil, err
	}
	attributes := make([]string, 0)
	unmarshal(msg["attributes"], &attributes)
	return attributes, err
}
