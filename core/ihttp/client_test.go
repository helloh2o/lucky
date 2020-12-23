package ihttp

import (
	"io"
	"io/ioutil"
	"lucky/log"
	"net/url"
	"testing"
)

func TestNewHttpClient(t *testing.T) {
	client := NewHttpClient()
	client.DoReq("GET", "https://baidu.com", func(url *url.URL, reader io.Reader) {
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		log.Debug(string(data))
	})
}
