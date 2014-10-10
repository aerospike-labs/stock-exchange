package main

import (
	as "github.com/aerospike/aerospike-client-go"
	"os"
)

func seed_db() {
	println("Sedding db...")

	var key *as.Key

	println("Sedding stocks...")

	// put stocks
	key, _ = as.NewKey(NAMESPACE, STOCKS, "GOOG")
	db.Put(nil, key, as.BinMap{
		"ticker":   "GOOG",
		"quantity": int64(1e9),
		"price":    int(5200),
	})

	key, _ = as.NewKey(NAMESPACE, STOCKS, "APPL")
	db.Put(nil, key, as.BinMap{
		"ticker":   "APPL",
		"quantity": int64(1e9),
		"price":    8200,
	})

	key, _ = as.NewKey(NAMESPACE, STOCKS, "FB")
	db.Put(nil, key, as.BinMap{
		"ticker":   "FB",
		"quantity": int64(1e9),
		"price":    5200,
	})

	key, _ = as.NewKey(NAMESPACE, STOCKS, "MSFT")
	db.Put(nil, key, as.BinMap{
		"ticker":   "MSFT",
		"quantity": int64(1e9),
		"price":    2100,
	})

	key, _ = as.NewKey(NAMESPACE, STOCKS, "GE")
	db.Put(nil, key, as.BinMap{
		"ticker":   "GE",
		"quantity": int64(1e9),
		"price":    1100,
	})

}

func seed_broker(broker_id int, broker_name string) {
	println("creating broker...")
	if broker_name == "" {
		println("give a name for the broker")
		os.Exit(1)
	}

	key, _ := as.NewKey(NAMESPACE, BROKERS, broker_id)
	db.Delete(nil, key)
	if err := db.Put(nil, key, as.BinMap{
		"brokerId":    broker_id,
		"broker_name": broker_name,
		"credit":      int64(1e6),
	}); err != nil {
		println("Failed creating the broker: ", err.Error())
	}
	println("broker `", broker_id, ":", broker_name, "` created successfully")
}
