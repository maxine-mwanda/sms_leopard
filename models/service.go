package models

import (
    "database/sql"
    "strings"
    "time"
)

type Service struct{
    DB *sql.DB
}

type Customer struct{
    ID int64 `json:"id"`
    Phone string `json:"phone"`
    FirstName string `json:"first_name"`
    LastName string `json:"last_name"`
    Metadata string `json:"metadata"`
}

type Campaign struct{
    ID int64 `json:"id"`
    Name string `json:"name"`
    Template string `json:"template"`
    Status string `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

type Outbound struct{
    ID int64 `json:"id"`
    CampaignID int64 `json:"campaign_id"`
    CustomerID int64 `json:"customer_id"`
    To string `json:"to_phone"`
    Body string `json:"body"`
    Status string `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

func NewService(db *sql.DB) *Service { return &Service{DB: db} }

func (s *Service) CreateCampaign(name, template string) (*Campaign, error){
    res, err := s.DB.Exec("INSERT INTO campaigns(name, template) VALUES(?,?)", name, template)
    if err!=nil { return nil, err }
    id, _ := res.LastInsertId()
    return &Campaign{ID:id, Name:name, Template:template, Status:"draft"}, nil
}

func (s *Service) ListCampaigns(limit, offset int) ([]Campaign, error){
    rows, err := s.DB.Query("SELECT id, name, template, status, created_at FROM campaigns ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
    if err!=nil { return nil, err }
    defer rows.Close()
    out := []Campaign{}
    for rows.Next(){
        var c Campaign; var created string
        if err := rows.Scan(&c.ID, &c.Name, &c.Template, &c.Status, &created); err!=nil { return nil, err }
        c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
        out = append(out, c)
    }
    return out, nil
}

func (s *Service) GetCampaign(id int64) (*Campaign, error){
    var c Campaign; var created string
    if err := s.DB.QueryRow("SELECT id, name, template, status, created_at FROM campaigns WHERE id = ?", id).Scan(&c.ID, &c.Name, &c.Template, &c.Status, &created); err!=nil { return nil, err }
    c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
    return &c, nil
}

func (s *Service) EnqueueCampaign(campaignID int64) error{
    rows, err := s.DB.Query("SELECT id, phone FROM customers")
    if err!=nil { return err }
    defer rows.Close()
    for rows.Next(){
        var id int64; var phone string
        if err := rows.Scan(&id, &phone); err!=nil { return err }
        if _, err := s.DB.Exec("INSERT INTO outbound_messages(campaign_id, customer_id, to_phone, body, status) VALUES(?,?,?,?, 'queued')", campaignID, id, phone, ""); err!=nil { return err }
    }
    if _, err := s.DB.Exec("UPDATE campaigns SET status = 'sent' WHERE id = ?", campaignID); err!=nil { return err }
    return nil
}

func (s *Service) ListOutbound(campaignID int64) ([]Outbound, error){
    rows, err := s.DB.Query("SELECT id, campaign_id, customer_id, to_phone, body, status, created_at FROM outbound_messages WHERE campaign_id = ?", campaignID)
    if err!=nil{ return nil, err }
    defer rows.Close()
    out := []Outbound{}
    for rows.Next(){
        var m Outbound; var created string
        if err := rows.Scan(&m.ID, &m.CampaignID, &m.CustomerID, &m.To, &m.Body, &m.Status, &created); err!=nil { return nil, err }
        m.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
        out = append(out, m)
    }
    return out, nil
}

func (s *Service) RenderTemplate(tmpl string, ctx map[string]string) (string, error){
    out := tmpl
    for k,v := range ctx {
        out = strings.ReplaceAll(out, "{{"+k+"}}", v)
    }
    for {
        i := strings.Index(out, "{{")
        if i==-1 { break }
        j := strings.Index(out[i:], "}}")
        if j==-1 { break }
        out = out[:i] + out[i+j+2:]
    }
    return out, nil
}
