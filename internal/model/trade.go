package model

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