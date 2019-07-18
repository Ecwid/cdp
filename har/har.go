package har

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/ecwid/cdp"
)

var errPoorRequest = fmt.Errorf("got a response event with no matching request event")

// Recorder har recorder
type Recorder struct {
	Log  *HAR   `json:"log"`
	Stop func() `json:"-"`
}

// NewRecorder new har
func NewRecorder(session *cdp.Session) *Recorder {
	session.NetworkEnable()
	har := &HAR{
		Version: "1.2",
		Creator: &Creator{
			Name:    "ecwid-cdp",
			Version: "0.1",
		},
		Pages:   make([]*Page, 0),
		Entries: make([]*Entry, 0),
	}

	events := map[string]func(cdp.Params){
		"Page.frameStartedLoading":       har.frameStartedLoading,
		"Page.navigatedWithinDocument":   har.frameStartedLoading,
		"Page.loadEventFired":            har.loadEventFired,
		"Page.domContentEventFired":      har.domContentEventFired,
		"Network.requestWillBeSent":      har.requestWillBeSent,
		"Network.responseReceived":       har.responseReceived,
		"Network.dataReceived":           har.dataReceived,
		"Network.loadingFinished":        har.loadingFinished,
		"Network.requestServedFromCache": har.requestServedFromCache,
		"Network.loadingFailed":          har.loadingFailed,
	}

	unsubscribes := make([]func(), 0)
	for method, event := range events {
		unsubscribes = append(unsubscribes, session.Subscribe(method, event))
	}

	return &Recorder{
		Log: har,
		Stop: func() {
			for _, un := range unsubscribes {
				un()
			}
		},
	}

}

// GetRequest find Request by ID
func (rec *Recorder) GetRequest(requestID string) *Request {
	for _, e := range rec.Log.Entries {
		if e.Request.ID == requestID {
			return e.Request
		}
	}
	return nil
}

func (har *HAR) frameStartedLoading(parameters cdp.Params) {
	page := &Page{
		StartedDateTime: time.Now(),
		PageTiming:      &PageTiming{},
	}
	har.Pages = append(har.Pages, page)
	page.ID = har.harPageID()
}

func (har *HAR) requestWillBeSent(parameters cdp.Params) {
	if len(har.Pages) < 1 {
		// log.Print("sending request object, but frame not started")
		return
	}
	willBeSent := &cdp.RequestWillBeSent{}
	b, err := json.Marshal(parameters)
	if err != nil {
		// log.Print(err)
		return
	}
	json.Unmarshal(b, willBeSent)
	request := &Request{
		ID:        willBeSent.RequestID,
		Method:    willBeSent.Request.Method,
		URL:       willBeSent.Request.URL,
		BodySize:  len(willBeSent.Request.PostData),
		Timestamp: willBeSent.Timestamp,
		Headers:   parseHeaders(willBeSent.Request.Headers),
	}

	if willBeSent.Request.HasPostData {
		mimeType := ""
		if ct, ok := willBeSent.Request.Headers["Content-Type"]; ok {
			mimeType = ct.(string)
		}
		request.PostData = &PostData{
			MimeType: mimeType,
			Text:     willBeSent.Request.PostData,
		}
	}

	request.QueryString, err = parseURL(willBeSent.Request.URL)
	if err != nil {
		// log.Print(err)
		return
	}

	entry := &Entry{
		StartedDateTime: epoch(willBeSent.WallTime), //epoch float64, eg 1440589909.59248
		Pageref:         har.harPageID(),
		Request:         request,
		Cache:           &Cache{},
		PageTimings:     &PageTimings{},
	}
	entry.Request.ID = willBeSent.RequestID

	if willBeSent.RedirectResponse != nil {
		e := har.entryByRequestID(willBeSent.RequestID)
		if e == nil {
			// log.Print(errPoorRequest)
			return
		}
		e.Request.ID = e.Request.ID + "r"
		addResponse(e, willBeSent.Timestamp, willBeSent.RedirectResponse)
		e.Response.RedirectURL = willBeSent.Request.URL
		e.PageTimings.Receive = 0.0
	}

	har.Entries = append(har.Entries, entry)
	page := har.currentPage()

	// if this is the primary page, set the Page.Title to the request URL
	if page.Title == "" {
		page.Title = request.URL
		page.Timestamp = willBeSent.Timestamp
	}
}

