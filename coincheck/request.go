package coincheck

import (
	"io/ioutil"
	"fmt"
	"net/http"
	"sync"
	"time"
	"bytes"
	"crypto/sha256"
	"crypto/hmac"
	"strconv"
	"context"
	"encoding/hex"
)

const (
	URL_BASE     = "https://coincheck.com"

	PRIVATE_LIMIT_MILLISEC int = 301
	PUBLICK_LIMIT_MILIISEC int = 301
)

type Client struct {
	key    string
	secret string

	rq_pub_c  chan *request
	rq_pri_c  chan *request
}

func NewClient(key string, secret string) *Client {
	return &Client {
		key: key,
		secret: secret,
		rq_pub_c: make(chan *request),
		rq_pri_c: make(chan *request),
	}
}

func (self *Client) NewRequest(method string, path string, param string, body []byte) (*Request, error) {
	if body == nil {
		body = []byte{}
	}

	t_stmp := NewTimestamp()
	sig := self.genhmac([]byte(t_stmp.UnixString() + method + path + string(body)))

	var url string
	if param == "" {
		url = URL_BASE + path
	} else {
		url = URL_BASE + path + "?" + param
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("API-KEY", self.key)
	req.Header.Set("ACCESS-NONCE", t_stmp.UnixString())
	req.Header.Set("ACCESS-SIGNATURE", string(sig))
	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")

	return NewRequest(req), nil
}

func (self *Client) RunPool(ctx context.Context) {
	tmr_pub := time.NewTicker(time.Millisecond * time.Duration(PUBLICK_LIMIT_MILIISEC))
	tmr_pri := time.NewTicker(time.Millisecond * time.Duration(PRIVATE_LIMIT_MILLISEC))

	go func() {
		for {
			select {
			case <- ctx.Done():
				return
			case rq := <- self.rq_pub_c:
				select {
				case <- ctx.Done():
					return
				case <- rq.life.C:
					rq.Error(fmt.Errorf("HttpRequest: Timeout."))
					continue
				case <- tmr_pub.C:
					b, err := rq.req.Do()
					if err != nil {
						rq.Error(fmt.Errorf("HttpRequestError: %s", err))
						continue
					}
					rq.Return(b)
				}
			case rq := <- self.rq_pri_c:
				select {
				case <- ctx.Done():
					return
				case <- rq.life.C:
					rq.Error(fmt.Errorf("HttpRequest: Timeout."))
					continue
				case <- tmr_pri.C:
					b, err := rq.req.Do()
					if err != nil {
						rq.Error(fmt.Errorf("HttpRequestError: %s", err))
						continue
					}
					rq.Return(b)
				}
			}
		}
	}()
}

func (self *Client) PostPublicPool(r *Request) ([]byte, error) {
	if self.rq_pub_c == nil {
		return nil, fmt.Errorf("undefined pool channel.")
	}

	rq := newRequest(r)
	go func() {
		self.rq_pub_c <- rq
	}()

	rq.WaitDone()
	return rq.Result()
}

func (self *Client) PostPrivatePool(r *Request) ([]byte, error) {
	if self.rq_pri_c == nil {
		return nil, fmt.Errorf("undefined pool channel.")
	}

	rq := newRequest(r)
	go func() {
		self.rq_pri_c <- rq
	}()

	rq.WaitDone()
	return rq.Result()
}

type request struct {
	req   *Request

	life  *time.Timer
	block chan struct{}

	ret   []byte
	err   error
}

func newRequest(req *Request) *request {
	block := make(chan struct{})
	life := time.NewTimer(time.Second * 3)
	return &request{req:req, block:block, life:life}
}

func (self *request) Result() ([]byte, error) {
	if self.err != nil {
		return nil, self.err
	}

	if self.ret == nil {
		return nil, fmt.Errorf("request.Result: empty body")
	}
	return self.ret, nil
}

func (self *request) WaitDone() {
	<-self.block
}

func (self *request) Error(err error) {
	go func() {
		defer self.done()

		self.err = err
	}()
}

func (self *request) Return(b []byte) {
	cp_b := make([]byte, len(b))
	copy(cp_b, b)

	go func() {
		defer self.done()

		self.ret = cp_b
	}()
}

func (self *request) done() {
	close(self.block)
}

type Request struct {
	r  *http.Request

	mtx *sync.Mutex
}

func NewRequest(r *http.Request) *Request {
	return &Request {
		r: r,
		mtx: new(sync.Mutex),
	}
}

func (self *Request) Do() ([]byte, error) {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	c := self.createClient()
	res, err := c.Do(self.r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (self *Request) createClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport)
	t.MaxConnsPerHost = 1

	return &http.Client{
		Timeout: time.Second * 10,
		Transport: t,
	}
}

func (self *Client) genhmac(msg []byte) []byte {
	m := hmac.New(sha256.New, []byte(self.secret))
	m.Write(msg)

	msum := m.Sum(nil)
	ret := make([]byte, hex.EncodedLen(len(msum)))
	hex.Encode(ret, msum)

	return ret
}

type Timestamp struct {
	t  time.Time
}

func NewTimestamp() *Timestamp {
	return &Timestamp{t:time.Now()}
}

func (self *Timestamp) UnixString() string {
	return strconv.FormatInt(self.t.UnixNano()/int64(time.Millisecond), 10)
}
