package main

import (
	"context"
	"testing"
)

func TestPagerDutyPlugin_ValidateConfig(t *testing.T) {
	p := &PagerDutyPlugin{}
	r, _ := p.ValidateConfig(context.Background(), map[string]string{})
	if r.Valid { t.Fatal("expected invalid for empty config") }
	r, _ = p.ValidateConfig(context.Background(), map[string]string{"api_token": "test"})
	if !r.Valid { t.Fatal("expected valid with api_token") }
}

func TestPagerDutyPlugin_Info(t *testing.T) {
	p := &PagerDutyPlugin{}
	info, _ := p.Info(context.Background())
	if info.Name != "pagerduty" { t.Fatalf("expected pagerduty, got %s", info.Name) }
}
