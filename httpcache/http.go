// A dumb memory consuming http request cacher.
package httpcache

import (
	"io/ioutil"
	"net/http"
)

type Client struct {
	queue  chan *request
	wqueue chan *request
	cache  map[string]*request
}

type request struct {
	u    string
	r    *http.Request
	body []byte
	chs  chan chan []byte
	ch   chan []byte
}

func (r *request) fullfill() {
	for i := range r.chs {
		i <- r.body
	}
}

func New(concurrency int) *Client {
	c := &Client{
		make(chan *request),
		make(chan *request),
		make(map[string]*request),
	}

	c.run(concurrency)
	return c
}

func (c *Client) Exec(req *http.Request) chan []byte {
	url := req.Method + req.URL.String()
	ret := make(chan []byte, 10)
	c.queue <- &request{url, req, nil, nil, ret}

	return ret
}

func (c *Client) run(concurrency int) {
	go func() {
		for r := range c.queue {
			if _, ok := c.cache[r.u]; !ok {
				c.cache[r.u] = r
				r.chs = make(chan chan []byte, 10)
				c.wqueue <- r
			}

			c.cache[r.u].chs <- r.ch
		}
	}()

	for i := 0; i < concurrency; i++ {
		go func() {
			hc := &http.Client{}
			for r := range c.wqueue {
				resp, err := hc.Do(r.r)
				if err != nil {
					continue
				}
				data, err := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					continue
				}

				r.body = data
				go r.fullfill()
			}
		}()
	}
}
