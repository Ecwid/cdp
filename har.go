package cdp

import (
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/ecwid/cdp/har"
)

var errPoorRequest = fmt.Errorf("got a response event with no matching request event")

// HAR ...
type HAR struct {
	Log *har.HAR `json:"log"`
}

func (har *HAR) harPageID() string {
	return fmt.Sprintf("page_%d", len(har.Log.Pages))
}

func (har *HAR) currentPage() *har.Page {
	if len(har.Log.Pages) < 1 {
		return nil
	}
	return har.Log.Pages[len(har.Log.Pages)-1]
}

func (har *HAR) entryByRequestID(id string) *har.Entry {
	for _, e := range har.Log.Entries {
		if e.Request.ID == id {
			return e
		}
	}
	return nil
}

func epoch(epoch float64) time.Time {
	return time.Unix(0, int64(epoch*1000)*int64(time.Millisecond))
}

func parseHeaders(headers map[string]interface{}) []*har.NVP {
	h := make([]*har.NVP, 0)
	for k, v := range headers {
		h = append(h, &har.NVP{
			Name:  k,
			Value: v.(string),
		})
	}
	return h
}

func parseURL(strURL string) ([]*har.NVP, error) {
	nvps := make([]*har.NVP, 0)
	reqURL, err := url.Parse(strURL)
	if err != nil {
		return nil, err
	}
	for k, v := range reqURL.Query() {
		for _, value := range v {
			nvps = append(nvps, &har.NVP{
				Name:  k,
				Value: value,
			})
		}
	}
	return nvps, nil
}

func (session *Session) eventHar(method string, parameters interface{}) error {
	var shar = session.har
	var err error
	switch method {

	case "Page.frameStartedLoading", "Page.navigatedWithinDocument":
		page := har.Page{
			StartedDateTime: time.Now(),
			PageTiming:      &har.PageTiming{},
		}
		shar.Log.Pages = append(shar.Log.Pages, &page)
		page.ID = shar.harPageID()

	case "Network.requestWillBeSent":
		if len(shar.Log.Pages) < 1 {
			return fmt.Errorf("sending request object, but frame not started")
		}
		willBeSent := &RequestWillBeSent{}
		unmarshal(parameters, willBeSent)
		request := &har.Request{
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
			request.PostData = &har.PostData{
				MimeType: mimeType,
				Text:     willBeSent.Request.PostData,
			}
		}

		request.QueryString, err = parseURL(willBeSent.Request.URL)
		if err != nil {
			return err
		}

		entry := &har.Entry{
			StartedDateTime: epoch(willBeSent.WallTime), //epoch float64, eg 1440589909.59248
			Pageref:         shar.harPageID(),
			Request:         request,
			Cache:           &har.Cache{},
			PageTimings:     &har.PageTimings{},
		}
		entry.Request.ID = willBeSent.RequestID

		if willBeSent.RedirectResponse != nil {
			e := shar.entryByRequestID(willBeSent.RequestID)
			if e == nil {
				return errPoorRequest
			}
			e.Request.ID = e.Request.ID + "r"
			addResponse(e, willBeSent.Timestamp, willBeSent.RedirectResponse)
			e.Response.RedirectURL = willBeSent.Request.URL
			e.PageTimings.Receive = 0.0
		}

		shar.Log.Entries = append(shar.Log.Entries, entry)
		page := shar.currentPage()

		// if this is the primary page, set the Page.Title to the request URL
		if page.Title == "" {
			page.Title = request.URL
			page.Timestamp = willBeSent.Timestamp
		}

	case "Network.responseReceived":
		if len(shar.Log.Pages) < 1 {
			return nil
		}
		received := &ResponseReceived{}
		unmarshal(parameters, &received)
		entry := shar.entryByRequestID(received.RequestID)
		if entry == nil {
			return errPoorRequest
		}
		addResponse(entry, received.Timestamp, received.Response)

	case "Network.dataReceived":
		if len(shar.Log.Pages) < 1 {
			return nil
		}
		received := &DataReceived{}
		unmarshal(parameters, received)
		entry := shar.entryByRequestID(received.RequestID)
		if entry == nil {
			return errPoorRequest
		}
		entry.Response.Content.Size += received.DataLength

	case "Network.loadingFinished":
		if len(shar.Log.Pages) < 1 {
			return nil
		}
		finished := &LoadingFinished{}
		unmarshal(parameters, finished)
		entry := shar.entryByRequestID(finished.RequestID)
		if entry == nil {
			return errPoorRequest
		}
		entry.Response.BodySize = int64(finished.EncodedDataLength) - int64(entry.Response.HeadersSize)
		entry.Response.Content.Compression = entry.Response.Content.Size - entry.Response.BodySize
		entry.Time = (finished.Timestamp - entry.Request.Timestamp) * 1000
		entry.PageTimings.Receive = entry.PageTimings.Receive + int(finished.Timestamp)*1000

	case "Network.requestServedFromCache":
		if len(shar.Log.Pages) < 1 {
			return nil
		}
		servedFromCache := &ServedFromCache{}
		unmarshal(parameters, servedFromCache)
		entry := shar.entryByRequestID(servedFromCache.RequestID)
		if entry == nil {
			return errPoorRequest
		}
		entry.Cache.BeforeRequest = &har.CacheObject{
			LastAccess: "",
			ETag:       "",
			HitCount:   0,
		}

	case "Network.loadingFailed":
		if len(shar.Log.Pages) < 1 {
			return nil
		}
		loadingFailed := &LoadingFailed{}
		unmarshal(parameters, loadingFailed)
		entry := shar.entryByRequestID(loadingFailed.RequestID)
		if entry == nil {
			return errPoorRequest
		}
		entry.Response = &har.Response{
			Status:     0,
			StatusText: loadingFailed.ErrorText,
			Timestamp:  loadingFailed.Timestamp,
		}

	case "Page.loadEventFired":
		if len(shar.Log.Pages) < 1 {
			return nil
		}
		loadEvent := &PageLoadEventFired{}
		unmarshal(parameters, loadEvent)
		page := shar.currentPage()
		page.PageTiming.OnLoad = int64((loadEvent.Timestamp - page.Timestamp) * 1000)

	case "Page.domContentEventFired":
		if len(shar.Log.Pages) < 1 {
			return nil
		}
		domContentEvent := &PageDomContentEventFired{}
		unmarshal(parameters, domContentEvent)
		page := shar.currentPage()
		page.PageTiming.OnContentLoad = int64((domContentEvent.Timestamp - page.Timestamp) * 1000)
	}
	return nil
}

