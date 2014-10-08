# Stock Exchange in Go

This project contains code for a workshop where we highlight features of both Go and Aerospike.

The project will run a simple stock exchange with multiple brokers, who can buy and sell stock. The exchange is fully functional for the context of this workshop. The broker is simply scafolding, providing the building blocks for developers to implement the logic in Go with persistence in Aerospike.

## How it Works

**Compoents**

- Exchange – will faciliate transactions, manage banks for brokers, and tax brokers for penalties.
- Broker – will make and accept offers.

**Sequence of Events**

- A broker can connect to the exchange at any time. 
- A broker cannnot begin trading until it recieves an "open" event from the exchange.
- Once trading begins, the broker can buy or sell stock. 
	- In order for the borker to buy or sell stock, they have to make an offer.
	- The buy offer is via a "buy" message
	- The sell offer is via a "sell" message
	- If a broker accepts an offer, it sends an "accept" message with the offer id.
	- The exchange will simply complete the transaction if the TTL did not expire and send both brokers an acknowledgement via the "complete" message.
- This will continue until the broker receives a "close" event from the exchange, signaling the closing of the exchange.

**Penalties**

- The Exchange will tax Brokers for inactivity.
- The Exchange will tax Brokers for requesting Price Lists.


## Protocol

The exchange and brokers will communicate over Web Sockers on HTTP. The message format will be [JSON-RPC](http://json-rpc.org/wiki/specification).

To summarize:

1. Request message is as follows:

		{"method": string, "params": [any...], "id": any}

2. Response message is as follows:

		{"result": any, "error": any, "id": any}

3. Notification message is as follows:

		{"method": string, "params": [any...]}

	Note Event and Request are the same, except the key difference between the two is the `id` field.


### Offer

When a broker wants to offer to buy or sell shares, it will send:

	{ "method": "Offer", 
	  "params": [{
	    "BrokerId": BROKER_ID,
	    "OfferId": OFFER_ID,
	    "OfferType": OFFER_TYPE,
	    "TTL": OFFER_TTL,
	    "Ticker": TICKER,
	    "Quantity": QUANTITY,
	    "Price": PRICE
	  }],
	  "id": ID
	}

Where:

- `BROKER_ID` is the broker's identifier.
- `OFFER_ID` is the broker's identifier for the offer. Can be used to reference the offer later, i.e. Cancel.
- `OFFER_TYPE` is either "buy" or "sell"
- `OFFER_TTL` the time to live for the offer.
- `TICKER` is the ticker symbol to make an offer on.
- `QUANTITY` is the number of shares to of the ticket the offer is valid for.
- `PRICE` is the price per share on the offer.
- `ID` the opaque value to be used to match the response to the request.

The offer will live in the exchange until either it expires, is cancelled or is accepted.

The response will be either an error or an ok:

	{"result": "ok", "id": ID}


### Cancel Offer

When a broker wants to offer to sell shares, it will send:

	{ "method": "Cancel", 
	  "params": [{
	    "BrokerId": BROKER_ID, 
	    "OfferId": OFFER_ID
	  }], 
	  "id": ID
	}

Where:

- `BROKER_ID` is the broker's token.
- `OFFER_ID` is the broker's identifier for the offer.
- `ID` the opaque value to be used to match the response to the request.

The response will be either an error or an ok:

	{"result": "ok", "id": ID}

### Transaction Notification

This notification is send to the parties involved in the transaction.

	{ "method": "Transaction", 
	  "params": [{
	    "Buyer": {
	      "BrokerId": BROKER_ID,
	      "OfferId": OFFER_ID
	    },
	    "Seller": {
	      "BrokerId": BROKER_ID,
	      "OfferId": OFFER_ID
	    },
	    "Stock" : {
	      "Ticker": TICKER,
	      "Quantity": QUANTITY,
	      "Price": PRICE
	    }
	  }]
	}

Where:

- `BROKER_ID` is the broker's identifier..
- `OFFER_ID` is the offer's idenfifier. Can be used to reference the offer later, i.e. Cancel.
- `TICKER` is the ticker symbol to make an offer on.
- `QUANTITY` is the number of shares to of the ticket the offer is valid for.
- `PRICE` is the price per share on the offer.
- `ID` the opaque value to be used to match the response to the request.


### Stock List

Broker can request the current list of stocks and last sale price for all stocks.

	{ "method": "StockList", 
	  "params": [], 
	  "id": ID
	}

Where:

- `ID` the opaque value to be used to match the response to the request.

The response will be:

	{ "result": [
	    { "Ticker": TICKER, 
	      "Quantity": QUANTITY,
	      "Price": PRICE
	    }, 
		...
	  ], 
	  "id": ID
	}


### Offer List


Broker can request the current pricelist for all stocks.

	{ "method": "OfferList", 
	  "params": [], 
	  "id": ID
	}

Where:

- `ID` the opaque value to be used to match the response to the request.

The response will be:

	{ "result": [
	    { "BrokerId": BROKER_ID,
	      "OfferId": OFFER_ID,
	      "OfferType": OFFER_TYPE,
	      "Ticker": TICKER,
	      "Quantity": QUANTITY,
	      "Price": PRICE,
	      "TTL": TTL
	    }, 
		...
	  ], 
	  "id": ID
	}
