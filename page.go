package cdp

import (
	"strings"

	"github.com/ecwid/cdp/pkg/devtool"
)

const blankPage = "about:blank"

func (session Page) getNavigationHistory() (*devtool.NavigationHistory, error) {
	history := new(devtool.NavigationHistory)
	err := session.call("Page.getNavigationHistory", nil, history)
	return history, err
}

func (session Page) createIsolatedWorld(frameID, worldName string) (int64, error) {
	result := Map{}
	if err := session.call("Page.createIsolatedWorld", Map{"frameId": frameID, "worldName": worldName}, &result); err != nil {
		return 0, err
	}
	return int64(result["executionContextId"].(float64)), nil
}

func (session Page) navigateToHistoryEntry(entryID int64) error {
	return session.call("Page.navigateToHistoryEntry", Map{"entryId": entryID}, nil)
}

// GetLayoutMetrics ...
func (session Page) GetLayoutMetrics() (*devtool.LayoutMetrics, error) {
	metrics := new(devtool.LayoutMetrics)
	err := session.call("Page.getLayoutMetrics", nil, metrics)
	return metrics, err
}

// HandleJavaScriptDialog ...
func (session Page) HandleJavaScriptDialog(accept bool, promptText string) error {
	return session.call("Page.handleJavaScriptDialog", Map{"accept": accept, "promptText": promptText}, nil)
}

// GetFrameTree ...
func (session Page) GetFrameTree() (*devtool.FrameTree, error) {
	tree := new(devtool.FrameTreeResult)
	if err := session.call("Page.getFrameTree", nil, tree); err != nil {
		return nil, err
	}
	return tree.FrameTree, nil
}

// Activate activate current Target
func (session Page) activate(targetID string) error {
	return session.call("Target.activateTarget", Map{"targetId": targetID}, nil)
}

// AddScriptToEvaluateOnNewDocument https://chromedevtools.github.io/devtools-protocol/tot/Page#method-addScriptToEvaluateOnNewDocument
func (session Page) AddScriptToEvaluateOnNewDocument(source string) (string, error) {
	var result = Map{}
	err := session.call("Page.addScriptToEvaluateOnNewDocument", Map{"source": source}, &result)
	if err != nil {
		return "", err
	}
	return result["identifier"].(string), nil
}

// RemoveScriptToEvaluateOnNewDocument https://chromedevtools.github.io/devtools-protocol/tot/Page#method-removeScriptToEvaluateOnNewDocument
func (session Page) RemoveScriptToEvaluateOnNewDocument(identifier string) error {
	return session.call("Page.removeScriptToEvaluateOnNewDocument", Map{"identifier": identifier}, nil)
}

// SetDownloadBehavior https://chromedevtools.github.io/devtools-protocol/tot/Page#method-setDownloadBehavior
func (session Page) SetDownloadBehavior(behavior devtool.DownloadBehavior, downloadPath string) error {
	return session.call("Page.setDownloadBehavior", Map{
		"behavior":     string(behavior),
		"downloadPath": downloadPath,
	}, nil)
}

func (session Session) query(parent *Element, selector string) (*Element, error) {
	selector = strings.ReplaceAll(selector, `"`, `\"`)
	var (
		err     error
		e       *devtool.RemoteObject
		context = session.currentContext()
	)
	if parent == nil {
		e, err = session.evaluate(`document.querySelector("`+selector+`")`, context, false, false)
	} else {
		e, err = parent.call(`function(s){return this.querySelector(s)}`, selector)
	}
	if err != nil {
		return nil, err
	}
	if e.ObjectID == "" {
		return nil, NoSuchElementError{selector: selector, context: context}
	}
	return newElement(&session, parent, e), nil
}

func (session Session) queryAll(parent *Element, selector string) ([]*Element, error) {
	selector = strings.ReplaceAll(selector, `"`, `\"`)
	var (
		err     error
		array   *devtool.RemoteObject
		context = session.currentContext()
	)
	if parent == nil {
		array, err = session.evaluate(`document.querySelectorAll("`+selector+`")`, context, false, false)
	} else {
		array, err = parent.call(`function(s){return this.querySelectorAll(s)}`, selector)
	}
	if err != nil {
		return nil, err
	}
	if array == nil || array.Description == "NodeList(0)" {
		_ = session.releaseObject(array.ObjectID)
		return nil, NoSuchElementError{selector: selector, context: context}
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
		all = append(all, newElement(&session, parent, d.Value))
	}
	return all, nil
}