func (har *HAR) responseReceived(parameters cdp.Params) {
	if len(har.Pages) < 1 {
		return
	}
	received := &cdp.ResponseReceived{}
	b, err := json.Marshal(parameters)
	if err != nil {
		// log.Print(err)
		return
	}
	json.Unmarshal(b, received)
	entry := har.entryByRequestID(received.RequestID)
	if entry == nil {
		// log.Print(errPoorRequest)
		return
	}
	addResponse(entry, received.Timestamp, received.Response)
}

func (har *HAR) dataReceived(parameters cdp.Params) {
	if len(har.Pages) < 1 {
		return
	}
	received := &cdp.DataReceived{}
	b, err := json.Marshal(parameters)
	if err != nil {
		// log.Print(err)
		return
	}
	json.Unmarshal(b, received)
	entry := har.entryByRequestID(received.RequestID)
	if entry == nil {
		// log.Print(errPoorRequest)
		return
	}
	entry.Response.Content.Size += received.DataLength
}

func (har *HAR) loadingFinished(parameters cdp.Params) {
	if len(har.Pages) < 1 {
		return
	}
	finished := &cdp.LoadingFinished{}
	b, err := json.Marshal(parameters)
	if err != nil {
		// log.Print(err)
		return
	}
	json.Unmarshal(b, finished)
	entry := har.entryByRequestID(finished.RequestID)
	if entry == nil {
		// log.Print(errPoorRequest)
		return
	}
	entry.Response.BodySize = int64(finished.EncodedDataLength) - int64(entry.Response.HeadersSize)
	entry.Response.Content.Compression = entry.Response.Content.Size - entry.Response.BodySize
	entry.Time = (finished.Timestamp - entry.Request.Timestamp) * 1000
	entry.PageTimings.Receive = entry.PageTimings.Receive + int(finished.Timestamp)*1000
}

func (har *HAR) requestServedFromCache(parameters cdp.Params) {
	if len(har.Pages) < 1 {
		return
	}
	servedFromCache := &cdp.ServedFromCache{}
	b, err := json.Marshal(parameters)
	if err != nil {
		// log.Print(err)
		return
	}
	json.Unmarshal(b, servedFromCache)
	entry := har.entryByRequestID(servedFromCache.RequestID)
	if entry == nil {
		// log.Print(errPoorRequest)
		return
	}
	entry.Cache.BeforeRequest = &CacheObject{
		LastAccess: "",
		ETag:       "",
		HitCount:   0,
	}
}

func (har *HAR) loadingFailed(parameters cdp.Params) {
	if len(har.Pages) < 1 {
		return
	}
	loadingFailed := &cdp.LoadingFailed{}
	b, err := json.Marshal(parameters)
	if err != nil {
		// log.Print(err)
		return
	}
	json.Unmarshal(b, loadingFailed)
	entry := har.entryByRequestID(loadingFailed.RequestID)
	if entry == nil {
		// log.Print(errPoorRequest)
		return
	}
	entry.Response = &Response{
		Status:     0,
		StatusText: loadingFailed.ErrorText,
		Timestamp:  loadingFailed.Timestamp,
	}
}

func (har *HAR) loadEventFired(parameters cdp.Params) {
	if len(har.Pages) < 1 {
		return
	}
	loadEvent := &cdp.PageLoadEventFired{}
	b, err := json.Marshal(parameters)
	if err != nil {
		// log.Print(err)
		return
	}
	json.Unmarshal(b, loadEvent)
	page := har.currentPage()
	page.PageTiming.OnLoad = int64((loadEvent.Timestamp - page.Timestamp) * 1000)
}

