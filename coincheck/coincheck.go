package coincheck

import (
	"sync"
	"context"
	"encoding/json"
)

type Coincheck struct {
	client  *Client

	cancel  context.CancelFunc
	mtx     *sync.Mutex
}

func NewCoincheck(api_key string, secret_key string, b_ctx context.Context) (*Coincheck, error) {
	ctx, cancel := context.WithCancel(b_ctx)
	client := NewClient(api_key, secret_key)
	client.RunPool(ctx)

	self := &Coincheck {
		client:client,
		cancel:cancel,
		mtx:new(sync.Mutex),
	}
	return self, nil
}

func (self *Coincheck) GetRates() (map[string]*Rate, error) {
	self.lock()
	defer self.unlock()

	param := "pair=" + PAIR_BTC_JPY
	ret, err := self.request2PublicPool("GET", "/api/ticker", param, nil)
	if err != nil {
		return nil, err
	}
	var r *Rate
	if err := json.Unmarshal(ret, &r); err != nil {
		return nil, err
	}

	r.fixStruct(PAIR_BTC_JPY)
	rates := make(map[string]*Rate)
	rates[PAIR_BTC_JPY] = r

	return rates, nil
}

func (self *Coincheck) Close() error {
	self.lock()
	defer self.unlock()

	self.cancel()
	return nil
}

func (self *Coincheck) request2PublicPool(method string, path string, param string, body []byte) ([]byte, error) {
	req, err := self.client.NewRequest(method, path, param, body)
	if err != nil {
		return nil, err
	}
	return self.client.PostPublicPool(req)
}

func (self *Coincheck) request2PrivatePool(method string, path string, param string, body []byte) ([]byte, error) {
	req, err := self.client.NewRequest(method, path, param, body)
	if err != nil {
		return nil, err
	}
	return self.client.PostPrivatePool(req)
}

func (self *Coincheck) lock() {
	self.mtx.Lock()
}

func (self *Coincheck) unlock() {
	self.mtx.Unlock()
}
