package model

import "time"

type Candle struct{
	Symbol 		string
	Timeframe 	string
	
	Open 		float64
	High		float64
	Low			float64
	Close		float64

	Volume		float64

	StartTime 	time.Time
	EndTime		time.Time
}