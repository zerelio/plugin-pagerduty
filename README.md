# Zerelio PagerDuty Plugin

Discovers PagerDuty resources and streams them into the Zerelio knowledge graph.

## What It Discovers

| Resource type                    | Description                              |
|----------------------------------|------------------------------------------|
| `pagerduty.service`              | PagerDuty services                       |
| `pagerduty.escalation_policy`    | Escalation policies                      |
| `pagerduty.team`                 | Teams                                    |
| `pagerduty.schedule`             | On-call schedules                        |
| `pagerduty.user`                 | Users                                    |

## Configuration

| Environment variable  | Required | Description                                               |
|-----------------------|----------|-----------------------------------------------------------|
| `PAGERDUTY_API_KEY`   | Yes      | REST API key — a read-only key is sufficient              |

To create an API key: go to **Integrations → API Access Keys → Create New API Key**. A read-only key is recommended.

## Adding to a Zerelio Scope

```json
{
  "scope_name": "My PagerDuty",
  "plugin_name": "pagerduty",
  "config": {
    "account_subdomain": "my-company"
  },
  "schedule": "@every 1h"
}
```

## Part of Zerelio

This plugin is part of the [Zerelio](https://github.com/zerelio/knowledge) infrastructure knowledge graph platform.
