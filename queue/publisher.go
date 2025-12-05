package queue

import (
    "encoding/json"

    "github.com/streadway/amqp"
)

type Publisher struct{
    ch *amqp.Channel
    exchange string
}

func NewPublisher(conn *amqp.Connection, exchange string) (*Publisher, error){
    ch, err := conn.Channel()
    if err!=nil { return nil, err }
    if err := ch.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil); err!=nil { return nil, err }
    return &Publisher{ch: ch, exchange: exchange}, nil
}

func (p *Publisher) PublishSend(campaignID int64) error{
    body, _ := json.Marshal(map[string]int64{"campaign_id": campaignID})
    return p.ch.Publish(p.exchange, "", false, false, amqp.Publishing{
        DeliveryMode: amqp.Persistent,
        ContentType: "application/json",
        Body: body,
    })
}
