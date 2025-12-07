package models

import (
	"database/sql"
	"strings"
	"time"
)

type Service struct {
	DB *sql.DB
}

type Customer struct {
	ID        int64  `json:"id"`
	Phone     string `json:"phone"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Metadata  string `json:"metadata"`
}

type Campaign struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Template  string    `json:"template"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Outbound struct {
	ID         int64     `json:"id"`
	CampaignID int64     `json:"campaign_id"`
	CustomerID int64     `json:"customer_id"`
	To         string    `json:"to_phone"`
	Body       string    `json:"body"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type CampaignDetails struct {
	ID           int64          `json:"id"`
	Name         string         `json:"name"`
	Channel      string         `json:"channel"`
	Status       string         `json:"status"`
	BaseTemplate string         `json:"base_template"`
	ScheduledAt  *time.Time     `json:"scheduled_at"`
	CreatedAt    time.Time      `json:"created_at"`
	Stats        map[string]int `json:"stats"`
}

func NewService(db *sql.DB) *Service { return &Service{DB: db} }

func (s *Service) CreateCampaign(name, template string) (*Campaign, error) {
	res, err := s.DB.Exec("INSERT INTO campaigns(name, template) VALUES(?,?)", name, template)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Campaign{ID: id, Name: name, Template: template, Status: "draft"}, nil
}

func (s *Service) ListCampaigns(limit, offset int) ([]Campaign, error) {
	rows, err := s.DB.Query("SELECT id, name, template, status, created_at FROM campaigns ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Campaign{}
	for rows.Next() {
		var c Campaign
		var created string
		if err := rows.Scan(&c.ID, &c.Name, &c.Template, &c.Status, &created); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		out = append(out, c)
	}
	return out, nil
}

func (s *Service) GetCampaign(id int64) (*Campaign, error) {
	var c Campaign
	var created string
	if err := s.DB.QueryRow("SELECT id, name, template, status, created_at FROM campaigns WHERE id = ?", id).Scan(&c.ID, &c.Name, &c.Template, &c.Status, &created); err != nil {
		return nil, err
	}
	c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	return &c, nil
}

func (s *Service) EnqueueCampaign(campaignID int64) error {
	rows, err := s.DB.Query("SELECT id, phone FROM customers")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var phone string
		if err := rows.Scan(&id, &phone); err != nil {
			return err
		}
		if _, err := s.DB.Exec("INSERT INTO outbound_messages(campaign_id, customer_id, to_phone, body, status) VALUES(?,?,?,?, 'queued')", campaignID, id, phone, ""); err != nil {
			return err
		}
	}
	if _, err := s.DB.Exec("UPDATE campaigns SET status = 'sent' WHERE id = ?", campaignID); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListOutbound(campaignID int64) ([]Outbound, error) {
	rows, err := s.DB.Query("SELECT id, campaign_id, customer_id, to_phone, body, status, created_at FROM outbound_messages WHERE campaign_id = ?", campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Outbound{}
	for rows.Next() {
		var m Outbound
		var created string
		if err := rows.Scan(&m.ID, &m.CampaignID, &m.CustomerID, &m.To, &m.Body, &m.Status, &created); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		out = append(out, m)
	}
	return out, nil
}

func (s *Service) RenderTemplate(tmpl string, ctx map[string]string) (string, error) {
	out := tmpl
	for k, v := range ctx {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	for {
		i := strings.Index(out, "{{")
		if i == -1 {
			break
		}
		j := strings.Index(out[i:], "}}")
		if j == -1 {
			break
		}
		out = out[:i] + out[i+j+2:]
	}
	return out, nil
}

// campaign details
// GetCampaignDetails returns campaign info with basic stats
func (s *Service) GetCampaignDetails(id int64) (map[string]interface{}, bool, error) {
	c, err := s.GetCampaign(id)
	if err != nil {
		return nil, false, err
	}

	rows, err := s.DB.Query(
		"SELECT status, COUNT(*) FROM outbound_messages WHERE campaign_id = ? GROUP BY status", id,
	)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	stats := map[string]int{"pending": 0, "sent": 0, "failed": 0, "total": 0}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, false, err
		}
		stats[status] = count
		stats["total"] += count
	}

	return map[string]interface{}{
		"id":         c.ID,
		"name":       c.Name,
		"template":   c.Template,
		"status":     c.Status,
		"created_at": c.CreatedAt,
		"stats":      stats,
	}, true, nil
}

// GetStats returns aggregate info for /stats endpoint
func (s *Service) GetStats() (map[string]int, error) {
	rows, err := s.DB.Query("SELECT status, COUNT(*) FROM campaigns GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{"draft": 0, "scheduled": 0, "sending": 0, "sent": 0, "failed": 0}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		out[status] = count
	}
	return out, nil
}

// Ping checks database connectivity for /health endpoint
func (s *Service) Ping() error {
	return s.DB.Ping()
}
