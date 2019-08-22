package cdp

// ErrorReason https://chromedevtools.github.io/devtools-protocol/tot/Network#type-ErrorReason
type ErrorReason string

// error reasons
const (
	Failed               ErrorReason = "Failed"
	Aborted              ErrorReason = "Aborted"
	TimedOut             ErrorReason = "TimedOut"
	AccessDenied         ErrorReason = "AccessDenied"
	ConnectionClosed     ErrorReason = "ConnectionClosed"
	ConnectionReset      ErrorReason = "ConnectionReset"
	ConnectionRefused    ErrorReason = "ConnectionRefused"
	ConnectionAborted    ErrorReason = "ConnectionAborted"
	ConnectionFailed     ErrorReason = "ConnectionFailed"
	NameNotResolved      ErrorReason = "NameNotResolved"
	InternetDisconnected ErrorReason = "InternetDisconnected"
	AddressUnreachable   ErrorReason = "AddressUnreachable"
	BlockedByClient      ErrorReason = "BlockedByClient"
	BlockedByResponse    ErrorReason = "BlockedByResponse"
)

// HeaderEntry https://chromedevtools.github.io/devtools-protocol/tot/Fetch#type-HeaderEntry
type HeaderEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// RequestPattern https://chromedevtools.github.io/devtools-protocol/tot/Fetch#type-RequestPattern
type RequestPattern struct {
	URLPattern   string `json:"urlPattern"`
	ResourceType string `json:"resourceType"`
	RequestStage string `json:"requestStage"`
}

// RequestPaused RequestPaused
type RequestPaused struct {
	RequestID           string         `json:"requestId"`
	Request             *Request       `json:"request"`
	FrameID             string         `json:"frameId"`
	ResponseErrorReason *ErrorReason   `json:"responseErrorReason,omitempty"`
	ResponseStatusCode  int            `json:"responseStatusCode,omitempty"`
	ResponseHeaders     []*HeaderEntry `json:"responseHeaders,omitempty"`
	NetworkID           string         `json:"networkId,omitempty"`
}

// fetchEnable https://chromedevtools.github.io/devtools-protocol/tot/Fetch#method-enable
func (session *Session) fetchEnable(patterns []*RequestPattern, handleAuthRequests bool) error {
	_, err := session.blockingSend("Fetch.enable", &Params{
		"patterns":           patterns,
		"handleAuthRequests": handleAuthRequests,
	})
	return err
}

// fetchDisable https://chromedevtools.github.io/devtools-protocol/tot/Fetch#method-disable
func (session *Session) fetchDisable() error {
	_, err := session.blockingSend("Fetch.enable", &Params{})
	return err
}

// failRequest https://chromedevtools.github.io/devtools-protocol/tot/Fetch#method-failRequest
func (session *Session) failRequest(requestID string, reason ErrorReason) error {
	_, err := session.blockingSend("Fetch.failRequest", &Params{
		"requestId":   requestID,
		"errorReason": string(reason),
	})
	return err
}

// fulfillRequest https://chromedevtools.github.io/devtools-protocol/tot/Fetch#method-fulfillRequest
func (session *Session) fulfillRequest(requestID string, responseCode int, responseHeaders []*HeaderEntry, body *string, responsePhrase *string) error {
	p := Params{
		"requestId":       requestID,
		"responseCode":    responseCode,
		"responseHeaders": responseHeaders,
	}
	if body != nil {
		p["body"] = body
	}
	if responsePhrase != nil {
		p["responsePhrase"] = responsePhrase
	}
	_, err := session.blockingSend("Fetch.fulfillRequest", &p)
	return err
}

// continueRequest https://chromedevtools.github.io/devtools-protocol/tot/Fetch#method-continueRequest
func (session *Session) continueRequest(requestID string, url *string, method *string, postData *string, headers []*HeaderEntry) error {
	p := Params{
		"requestId": requestID,
	}
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
	_, err := session.blockingSend("Fetch.continueRequest", &p)
	return err
}

// Proceed continue paused request
type Proceed struct {
	Fail     func(requestID string, reason ErrorReason) error
	Fulfill  func(requestID string, responseCode int, responseHeaders []*HeaderEntry, body *string, responsePhrase *string) error
	Continue func(requestID string, url *string, method *string, postData *string, headers []*HeaderEntry) error
}

// Fetch ...
func (session *Session) Fetch(patterns []*RequestPattern, fn func(*RequestPaused, *Proceed)) func() {
	proceed := &Proceed{
		Fail:     session.failRequest,
		Fulfill:  session.fulfillRequest,
		Continue: session.continueRequest,
	}
	unsubscribe := session.Subscribe("Fetch.requestPaused", func(msg Params) {
		go func() {
			request := &RequestPaused{}
			unmarshal(msg, request)
			fn(request, proceed)
		}()
	})
	if err := session.fetchEnable(patterns, false); err != nil {
		panic(err)
	}
	return func() {
		unsubscribe()
		if err := session.fetchDisable(); err != nil {
			panic(err)
		}
	}
}
