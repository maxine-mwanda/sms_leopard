package queue

import (
    "encoding/json"
    "log"

    "github.com/streadway/amqp"
)

type Consumer struct{
    ch *amqp.Channel
    queueName string
}

type SendJob struct{ CampaignID int64 `json:"campaign_id"` }

func NewConsumer(conn *amqp.Connection, queueName string) (*Consumer, error){
    ch, err := conn.Channel()
    if err!=nil { return nil, err }
    q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
    if err!=nil { return nil, err }
    if err := ch.QueueBind(q.Name, "", "smsleopard-exchange", false, nil); err!=nil { return nil, err }
    return &Consumer{ch: ch, queueName: q.Name}, nil
}

func (c *Consumer) Consume(handler func(SendJob) error) error{
    msgs, err := c.ch.Consume(c.queueName, "", false, false, false, false, nil)
    if err!=nil { return err }
    for d := range msgs {
        var job SendJob
        if err := json.Unmarshal(d.Body, &job); err!=nil { log.Printf("invalid job: %v", err); d.Nack(false, false); continue }
        if err := handler(job); err!=nil { log.Printf("handler err: %v", err); d.Nack(false, true); continue }
        d.Ack(false)
    }
    return nil
}
