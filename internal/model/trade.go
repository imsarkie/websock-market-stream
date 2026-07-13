package model

import (
	"strconv"
	"time"
)

type AggTrade struct{
	EventType 		string	`json:"e"`
	EventTime 		int64	`json:"E"`
	AggId 			int64	`json:"a"`
	Symbol 			string	`json:"s"`
	Price 			string	`json:"p"`
	Quantity 		string	`json:"q"`
	FirstTradeID 	int64	`json:"f"`
	LastTradeID 	int64	`json:"l"`
	TradeTime 		int64	`json:"T"`
	MarketMaker 	bool	`json:"m"`
	Ignore 			bool 	`json:"M"`
}

// Trade is the cleaned-up internal representation used by the
// candle engine and broadcast to clients.
type Trade struct{
	Symbol 		string
	Price 		float64
	Quantity 	float64
	TradeTime 	time.Time
}

// ToTrade converts the raw Binance wire format into Trade.
func (a AggTrade) ToTrade() (Trade, error){
	price, err := strconv.ParseFloat(a.Price, 64)
	if err != nil {
		return Trade{}, err
	}

	quantity, err := strconv.ParseFloat(a.Quantity, 64)
	if err != nil {
		return Trade{}, err
	}

	return Trade{
		Symbol: 	a.Symbol,
		Price: 		price,
		Quantity: 	quantity,
		TradeTime: 	time.UnixMilli(a.TradeTime),
	}, nil
}