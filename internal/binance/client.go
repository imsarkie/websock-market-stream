package binance

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/imsarkie/websock-market-stream/internal/model"
)

type Client struct{
	URL string
	Conn *websocket.Conn
}

// Creating constructor for new client
func NewClient(url string) *Client{
	return &Client{
		URL: url,
	}
}

func (c *Client) Connect() error{
	conn, _, err := websocket.DefaultDialer.Dial(c.URL, nil)
	c.Conn = conn
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ReadTrade() (model.Trade, error){
	messageType, message, err := c.Conn.ReadMessage()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Message Type: ", messageType)

	var agg model.AggTrade
	err = json.Unmarshal(message, &agg)
	if err != nil {
		log.Fatal(err)
	}

	trade, err := agg.ToTrade()
	if err != nil {
		log.Fatal(err)
	}
	return trade, nil
}