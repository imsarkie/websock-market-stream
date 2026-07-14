package mysql

import (
	"github.com/imsarkie/websock-market-stream/internal/model"
)

func (s *Store) SaveCandle(candle model.Candle) error{
	_, err := s.db.Exec(
		`
		Insert into candles(
		symbol,
		open, 
		high,
		low,
		close,
		volume,
		start_time,
		end_time
		)	
		values (?, ?, ?, ?, ?, ?, ?, ?)
		`,
		candle.Symbol,
		candle.Open,
		candle.High,
		candle.Low,
		candle.Close,
		candle.Volume,
		candle.StartTime,
		candle.EndTime,
	)

	if err != nil {
		return err
	}
	// defer s.db.Close()
	return nil
}

func (s *Store) GetCandles(limit int) ([]model.Candle, error){
	rows, err := s.db.Query(
		`
		SELECT
        symbol,
        open,
        high,
        low,
        close,
        volume,
        start_time,
        end_time
    FROM candles
    ORDER BY start_time DESC
    LIMIT ?
		`, limit,
	)
	if err != nil {
    	return nil, err
	}

	var candles []model.Candle
	for rows.Next(){
		var candle model.Candle
		err := rows.Scan(
			&candle.Symbol,
    		&candle.Open,
    		&candle.High,
    		&candle.Low,
    		&candle.Close,
    		&candle.Volume,
    		&candle.StartTime,
    		&candle.EndTime,
		)
		if err != nil {
    		return nil, err
		}
		candles = append(candles, candle)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i, j := 0, len(candles)-1; i < j; i, j = i+1, j-1 {
		candles[i], candles[j] = candles[j], candles[i]
	}

	// defer s.db.Close()
	return candles, nil
}