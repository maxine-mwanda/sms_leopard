package tests

import (
    //"database/sql"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    m "sms_leopard/models"
)

func TestListCampaignsPagination(t *testing.T){
    db, mock, err := sqlmock.New()
    if err!=nil{ t.Fatalf("mock: %v", err) }
    defer db.Close()
    rows := sqlmock.NewRows([]string{"id","name","template","status","created_at"})
    for i:=0;i<3;i++{ rows.AddRow(i+1, "c","t","draft","2020-01-01 00:00:00") }
    mock.ExpectQuery("SELECT id, name, template, status, created_at FROM campaigns").WillReturnRows(rows)
    svc := m.NewService(db)
    _, err = svc.ListCampaigns(3,0)
    if err!=nil{ t.Fatalf("list: %v", err) }
    if err := mock.ExpectationsWereMet(); err!=nil{ t.Fatalf("unmet: %v", err) }
}
