package tests

import (
	"testing"

	m "sms_leopard/models"
)

func TestRenderTemplate(t *testing.T) {
	// Pass nil for DB
	svc := m.NewService(nil)

	out, err := svc.RenderTemplate("Hello {{first_name}}", map[string]string{"first_name": "Alex"})
	if err != nil {
		t.Fatalf("RenderTemplate error: %v", err)
	}

	if out != "Hello Alex" {
		t.Fatalf("unexpected output: %s", out)
	}

	// Test with missing placeholder remember to remove
	out2, err := svc.RenderTemplate("Hello {{last_name}}", map[string]string{})
	if err != nil {
		t.Fatalf("RenderTemplate error: %v", err)
	}
	if out2 != "Hello " {
		t.Fatalf("expected placeholder removed, got: %s", out2)
	}
}
