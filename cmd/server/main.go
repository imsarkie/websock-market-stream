package main

import (
	// "encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/imsarkie/websock-market-stream/internal/binance"
	"github.com/imsarkie/websock-market-stream/internal/candle"
	"github.com/imsarkie/websock-market-stream/internal/pipeline"
	"github.com/imsarkie/websock-market-stream/internal/ws"
)



func main(){
	client := binance.NewClient(
		"wss://stream.binance.com:9443/ws/bnbusdt@aggTrade",
	)
	// var trade model.AggTrade

	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}

	server := ws.NewServer()
	engine := candle.New(10 * time.Second)

	go server.Start()

	pipe := pipeline.New(server, engine)


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