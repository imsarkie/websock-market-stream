package pipeline

import (
	"fmt"

	"github.com/imsarkie/websock-market-stream/internal/candle"
	"github.com/imsarkie/websock-market-stream/internal/model"
	"github.com/imsarkie/websock-market-stream/internal/ws"
)

// Defining the pipeline
type Pipeline struct {
	server *ws.Server
	engine *candle.Engine
}

// Constructor
func New(ws *ws.Server, engine *candle.Engine) *Pipeline {
	return &Pipeline{
		server: ws,
		engine: engine,
	}
}

func (p *Pipeline) broadcast(trade model.Trade) error {
	return p.server.Broadcast(trade)
}

// ProcessTrade Method
func (p *Pipeline) ProcessTrade(trade model.Trade) error {

	// if err := p.broadcast(trade); err != nil {
	// 	return err
	// }

	candle, completed := p.engine.Update(trade)
	if completed {
		fmt.Println("Candle Broadcast sent !!!")
		p.server.Broadcast(candle)
	}

	return nil
}
