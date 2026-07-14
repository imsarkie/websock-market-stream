package history

import "github.com/imsarkie/websock-market-stream/internal/model"

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

func (s *Store) SaveCandle(candle model.Candle){
	
	if len(s.candles) >= s.capacity{
		s.candles =s.candles[1:]
	}

	s.candles = append(s.candles, candle)
}

func (s *Store) GetAll() []model.Candle{
	return s.candles
}