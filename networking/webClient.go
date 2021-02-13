package networking

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type WebClient struct {
}

func NewWebClient() *WebClient {
	wc := WebClient{}
	return &wc
}

func (wc WebClient) SendTask(url string, requestPayload []byte, ch chan []byte) {
	// To disable connection pooling use:
	// t := http.DefaultTransport.(*http.Transport).Clone()
	// t.DisableKeepAlives = true
	// c := &http.Client{Transport: t}
	// response, err := c.Post(url, "application/octet-stream", bytes.NewReader(requestPayload))

	response, err := http.Post(url, "application/octet-stream", bytes.NewReader(requestPayload))
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	bytes, _ := ioutil.ReadAll(response.Body)

	ch <- bytes
}
