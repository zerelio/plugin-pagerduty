package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	plugintypes "github.com/zerelio/plugin-sdk/plugin"
	sdk "github.com/zerelio/plugin-sdk"
)

const (
	pluginName    = "pagerduty"
	pluginVersion = "0.1.0"
)

type PagerDutyPlugin struct {
	sdk.BasePlugin
}

type pagerDutyConfig struct {
	APIToken string `json:"api_token"`
}

type pagerDutyServicesResponse struct {
	Services []pagerDutyService `json:"services"`
}

type pagerDutyService struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (p *PagerDutyPlugin) Info(ctx context.Context) (*plugintypes.PluginInfo, error) {
	return &plugintypes.PluginInfo{
		Name:        pluginName,
		Version:     pluginVersion,
		Description: "Discovers PagerDuty services and escalation policies",
		Capabilities: []string{
			"incident_management",
			"oncall",
			"alerting",
		},
		ConfigSchema: `{
			"type": "object",
			"properties": {
				"api_token": {
					"type": "string",
					"description": "PagerDuty API token"
				}
			},
			"required": ["api_token"]
		}`,
		CredentialTypes: []string{"pagerduty_api_token"},
	}, nil
}

func (p *PagerDutyPlugin) Discover(ctx context.Context, req *plugintypes.DiscoverRequest) (<-chan plugintypes.DiscoveryEvent, error) {
	ch := make(chan plugintypes.DiscoveryEvent, 100)

	go func() {
		defer close(ch)

		var cfg pagerDutyConfig
		if err := sdk.ParseConfig(req.Config, &cfg); err != nil {
			ch <- sdk.SendError("CONFIG_ERROR", fmt.Sprintf("failed to parse config: %v", err), false)
			return
		}

		if cfg.APIToken == "" {
			ch <- sdk.SendError("CONFIG_ERROR", "api_token is required", false)
			return
		}

		p.LogInfo("starting PagerDuty discovery")
		ch <- sdk.SendProgress("Connecting to PagerDuty API", 5, 0, 0)

		client := &http.Client{Timeout: 30 * time.Second}
		var resourcesFound int32

		services, err := fetchPagerDutyServices(ctx, client, cfg.APIToken)
		if err != nil {
			ch <- sdk.SendError("API_ERROR", fmt.Sprintf("failed to fetch services: %v", err), false)
			return
		}

		ch <- sdk.SendProgress(fmt.Sprintf("Found %d services", len(services)), 30, resourcesFound, 0)

		for i, svc := range services {
			select {
			case <-ctx.Done():
				ch <- sdk.SendError("CANCELLED", "discovery cancelled", false)
				return
			default:
			}

			resource := sdk.NewResourceWithID(
				fmt.Sprintf("pagerduty:service:%s", svc.ID),
				"pagerduty",
				"pagerduty.service",
				"",
				"",
				svc.Name,
			)
			raw, _ := json.Marshal(svc)
			resource.RawData = raw
			resource.Properties = map[string]string{
				"service_id":   svc.ID,
				"status":       svc.Status,
				"description":  svc.Description,
			}
			sdk.EmitResource(ch, resource)
			resourcesFound++

			pct := int32(30 + (i+1)*70/len(services))
			ch <- sdk.SendProgress(fmt.Sprintf("Processed service %s", svc.Name), pct, resourcesFound, 0)
		}

		ch <- sdk.SendProgress(
			fmt.Sprintf("PagerDuty discovery complete: %d services", resourcesFound),
			100, resourcesFound, 0,
		)
		p.LogInfo("PagerDuty discovery complete", "services", resourcesFound)
	}()

	return ch, nil
}

func fetchPagerDutyServices(ctx context.Context, client *http.Client, apiToken string) ([]pagerDutyService, error) {
	url := "https://api.pagerduty.com/services"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", apiToken))
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result pagerDutyServicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Services, nil
}

func (p *PagerDutyPlugin) ValidateConfig(ctx context.Context, config map[string]string) (*plugintypes.ValidationResult, error) {
	var cfg pagerDutyConfig
	if err := sdk.ParseConfig(config, &cfg); err != nil {
		return &plugintypes.ValidationResult{Valid: false, Errors: []string{fmt.Sprintf("invalid config: %v", err)}}, nil
	}

	var errors []string
	if cfg.APIToken == "" {
		errors = append(errors, "api_token is required")
	}

	return &plugintypes.ValidationResult{Valid: len(errors) == 0, Errors: errors}, nil
}

var _ plugintypes.DiscoveryPlugin = (*PagerDutyPlugin)(nil)
