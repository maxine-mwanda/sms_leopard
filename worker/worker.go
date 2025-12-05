package worker

import (
    "database/sql"
    "log"

    "github.com/example/smsleopard/models"
    "github.com/example/smsleopard/queue"
)

type Worker struct{
    svc *models.Service
    consumer *queue.Consumer
}

func NewWorker(svc *models.Service, consumer *queue.Consumer) *Worker {
    return &Worker{svc: svc, consumer: consumer}
}

func (w *Worker) Start() error {
    return w.consumer.Consume(func(job queue.SendJob) error {
        log.Printf("processing campaign %d", job.CampaignID)
        return w.process(job.CampaignID)
    })
}

func (w *Worker) process(campaignID int64) error {
    c, err := w.svc.GetCampaign(campaignID)
    if err!=nil { return err }
    outs, err := w.svc.ListOutbound(campaignID)
    if err!=nil { return err }
    for _, m := range outs {
        if m.Status != "queued" { continue }
        var first, last string
        row := w.svc.DB.QueryRow("SELECT first_name, last_name FROM customers WHERE id = ?", m.CustomerID)
        var fn, ln sql.NullString
        if err := row.Scan(&fn, &ln); err==nil {
            if fn.Valid { first = fn.String }
            if ln.Valid { last = ln.String }
        }
        body, _ := w.svc.RenderTemplate(c.Template, map[string]string{"first_name": first, "last_name": last})
        log.Printf("send to %s: %s", m.To, body)
        if _, err := w.svc.DB.Exec("UPDATE outbound_messages SET body = ?, status = 'sent' WHERE id = ?", body, m.ID); err!=nil {
            return err
        }
    }
    return nil
}
