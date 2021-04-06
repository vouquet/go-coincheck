go-coincheck
===

* the library for easy use of [https://coincheck.com/ja/documents/exchange/api](https://coincheck.com/ja/documents/exchange/api)
	* cannot order yet.

## sample

* easy
```

import (
	"fmt"
	"context"
)

import "github.com/vouquet/go-coincheck/coincheck"

func main() {
	API_KEY = "your api key"
	SECRET_KEY = "your secret key"

	shop, err := coincheck.NewCoincheck(API_KEY, SECRET_KEY, context.Background())
	if err != nil {
		panic(err)
	}
	defer shop.Close()

	rates, err = shop.GetRates()
	if err != nil {
		panic(err)
	}

	for symbol, rate := range rates {
		fmt.Println(symbol, rate)
	}
}
```

