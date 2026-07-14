package history

import (
	"errors"

	"github.com/imsarkie/websock-market-stream/internal/model"
)

var ErrLimitExceedsCapacity = errors.New("limit exceeds store capacity")

type Store struct{
	candles []model.Candle
	capacity int
}

func NewStore(capacity int) *Store{
	return &Store{
		candles: make([]model.Candle, 0, capacity),
		capacity: capacity,
	}
}

func (s *Store) SaveCandle(candle model.Candle) error{
	
	if len(s.candles) >= s.capacity{
		s.candles =s.candles[1:]
	}

	s.candles = append(s.candles, candle)
	return nil
}

func (s *Store) GetAll() []model.Candle{
	return s.candles
}

func (s *Store) GetCandles(limit int) ([]model.Candle, error){
	if limit > s.capacity{
		return nil, ErrLimitExceedsCapacity
	}
	if limit >= len(s.candles){
		return s.candles, nil
	}
	start := len(s.candles) - limit
	return s.candles[start:], nil
} 