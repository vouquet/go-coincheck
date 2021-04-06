package coincheck

import (
	"time"
)

type Rate struct {
	RawAsk     float64 `json:"ask"`
	RawBid     float64 `json:"bid"`
	RawHigh    float64 `json:"high"`
	RawLow     float64 `json:"low"`
	RawTime    int64   `json:"timestamp"`
	RawVolume  float64 `json:"volume"`
	RawLast    float64 `json:"last"`

	pair       string
	time       time.Time
}

func (self *Rate) Ask() float64 {
	return self.RawAsk
}

func (self *Rate) Bid() float64 {
	return self.RawBid
}

func (self *Rate) Symbol() string {
	return self.pair
}

func (self *Rate) Time() time.Time {
	return self.time
}

func (self *Rate) High() float64 {
	return self.RawHigh
}

func (self *Rate) Low() float64 {
	return self.RawLow
}

func (self *Rate) Volume() float64 {
	return self.RawVolume
}

func (self *Rate) Last() float64 {
	return self.RawLast
}

func (self *Rate) fixStruct(pair string) {
	self.pair = pair
	self.time = time.Unix(self.RawTime, 0)
}
