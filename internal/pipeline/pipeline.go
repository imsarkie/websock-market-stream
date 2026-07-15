package pipeline

import (
	"fmt"
	"log"

	"github.com/imsarkie/websock-market-stream/internal/candle"
	"github.com/imsarkie/websock-market-stream/internal/history"
	"github.com/imsarkie/websock-market-stream/internal/model"
	"github.com/imsarkie/websock-market-stream/internal/mysql"
	"github.com/imsarkie/websock-market-stream/internal/timeframe"
	"github.com/imsarkie/websock-market-stream/internal/ws"
)

// Defining the pipeline
type Pipeline struct {
	server *ws.Server
	engine *candle.Engine
	history *history.Store
	mysql *mysql.Store
	manager *timeframe.Manager
}

// Constructor
func New(ws *ws.Server, engine *candle.Engine, history *history.Store, mysql *mysql.Store, manager *timeframe.Manager) *Pipeline {
	return &Pipeline{
		server: ws,
		engine: engine,
		history: history,
		mysql: mysql,
		manager: manager,
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

	// ----------------- With single timeframe implementation -----------------//

	// candle, completed := p.engine.Update(trade)

	// if completed {
	// 	fmt.Println("Candle Broadcast sent !!!")

	// 	if err := p.mysql.SaveCandle(*candle); err != nil {
	// 		return err
	// 	}

	// 	// Saving history in the hisotry
	// 	p.history.SaveCandle(*candle)
	// 	log.Printf("History size: %d", len(p.history.GetAll()))

	// 	// Broadcast to the browser for chart.
	// 	p.server.Broadcast(candle)
	// }

	// ----------------- With multiple timeframe implementation -----------------//

	candles := p.manager.Update(trade)
	for _, candle := range candles{
		p.history.SaveCandle(*candle)
		log.Printf("History size: %d", len(p.history.GetAll()))

		p.server.Broadcast(candle)
		fmt.Println("Candle Broadcast sent !!!")
	}

	return nil
}
