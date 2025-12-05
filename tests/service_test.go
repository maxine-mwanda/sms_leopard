package tests

import (
    "database/sql"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    m "github.com/example/smsleopard/models"
)

func TestRenderTemplate(t *testing.T){
    db, _, _ := sqlmock.New()
    svc := m.NewService(db)
    out, _ := svc.RenderTemplate("Hello {{first_name}}", map[string]string{"first_name":"Alex"})
    if out != "Hello Alex" { t.Fatalf("unexpected: %s", out) }
}
