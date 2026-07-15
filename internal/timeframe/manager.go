package timeframe

import (
	"time"

	"github.com/imsarkie/websock-market-stream/internal/candle"
	"github.com/imsarkie/websock-market-stream/internal/model"
)

type Manager struct{
	engines map[time.Duration]*candle.Engine
}

func New(timeframes ...time.Duration) *Manager{
	engines := make(map[time.Duration]*candle.Engine, len(timeframes))
	for _, tf := range timeframes{
		engines[tf] = candle.New(tf)
	}
	return &Manager{
		engines: engines,
	}
}

func (m *Manager) Update(trade model.Trade) []*model.Candle{
	completed := make([]*model.Candle, 0)
	for _, engine := range m.engines{
		candle, ok := engine.Update(trade)
		if ok{
			completed = append(completed, candle)
		}
	}
	return completed
}