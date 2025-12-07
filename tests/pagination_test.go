package tests

import (
	"testing"

	m "sms_leopard/models"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func TestListCampaignsPagination(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)

		defer db.Close()

		// Create schema
		_, err = db.Exec(`
		CREATE TABLE campaigns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			template TEXT,
			status TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
		if err != nil {
			t.Fatal(err)
		}

		// Seed data
		for i := 0; i < 3; i++ {
			_, err := db.Exec(
				"INSERT INTO campaigns(name, template, status) VALUES(?, ?, ?)",
				"c", "t", "draft",
			)
			if err != nil {
				t.Fatal(err)
			}
		}

		svc := m.NewService(db)
		res, err := svc.ListCampaigns(3, 0)
		if err != nil {
			t.Fatal(err)
		}

		if len(res) != 3 {
			t.Fatalf("expected 3 campaigns, got %d", len(res))
		}
	}
}
