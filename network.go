package cdp

// Request https://chromedevtools.github.io/devtools-protocol/tot/Network#type-Request
type Request struct {
	URL              string                 `json:"url"`
	URLFragment      string                 `json:"urlFragment"`
	Method           string                 `json:"method"`
	Headers          map[string]interface{} `json:"headers"`
	PostData         string                 `json:"postData"`
	HasPostData      bool                   `json:"hasPostData"`
	MixedContentType string                 `json:"mixedContentType"`
	InitialPriority  string                 `json:"initialPriority"`
	ReferrerPolicy   string                 `json:"referrerPolicy"`
	IsLinkPreload    bool                   `json:"isLinkPreload"`
}

// RequestWillBeSent https://chromedevtools.github.io/devtools-protocol/tot/Network#event-requestWillBeSent
type RequestWillBeSent struct {
	RequestID        string    `json:"requestId"`
	LoaderID         string    `json:"loaderId"`
	DocumentURL      string    `json:"documentURL"`
	Request          *Request  `json:"request"`
	Timestamp        float64   `json:"timestamp"`
	WallTime         float64   `json:"wallTime"`
	RedirectResponse *Response `json:"redirectResponse"`
	Type             string    `json:"type"`
	FrameID          string    `json:"frameId"`
	HasUserGesture   bool      `json:"hasUserGesture"`
}

// ResponseReceived https://chromedevtools.github.io/devtools-protocol/tot/Network#event-responseReceived
type ResponseReceived struct {
	RequestID string    `json:"requestId"`
	LoaderID  string    `json:"loaderId"`
	Timestamp float64   `json:"timestamp"`
	Type      string    `json:"type"`
	Response  *Response `json:"response"`
	FrameID   string    `json:"frameId"`
}

// DataReceived https://chromedevtools.github.io/devtools-protocol/tot/Network#event-dataReceived
type DataReceived struct {
	RequestID         string  `json:"requestId"`
	Timestamp         float64 `json:"timestamp"`
	DataLength        int64   `json:"dataLength"`
	EncodedDataLength int64   `json:"encodedDataLength"`
}

// LoadingFinished https://chromedevtools.github.io/devtools-protocol/tot/Network#event-loadingFinished
type LoadingFinished struct {
	RequestID                string  `json:"requestId"`
	Timestamp                float64 `json:"timestamp"`
	EncodedDataLength        float64 `json:"encodedDataLength"`
	ShouldReportCorbBlocking bool    `json:"shouldReportCorbBlocking"`
}

// ServedFromCache https://chromedevtools.github.io/devtools-protocol/tot/Network#event-requestServedFromCache
type ServedFromCache struct {
	RequestID string `json:"requestId"`
}

// PageLoadEventFired https://chromedevtools.github.io/devtools-protocol/tot/Page#event-loadEventFired
type PageLoadEventFired struct {
	Timestamp float64 `json:"timestamp"`
}

// PageDomContentEventFired https://chromedevtools.github.io/devtools-protocol/tot/Page#event-domContentEventFired
type PageDomContentEventFired struct {
	Timestamp float64 `json:"timestamp"`
}

// ResourceTiming https://chromedevtools.github.io/devtools-protocol/tot/Network#type-ResourceTiming
type ResourceTiming struct {
	RequestTime       float64 `json:"requestTime"`
	ProxyStart        float64 `json:"proxyStart"`
	ProxyEnd          float64 `json:"proxyEnd"`
	DNSStart          float64 `json:"dnsStart"`
	DNSEnd            float64 `json:"dnsEnd"`
	ConnectStart      float64 `json:"connectStart"`
	ConnectEnd        float64 `json:"connectEnd"`
	SSLStart          float64 `json:"sslStart"`
	SSLEnd            float64 `json:"sslEnd"`
	WorkerStart       float64 `json:"workerStart"`
	WorkerReady       float64 `json:"workerReady"`
	SendStart         float64 `json:"sendStart"`
	SendEnd           float64 `json:"sendEnd"`
	PushStart         float64 `json:"pushStart"`
	PushEnd           float64 `json:"pushEnd"`
	ReceiveHeadersEnd float64 `json:"receiveHeadersEnd"`
}

// Response https://chromedevtools.github.io/devtools-protocol/tot/Network#type-Response
type Response struct {
	URL                string                 `json:"url"`
	Status             int                    `json:"status"`
	StatusText         string                 `json:"statusText"`
	Headers            map[string]interface{} `json:"headers"`
	HeadersText        string                 `json:"headersText"`
	MimeType           string                 `json:"mimeType"`
	RequestHeaders     map[string]interface{} `json:"requestHeaders"`
	RequestHeadersText string                 `json:"requestHeadersText"`
	ConnectionReused   bool                   `json:"connectionReused"`
	ConnectionID       int64                  `json:"connectionId"`
	RemoteIPAddress    string                 `json:"remoteIPAddress"`
	RemotePort         int64                  `json:"remotePort"`
	FromDiskCache      bool                   `json:"fromDiskCache"`
	FromServiceWorker  bool                   `json:"fromServiceWorker"`
	FromPrefetchCache  bool                   `json:"fromPrefetchCache"`
	EncodedDataLength  int64                  `json:"encodedDataLength"`
	Timing             *ResourceTiming        `json:"timing"`
	Protocol           string                 `json:"protocol"`
	SecurityState      string                 `json:"securityState"`
}

// LoadingFailed https://chromedevtools.github.io/devtools-protocol/tot/Network#event-loadingFailed
type LoadingFailed struct {
	RequestID     string  `json:"requestId"`
	Timestamp     float64 `json:"timestamp"`
	Type          string  `json:"type"`
	ErrorText     string  `json:"errorText"`
	Canceled      bool    `json:"canceled"`
	BlockedReason string  `json:"blockedReason"`
}

// CookieParam https://chromedevtools.github.io/devtools-protocol/tot/Network#type-CookieParam
type CookieParam struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	URL      string `json:"url"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  int64  `json:"expires"`
	Size     int64  `json:"size"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

// NetworkEnable ...
func (session *Session) NetworkEnable() error {
	_, err := session.blockingSend("Network.enable", &Params{
		"maxPostDataSize": 1024,
	})
	return err
}

// ClearBrowserCookies ...
func (session *Session) ClearBrowserCookies() error {
	_, err := session.blockingSend("Network.clearBrowserCookies", &Params{})
	return err
}

// SetCookies ...
func (session *Session) SetCookies(cookies ...CookieParam) error {
	_, err := session.blockingSend("Network.setCookies", &Params{"cookies": cookies})
	return err
}

// SetExtraHTTPHeaders Specifies whether to always send extra HTTP headers with the requests from this page.
func (session *Session) SetExtraHTTPHeaders(headers map[string]string) error {
	_, err := session.blockingSend("Network.setExtraHTTPHeaders", &Params{"headers": headers})
	return err
}