func addResponse(entry *har.Entry, timestamp float64, response *Response) {
	entry.Request.HTTPVersion = response.Protocol
	// entry.Request.Headers = parseHeaders(response.RequestHeaders)
	entry.Request.SetHeadersSize()
	entry.Request.SetCookies()

	resp := &har.Response{
		Status:      response.Status,
		StatusText:  response.StatusText,
		HTTPVersion: entry.Request.HTTPVersion,
		Headers:     parseHeaders(response.Headers),
		Timestamp:   timestamp,
	}
	resp.SetHeadersSize()
	// cookies := response.parseCookie()
	// for _, c := range cookies {
	// 	resp.Cookies = append(resp.Cookies, &har.Cookie{
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

	entry.Response.Content = &har.Content{
		MimeType: response.MimeType,
	}
	if response.Timing == nil {
		return
	}
	entry.PageTimings = &har.PageTimings{
		Blocked: int(math.Min(0.0, response.Timing.DNSStart)),
		DNS:     int(math.Min(0.0, response.Timing.DNSEnd-response.Timing.DNSStart)),
		Connect: int(math.Min(0.0, response.Timing.ConnectEnd-response.Timing.ConnectStart)),
		Send:    int(math.Min(0.0, response.Timing.SendEnd-response.Timing.SendStart)),
		Wait:    int(math.Min(0.0, response.Timing.ReceiveHeadersEnd-response.Timing.SendEnd)),
		SSL:     int(math.Min(0.0, response.Timing.SSLEnd-response.Timing.SSLStart)),
		Receive: int(0.0 - (response.Timing.RequestTime*1000 + response.Timing.ReceiveHeadersEnd)),
	}
}
