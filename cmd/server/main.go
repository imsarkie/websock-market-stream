package main

import (
	// "encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/imsarkie/websock-market-stream/internal/binance"
	"github.com/imsarkie/websock-market-stream/internal/candle"
	"github.com/imsarkie/websock-market-stream/internal/history"
	"github.com/imsarkie/websock-market-stream/internal/mysql"

	// "github.com/imsarkie/websock-market-stream/internal/mysql"
	"github.com/imsarkie/websock-market-stream/internal/pipeline"
	"github.com/imsarkie/websock-market-stream/internal/ws"
)

func main() {
	client := binance.NewClient(
		"wss://stream.binance.com:9443/ws/bnbusdt@aggTrade",
	)
	// var trade model.AggTrade

	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}

	mysqlStore, err := mysql.NewStore(
    	"root:2159@tcp(localhost:3306)/marketstream?parseTime=true",
	)
	if err != nil {
    	log.Fatal(err)
	}
	defer mysqlStore.Close()

	history := history.NewStore(500)
	server := ws.NewServer(history, mysqlStore)
	engine := candle.New(30 * time.Second)


	go server.Start()

	// mysqlStore, err := mysql.NewStore(
	// 	"root:2159@tcp(localhost:3306)/marketstream",
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Print(mysqlStore)

	pipe := pipeline.New(
		server,
		engine,
		history,
		mysqlStore,
	)

	defer client.Conn.Close()
	fmt.Println("Connected to Binance!")

	for {
		// messageType, message, err := client.Conn.ReadMessage()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// log.Println("Message Type: ", messageType)
		// // log.Println("Message: ", string(message))

		// err = json.Unmarshal(message, &trade)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		trade, err := client.ReadTrade()
		if err != nil {
			log.Fatal(err)
		}

		// tradeJSON, err := json.Marshal(trade)
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// }

		// server.Broadcast(tradeJSON)

		err = pipe.ProcessTrade(trade)
		if err != nil {
			log.Println(err)
		}

		// fmt.Println(trade)
		// fmt.Printf(
		// 	"Symbol: %s | Price: %s | Quantity: %s\n",
		// 	trade.Symbol,
		// 	trade.Price,
		// 	trade.Quantity,
		// )
	}
}
