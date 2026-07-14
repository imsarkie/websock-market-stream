package repository

import "github.com/imsarkie/websock-market-stream/internal/model"

type CandleRepository interface{
	SaveCandle(candle model.Candle)
	GetCandles(limit int) []model.Candle
}