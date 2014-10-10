package main

import (
	as "github.com/aerospike/aerospike-client-go"
	"time"
)

type stat struct {
	ticker string
	price  uint64
}

var stats chan *stat

func statsBoss() {
	statsMap := map[string][]uint64{}
	for {
		select {
		case s := <-stats:
			if lst, exists := statsMap[s.ticker]; !exists {
				statsMap[s.ticker] = []uint64{s.price}
			} else {
				statsMap[s.ticker] = append(lst, s.price)
			}

		case <-time.After(1 * time.Second):
			for ticker, stats := range statsMap {
				if len(stats) == 0 {
					continue
				}

				sum := 0
				for _, p := range stats {
					sum += int(p)
				}

				key, _ := as.NewKey(NAMESPACE, STOCKS, ticker)
				go db.Put(nil, key, as.BinMap{"price": sum / len(stats)})
			}
			// reset the stats
			statsMap = map[string][]uint64{}
		}
	}
}
