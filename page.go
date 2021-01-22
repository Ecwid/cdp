package cdp

import (
	"strings"

	"github.com/ecwid/cdp/pkg/devtool"
)

const blankPage = "about:blank"

func (session *Session) setFrame(frameID string) error {
	if _, ok := session.frames.Load(frameID); !ok {
		return ErrNoSuchFrame
	}
	session.frameID = frameID
	return nil
}

func (page Page) getNavigationHistory() (*devtool.NavigationHistory, error) {
	history := new(devtool.NavigationHistory)
	err := page.call("Page.getNavigationHistory", nil, history)
	return history, err
}

func (page Page) navigateToHistoryEntry(entryID int64) error {
	return page.call("Page.navigateToHistoryEntry", Map{"entryId": entryID}, nil)
}

func (page Page) getLayoutMetrics() (*devtool.LayoutMetrics, error) {
	metrics := new(devtool.LayoutMetrics)
	err := page.call("Page.getLayoutMetrics", nil, metrics)
	return metrics, err
}

func (page Page) getFrameTree() (*devtool.FrameTree, error) {
	tree := new(devtool.FrameTreeResult)
	if err := page.call("Page.getFrameTree", nil, tree); err != nil {
		return nil, err
	}
	return tree.FrameTree, nil
}

// Activate activate current Target
func (page Page) activate(targetID string) error {
	return page.call("Target.activateTarget", Map{"targetId": targetID}, nil)
}

// AddScriptToEvaluateOnNewDocument https://chromedevtools.github.io/devtools-protocol/tot/Page#method-addScriptToEvaluateOnNewDocument
func (page Page) AddScriptToEvaluateOnNewDocument(source string) (string, error) {
	var result = Map{}
	err := page.call("Page.addScriptToEvaluateOnNewDocument", Map{"source": source}, &result)
	if err != nil {
		return "", err
	}
	return result["identifier"].(string), nil
}

// RemoveScriptToEvaluateOnNewDocument https://chromedevtools.github.io/devtools-protocol/tot/Page#method-removeScriptToEvaluateOnNewDocument
func (page Page) RemoveScriptToEvaluateOnNewDocument(identifier string) error {
	return page.call("Page.removeScriptToEvaluateOnNewDocument", Map{"identifier": identifier}, nil)
}

// SetDownloadBehavior https://chromedevtools.github.io/devtools-protocol/tot/Page#method-setDownloadBehavior
func (page Page) SetDownloadBehavior(behavior devtool.DownloadBehavior, downloadPath string) error {
	return page.call("Page.setDownloadBehavior", Map{
		"behavior":     string(behavior),
		"downloadPath": downloadPath,
	}, nil)
}

func (session Session) query(parent *Element, selector string) (*Element, error) {
	selector = strings.ReplaceAll(selector, `"`, `\"`)
	var (
		e       *devtool.RemoteObject
		context int64
		err     error
	)
	if parent == nil {
		if context, err = session.executionContext(); err != nil {
			return nil, err
		}
		e, err = session.evaluate(`document.querySelector("`+selector+`")`, context, false, false)
	} else {
		e, err = parent.call(`function(s){return this.querySelector(s)}`, selector)
	}
	if err != nil {
		return nil, err
	}
	if e.ObjectID == "" {
		return nil, NoSuchElementError{selector: selector, context: context, frame: session.frameID}
	}
	return newElement(&session, parent, e.ObjectID)
}

func (session Session) queryAll(parent *Element, selector string) ([]*Element, error) {
	selector = strings.ReplaceAll(selector, `"`, `\"`)
	var (
		array   *devtool.RemoteObject
		context int64
		err     error
	)
	if parent == nil {
		if context, err = session.executionContext(); err != nil {
			return nil, err
		}
		array, err = session.evaluate(`document.querySelectorAll("`+selector+`")`, context, false, false)
	} else {
		array, err = parent.call(`function(s){return this.querySelectorAll(s)}`, selector)
	}
	if err != nil {
		return nil, err
	}
	if array == nil || array.Description == "NodeList(0)" {
		_ = session.releaseObject(array.ObjectID)
		return nil, NoSuchElementError{selector: selector, context: context, frame: session.frameID}
	}
	all := make([]*Element, 0)
	descriptor, err := session.getProperties(array.ObjectID)
	if err != nil {
		return nil, err
	}
	for _, d := range descriptor {
		if !d.Enumerable {
			continue
		}
		e, err := newElement(&session, parent, d.Value.ObjectID)
		if err != nil {
			return nil, err
		}
		all = append(all, e)
	}
	return all, nil
}
