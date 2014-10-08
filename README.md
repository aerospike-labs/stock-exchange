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

3. Event message is as follows:

		{"method": string, "params": [any...]}

	Note Event and Request are the same, except the key difference between the two is the `id` field.


### Exchange is Open

The message brokers will recieve to indicate the market is open is:

	{"method": "open", "params": []}

### Exchange is Closed

The message brokers will recieve to indicate the market is closed is:

	{"method": "close", "params": []}

### Broker Buy Offer

When a broker wants to offer to buy shares, it will send:

	{"method": "buy", "params": [TICKER, QTY, PRICE, TTL], "id": [BROKER, OFFER]}

Where:

- `BROKER` is the broker's token.
- `OFFER` is the broker's identifier for the offer. (opaque value)
- `TICKER` is the ticker symbol to make an offer on.
- `QTY` is the number of shares to of the ticket the offer is valid for.
- `PRICE` is the price per share on the offer.
- `TTL` the time to live for the offer.

The tuple of `[BROKER,OFFER]` must be unique.

The offer will live in the exchange until either it expires, is cancelled or is accepted.

The response will be either an error or an ok:

	{"result": "ok", "id": [BROKER, OFFER]}


### Broker Sell Offer

When a broker wants to offer to sell shares, it will send:

	{"method": "sell", "params": [TICKER, QTY, PRICE, TTL], "id": [BROKER, OFFER]}

Where:

- `BROKER` is the broker's token.
- `OFFER` is the broker's identifier for the offer. (opaque value)
- `TICKER` is the ticker symbol to make an offer on.
- `QTY` is the number of shares to of the ticket the offer is valid for.
- `PRICE` is the price per share on the offer.
- `TTL` the time to live for the offer.

The tuple of `[BROKER,OFFER]` must be unique.

The offer will live in the exchange until either it expires, is cancelled or is accepted.

The response will be either an error or an ok:

	{"result": "ok", "id": [BROKER, OFFER]}


### Broker Cancel Offer

When a broker wants to offer to sell shares, it will send:

	{"method": "cancel", "params": [], "id": [BROKER, OFFER]}

Where:

- `BROKER` is the broker's token.
- `OFFER` is the broker's identifier for the offer. (opaque value)

The tuple of `[BROKER,OFFER]` must be unique.

The response will be either an error or an ok:

	{"result": "ok", "id": [BROKER, OFFER]}

### Broker Accept Offer

When a Broker accepts an offer, they will send an "accept" message containing the id of the offer.

	{"method": "accept": "params": [OFFER_BROKER, OFFER], "id": [ACCEPT_BROKER]}

Where:

- `OFFER_BROKER` is the offering broker's token.
- `OFFER` is the offering broker's offering id. (opaque value)
- `ACCEPT_BROKER` is the accepting broker's token.

The response will be either an error or an ok:

	{"result": "ok", "id": [BROKER, OFFER]}

The accept message will be sent to the offering broker as a signal that transaction has closed.

When the offer is accepted it is logged as a transaction in the exchange, then an "update" event is sent out to the brokers.

### Exchange Price Update Noticiation

When a transaction is closed, then the exchange will broadcast a notification of the last price of a stock.

	{"method": "price", "params": [TICKET, QTY, PRICE]}


### Exchange Price List

Broker can request the current pricelist for all stocks.

	{"method": "prices", "params": [], "id": [BROKER,REQ]}

The response will be:

	{"result": [[TICKET, QTY, PRICE], ...], "id": [BROKER,REQ]}



