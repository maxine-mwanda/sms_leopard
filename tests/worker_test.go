package tests

import (
    "database/sql"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    m "sms_leopard/models"
)

func TestWorkerProcessLogic(t *testing.T){
    db, mock, err := sqlmock.New()
    if err!=nil{ t.Fatalf("mock: %v", err) }
    defer db.Close()
    svc := m.NewService(db)
    mock.ExpectQuery("SELECT id, name, template, status, created_at FROM campaigns WHERE id = ?").WillReturnRows(
        sqlmock.NewRows([]string{"id","name","template","status","created_at"}).AddRow(1, "x", "Hi {{first_name}}", "sent", "2020-01-01 00:00:00"),
    )
    mock.ExpectQuery("SELECT id, campaign_id, customer_id, to_phone, body, status, created_at FROM outbound_messages WHERE campaign_id = ?").WillReturnRows(
        sqlmock.NewRows([]string{"id","campaign_id","customer_id","to_phone","body","status","created_at"}).AddRow(10,1,2,"0711000000","","queued","2020-01-01 00:00:00"),
    )
    mock.ExpectQuery("SELECT first_name, last_name FROM customers WHERE id = ?").WillReturnRows(
        sqlmock.NewRows([]string{"first_name","last_name"}).AddRow("Sam","Lake"),
    )
    mock.ExpectExec("UPDATE outbound_messages SET body = ?, status = 'sent' WHERE id = ?").WithArgs(sqlmock.AnyArg(), 10).WillReturnResult(sqlmock.NewResult(1,1))
    out, _ := svc.RenderTemplate("Hi {{first_name}}", map[string]string{"first_name":"Sam"})
    if out != "Hi Sam" { t.Fatalf("render: %s", out) }
    if err := mock.ExpectationsWereMet(); err!=nil{ t.Fatalf("unmet: %v", err) }
}
