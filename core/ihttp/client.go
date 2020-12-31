package ihttp

import (
	"crypto/tls"
	"github.com/helloh2o/lucky/log"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Req is http client req
type Req struct {
	http.Client
	proxy string
}

var (
	// UA is user agent
	UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.3578.98 Safari/537.36"
)

// NewHttpClient get new client
func NewHttpClient(proxyUrl ...string) *Req {
	req := new(Req)
	req.Client = http.Client{}
	req.Jar = new(Jar)
	req.Timeout = time.Second * 3
	tr := &http.Transport{}
	if len(proxyUrl) > 0 {
		proxyFunc := func(r *http.Request) (*url.URL, error) {
			r.Header.Set("User-Agent", UA)
			return url.Parse(proxyUrl[0])
		}
		tr.Proxy = proxyFunc
	}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req.Transport = tr
	return req
}

// DoReq client request target url
func (req *Req) DoReq(method string, targetUrl string, callback func(*url.URL, io.Reader)) {
	switch method {
	case "POST", "GET", "PUT", "DELETE", "HEAD":
		urlInfo, err := url.Parse(targetUrl)
		if err != nil {
			log.Error("can't parse url error %v", err)
		}
		reqInfo, err := http.NewRequest(method, targetUrl, nil)
		if err != nil {
			log.Error("new http request for url %s error", targetUrl)
			return
		}
		reqInfo.Header.Add("Host", urlInfo.Host)
		reqInfo.Header.Add("User-Agent", UA)
		var resp *http.Response
		resp, err = req.Do(reqInfo)
		if err != nil {
			log.Error("do request error %v", err)
			return
		}
		if resp.StatusCode == 200 {
			if callback != nil {
				callback(urlInfo, resp.Body)
			}
		} else {
			log.Error("resp code %d not ok", resp.StatusCode)
		}
	default:
		log.Error("no support method %s", method)
	}
}

// Jar store the cookies
type Jar struct {
	cookies []*http.Cookie
}

// SetCookies new cookies
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies = cookies
}

// Cookies get all cookies
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies
}
