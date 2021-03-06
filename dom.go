package cdp

import (
	"math"

	"github.com/ecwid/cdp/pkg/devtool"
)

// GetNode ...
func (session DOM) GetNode(objectID string) (*devtool.Node, error) {
	p := Map{
		"objectId": objectID,
		"depth":    1,
	}
	describeNode := new(devtool.DescribeNode)
	if err := session.call("DOM.describeNode", p, describeNode); err != nil {
		return nil, err
	}
	return describeNode.Node, nil
}

func (session DOM) scrollIntoViewIfNeeded(objectID string) error {
	return session.call("DOM.scrollIntoViewIfNeeded", Map{"objectId": objectID}, nil)
}

// GetContentQuads ...
func (session DOM) GetContentQuads(objectID string, viewportCorrection bool) (devtool.Quad, error) {
	cq := new(devtool.ContentQuads)
	if err := session.call("DOM.getContentQuads", Map{"objectId": objectID}, cq); err != nil {
		return nil, err
	}
	calc := cq.Calc()
	if len(calc) == 0 { // should be at least one
		return nil, ErrElementInvisible
	}
	metric, err := session.GetLayoutMetrics()
	if err != nil {
		return nil, err
	}
	for _, quad := range calc {
		/* correction is get sub-quad of element that in viewport
		 _______________  <- Viewport top
		|  1 _______ 2  |
		|   |visible|   | visible part of element
		|__4|visible|3__| <- Viewport bottom
		|   |invisib|   | this invisible part of element omits if viewportCorrection
		|...............|
		*/
		if viewportCorrection {
			for _, point := range quad {
				point.X = math.Min(math.Max(point.X, 0), float64(metric.LayoutViewport.ClientWidth))
				point.Y = math.Min(math.Max(point.Y, 0), float64(metric.LayoutViewport.ClientHeight))
			}
		}
		if quad.Area() > 1 {
			return quad, nil
		}
	}
	return nil, ErrElementIsOutOfViewport
}
