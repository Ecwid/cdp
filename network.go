package cdp

import (
	"encoding/base64"
	"encoding/json"

	"github.com/ecwid/cdp/pkg/devtool"
)

// ClearBrowserCookies ...
func (net Network) ClearBrowserCookies() error {
	return net.call("Network.clearBrowserCookies", nil, nil)
}

// SetCookies ...
func (net Network) SetCookies(cookies ...*devtool.Cookie) error {
	return net.call("Network.setCookies", Map{"cookies": cookies}, nil)
}

// GetCookies returns all browser cookies for the current URL
func (net Network) GetCookies(urls ...string) ([]*devtool.Cookie, error) {
	p := Map{}
	if urls != nil {
		p["urls"] = urls
	}
	cookies := new(devtool.GetCookies)
	err := net.call("Network.getCookies", p, cookies)
	if err != nil {
		return nil, err
	}
	return cookies.Cookies, nil
}

// SetExtraHTTPHeaders Specifies whether to always send extra HTTP headers with the requests from this page.
func (net Network) SetExtraHTTPHeaders(headers map[string]string) error {
	return net.call("Network.setExtraHTTPHeaders", Map{"headers": headers}, nil)
}

// SetOffline set offline/online mode
// SetOffline(false) - reset all network conditions to default
func (net Network) SetOffline(e bool) error {
	return net.emulateNetworkConditions(e, 0, -1, -1)
}

// SetThrottling set latency in milliseconds, download & upload throttling in bytes per second
func (net Network) SetThrottling(latencyMs, downloadThroughputBps, uploadThroughputBps int) error {
	return net.emulateNetworkConditions(false, latencyMs, downloadThroughputBps, downloadThroughputBps)
}

// SetBlockedURLs ...
func (net Network) SetBlockedURLs(urls []string) error {
	return net.call("Network.setBlockedURLs", Map{"urls": urls}, nil)
}

// GetRequestPostData https://chromedevtools.github.io/devtools-protocol/tot/Network/#method-getRequestPostData
func (net Network) GetRequestPostData(requestID string) (string, error) {
	result := new(devtool.RequestPostData)
	err := net.call("Network.getRequestPostData", Map{"requestId": requestID}, result)
	if err != nil {
		return "", err
	}
	return result.PostData, nil
}

// GetResponseBody https://chromedevtools.github.io/devtools-protocol/tot/Network/#method-getResponseBody
func (net Network) GetResponseBody(requestID string) (string, error) {
	result := new(devtool.ResponseBody)
	err := net.call("Network.getResponseBody", Map{"requestId": requestID}, result)
	if err != nil {
		return "", err
	}
	if result.Base64Encoded {
		b, err := base64.StdEncoding.DecodeString(result.Body)
		return string(b), err
	}
	return result.Body, nil
}

func (net Network) emulateNetworkConditions(offline bool, latencyMs, downloadThroughputBps, uploadThroughputBps int) error {
	p := Map{
		"offline":            offline,
		"latency":            latencyMs,
		"downloadThroughput": downloadThroughputBps,
		"uploadThroughput":   uploadThroughputBps,
	}
	return net.call("Network.emulateNetworkConditions", p, nil)
}

// fetchEnable https://chromedevtools.github.io/devtools-protocol/tot/Fetch#method-enable
func (net Network) fetchEnable(patterns []*devtool.RequestPattern, handleAuthRequests bool) error {
	return net.call("Fetch.enable", Map{
		"patterns":           patterns,
		"handleAuthRequests": handleAuthRequests,
	}, nil)
}

// fetchDisable https://chromedevtools.github.io/devtools-protocol/tot/Fetch#method-disable
func (net Network) fetchDisable() error {
	return net.call("Fetch.disable", nil, nil)
}

// Fail https://chromedevtools.github.io/devtools-protocol/tot/Fetch#method-failRequest
func (net Interceptor) Fail(requestID string, reason devtool.ErrorReason) error {
	return net.call("Fetch.failRequest", Map{
		"requestId":   requestID,
		"errorReason": string(reason),
	}, nil)
}

// Fulfill https://chromedevtools.github.io/devtools-protocol/tot/Fetch#method-fulfillRequest
func (net Interceptor) Fulfill(
	requestID string,
	responseCode int,
	responseHeaders []*devtool.HeaderEntry,
	body *string,
	responsePhrase *string) error {
	p := Map{
		"requestId":    requestID,
		"responseCode": responseCode,
	}
	if responseHeaders != nil {
		p["responseHeaders"] = responseHeaders
	}
	if body != nil {
		p["body"] = body
	}
	if responsePhrase != nil {
		p["responsePhrase"] = responsePhrase
	}
	return net.call("Fetch.fulfillRequest", p, nil)
}

// Continue https://chromedevtools.github.io/devtools-protocol/tot/Fetch#method-continueRequest
func (net Interceptor) Continue(requestID string, url *string, method *string, postData *string, headers []*devtool.HeaderEntry) error {
	p := Map{"requestId": requestID}
	if url != nil {
		p["url"] = url
	}
	if method != nil {
		p["method"] = method
	}
	if postData != nil {
		p["postData"] = postData
	}
	if headers != nil {
		p["headers"] = headers
	}
	return net.call("Fetch.continueRequest", p, nil)
}

// Interceptor ...
type Interceptor struct {
	*Network
}

// Intercept ...
func (net Network) Intercept(patterns []*devtool.RequestPattern, fn func(*devtool.RequestPaused, *Interceptor)) func() {
	unsubscribe := net.Subscribe("Fetch.requestPaused", func(e *Event) {
		request := new(devtool.RequestPaused)
		if err := json.Unmarshal(e.Params, request); err != nil {
			net.close(err)
			return
		}
		go fn(request, &Interceptor{Network: &net})
	})
	if err := net.fetchEnable(patterns, false); err != nil {
		net.close(err)
	}
	return func() {
		unsubscribe()
		if err := net.fetchDisable(); err != nil {
			net.close(err)
		}
	}
}
