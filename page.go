package cdp

import (
	"encoding/base64"
)

// Frame https://chromedevtools.github.io/devtools-protocol/tot/Page#type-Frame
type Frame struct {
	ID             string `json:"id"`
	ParentID       string `json:"parentId"`
	LoaderID       string `json:"loaderId"`
	Name           string `json:"name"`
	URL            string `json:"url"`
	SecurityOrigin string `json:"securityOrigin"`
	MimeType       string `json:"mimeTypeurl"`
	UnreachableURL string `json:"unreachableUrl"`
}

// FrameTree https://chromedevtools.github.io/devtools-protocol/tot/Page#type-FrameTree
type FrameTree struct {
	Frame       *Frame       `json:"frame"`
	ChildFrames []*FrameTree `json:"childFrames"`
}

type NavigationResult struct {
	FrameID   string `json:"frameId"`
	LoaderID  string `json:"loaderId"`
	ErrorText string `json:"errorText"`
}

type RequestedNavigation struct {
	FrameID string `json:"frameId"`
	Reason  string `json:"reason"`
	URL     string `json:"url"`
}

type ScreencastFrameMetadata struct {
	OffsetTop       float64 `json:"offsetTop"`
	PageScaleFactor float64 `json:"pageScaleFactor"`
	DeviceWidth     float64 `json:"deviceWidth"`
	DeviceHeight    float64 `json:"deviceHeight"`
	ScrollOffsetX   float64 `json:"scrollOffsetX"`
	ScrollOffsetY   float64 `json:"scrollOffsetY"`
	Timestamp       float64 `json:"timestamp"`
}

// NavigationEntry https://chromedevtools.github.io/devtools-protocol/tot/Page#type-NavigationEntry
type NavigationEntry struct {
	ID             int64  `json:"id"`
	URL            string `json:"url"`
	UserTypedURL   string `json:"userTypedURL"`
	Title          string `json:"title"`
	TransitionType string `json:"transitionType"`
}

// LayoutViewport https://chromedevtools.github.io/devtools-protocol/tot/Page#type-LayoutViewport
type LayoutViewport struct {
	PageX        int64 `json:"pageX"`
	PageY        int64 `json:"pageY"`
	ClientWidth  int64 `json:"clientWidth"`
	ClientHeight int64 `json:"clientHeight"`
}

// VisualViewport https://chromedevtools.github.io/devtools-protocol/tot/Page#type-VisualViewport
type VisualViewport struct {
	OffsetX      float64 `json:"offsetX"`
	OffsetY      float64 `json:"offsetY"`
	PageX        float64 `json:"pageX"`
	PageY        float64 `json:"pageY"`
	ClientWidth  float64 `json:"clientWidth"`
	ClientHeight float64 `json:"clientHeight"`
	Scale        float64 `json:"scale"`
	Zoom         float64 `json:"zoom"`
}

// LayoutMetrics https://chromedevtools.github.io/devtools-protocol/tot/Page#method-getLayoutMetrics
type LayoutMetrics struct {
	LayoutViewport *LayoutMetrics  `json:"layoutViewport"`
	VisualViewport *VisualViewport `json:"visualViewport"`
	ContentSize    *Rect           `json:"contentSize"`
}

type navigationHistory struct {
	CurrentIndex int64              `json:"currentIndex"`
	Entries      []*NavigationEntry `json:"entries"`
}

type lifecycleEvent struct {
	FrameID   string  `json:"frameId"`
	LoaderID  string  `json:"loaderId"`
	Name      string  `json:"name"`
	Timestamp float64 `json:"timestamp"`
}

func (session *Session) getFrameTree() (*FrameTree, error) {
	obj, err := session.blockingSend("Page.getFrameTree", &Params{})
	if err != nil {
		return nil, err
	}
	tree := &FrameTree{}
	unmarshal(obj["frameTree"], tree)
	return tree, nil
}

func (tree *FrameTree) get(id string) *Frame {
	if tree.Frame.ID == id {
		return tree.Frame
	}
	for _, f := range tree.ChildFrames {
		if child := f.get(id); child != nil {
			return child
		}
	}
	return nil
}

func (session *Session) setLifecycleEventsEnabled(enabled bool) error {
	_, err := session.blockingSend("Page.setLifecycleEventsEnabled", &Params{"enabled": enabled})
	return err
}

func (session *Session) getLayoutMetrics() (*LayoutMetrics, error) {
	msg, err := session.blockingSend("Page.getLayoutMetrics", &Params{})
	if err != nil {
		return nil, err
	}
	lm := &LayoutMetrics{}
	unmarshal(msg, lm)
	return lm, nil
}

func (session *Session) startScreencast(format string, quality int8, maxWidth int64, maxHeight int64, everyNthFrame int64) error {
	_, err := session.blockingSend("Page.startScreencast", &Params{
		"format":        format,
		"quality":       quality,
		"maxWidth":      maxWidth,
		"maxHeight":     maxHeight,
		"everyNthFrame": everyNthFrame,
	})
	return err
}

func (session *Session) stopScreencast() error {
	_, err := session.blockingSend("Page.stopScreencast", &Params{})
	return err
}

func (session *Session) screencastFrameAck() error {
	_, err := session.blockingSend("Page.screencastFrameAck", &Params{})
	return err
}

func (session *Session) navigate(url string, frameID string) (*NavigationResult, error) {
	msg, err := session.blockingSend("Page.navigate", &Params{
		"url":            url,
		"transitionType": "typed",
		"frameId":        frameID,
	})
	if err != nil {
		return nil, err
	}
	nav := &NavigationResult{}
	unmarshal(msg, nav)
	return nav, nil
}

func (session *Session) reload() error {
	_, err := session.blockingSend("Page.reload", &Params{
		"ignoreCache": true,
	})
	return err
}

func (session *Session) createIsolatedWorld(frameID string) (executionContextID int64, err error) {
	msg, err := session.blockingSend("Page.createIsolatedWorld", &Params{
		"frameId":             frameID,
		"name":                "__utilityWorld__",
		"grantUniveralAccess": true,
	})
	if err != nil {
		return 0, err
	}
	id := msg["executionContextId"].(float64)
	return int64(id), nil
}

func (session *Session) getNavigationHistory() (*navigationHistory, error) {
	msg, err := session.blockingSend("Page.getNavigationHistory", &Params{})
	if err != nil {
		return nil, err
	}
	history := &navigationHistory{}
	unmarshal(msg, history)
	return history, nil
}

func (session *Session) captureScreenshot(format string, quality int8) ([]byte, error) {
	msg, err := session.blockingSend("Page.captureScreenshot", &Params{
		"format":      format,
		"quality":     quality,
		"fromSurface": true,
	})
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(msg["data"].(string))
}