func (har *HAR) domContentEventFired(parameters cdp.Params) {
	if len(har.Pages) < 1 {
		return
	}
	domContentEvent := &cdp.PageDomContentEventFired{}
	b, err := json.Marshal(parameters)
	if err != nil {
		// log.Print(err)
		return
	}
	json.Unmarshal(b, domContentEvent)
	page := har.currentPage()
	page.PageTiming.OnContentLoad = int64((domContentEvent.Timestamp - page.Timestamp) * 1000)
}

func (har *HAR) harPageID() string {
	return fmt.Sprintf("page_%d", len(har.Pages))
}

func (har *HAR) currentPage() *Page {
	if len(har.Pages) < 1 {
		return nil
	}
	return har.Pages[len(har.Pages)-1]
}

func (har *HAR) entryByRequestID(id string) *Entry {
	for _, e := range har.Entries {
		if e.Request.ID == id {
			return e
		}
	}
	return nil
}

func epoch(epoch float64) time.Time {
	return time.Unix(0, int64(epoch*1000)*int64(time.Millisecond))
}

func parseHeaders(headers map[string]interface{}) []*NVP {
	h := make([]*NVP, 0)
	for k, v := range headers {
		h = append(h, &NVP{
			Name:  k,
			Value: v.(string),
		})
	}
	return h
}

func parseURL(strURL string) ([]*NVP, error) {
	nvps := make([]*NVP, 0)
	reqURL, err := url.Parse(strURL)
	if err != nil {
		return nil, err
	}
	for k, v := range reqURL.Query() {
		for _, value := range v {
			nvps = append(nvps, &NVP{
				Name:  k,
				Value: value,
			})
		}
	}
	return nvps, nil
}

func addResponse(entry *Entry, timestamp float64, response *cdp.Response) {
	entry.Request.HTTPVersion = response.Protocol
	// entry.Request.Headers = parseHeaders(response.RequestHeaders)
	entry.Request.SetHeadersSize()
	entry.Request.SetCookies()

	resp := &Response{
		Status:      response.Status,
		StatusText:  response.StatusText,
		HTTPVersion: entry.Request.HTTPVersion,
		Headers:     parseHeaders(response.Headers),
		Timestamp:   timestamp,
	}
	resp.SetHeadersSize()
	// cookies := response.parseCookie()
	// for _, c := range cookies {
	// 	resp.Cookies = append(resp.Cookies, &Cookie{
	// 		Name:    c.Name,
	// 		Value:   c.Value,
	// 		Path:    c.Path,
	// 		Domain:  c.Domain,
	// 		Expires: string(c.Expires),
	// 		Secure:  c.Secure,
	// 	})
	// }
	entry.Response = resp
	entry.ServerIPAddress = response.RemoteIPAddress

	entry.Response.Content = &Content{
		MimeType: response.MimeType,
	}
	if response.Timing == nil {
		return
	}
	entry.PageTimings = &PageTimings{
		Blocked: int(math.Min(0.0, response.Timing.DNSStart)),
		DNS:     int(math.Min(0.0, response.Timing.DNSEnd-response.Timing.DNSStart)),
		Connect: int(math.Min(0.0, response.Timing.ConnectEnd-response.Timing.ConnectStart)),
		Send:    int(math.Min(0.0, response.Timing.SendEnd-response.Timing.SendStart)),
		Wait:    int(math.Min(0.0, response.Timing.ReceiveHeadersEnd-response.Timing.SendEnd)),
		SSL:     int(math.Min(0.0, response.Timing.SSLEnd-response.Timing.SSLStart)),
		Receive: int(0.0 - (response.Timing.RequestTime*1000 + response.Timing.ReceiveHeadersEnd)),
	}
}
