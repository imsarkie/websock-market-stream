package candle

import (
	"time"

	"github.com/imsarkie/websock-market-stream/internal/model"
)

type Engine struct{
	current *model.Candle
	timeframe time.Duration
}

func New(tf time.Duration) *Engine{
	return &Engine{
		timeframe: tf,
	}
}

// Create the Candle
func (e *Engine) createCandle(trade model.Trade){
	start := trade.TradeTime.Truncate(e.timeframe)
	end := start.Add(e.timeframe)

	e.current = &model.Candle{
		Symbol: trade.Symbol,

		Open: trade.Price,
		High: trade.Price,
		Low: trade.Price,
		Close: trade.Price,

		Volume: trade.Quantity,
			
		StartTime: start,
		EndTime: end,
	}
}

func (e *Engine) Update(trade model.Trade) (*model.Candle, bool){
	// TODO
	/*
	1. Create the Candle
	2. Update High
	3. Update Low
	4. Update Close
	5. Update Volume
	6. Detect when timeframe ends
	7. Emit the complete candle out
	*/

	

	// create candle if not already there
	if e.current == nil{
		// Create new one
		e.createCandle(trade)
		return nil, false
	}
	
	if !trade.TradeTime.Before(e.current.EndTime){
		// current candle finished
		completed := e.current 	// saving the previous candle in 'completed' to return it
		e.createCandle(trade) 	// new candle for fresh timeframe
		return completed, true
	}

	if trade.Price > e.current.High{
		e.current.High = trade.Price
	}
	if trade.Price < e.current.Low{
		e.current.Low = trade.Price
	}
	
	e.current.Volume += trade.Quantity
	e.current.Close = trade.Price
	// e.current.EndTime = trade.TradeTime
	
	return nil, false
}