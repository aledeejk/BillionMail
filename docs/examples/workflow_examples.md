# Workflow Usage Examples

This document provides practical examples of how to use the BillionMail workflow system for common email marketing scenarios.

## Example 1: Welcome Series (User Onboarding)

### Workflow JSON Schema

```json
{
  "name": "Welcome Series",
  "description": "Automated onboarding emails for new subscribers",
  "status": 1,
  "version": 1,
  "trigger": "contact_created",
  "nodes": [
    {
      "id": "welcome_email_1",
      "type": "email",
      "position": {"x": 100, "y": 100},
      "data": {
        "template_id": 1,
        "subject": "Welcome to BillionMail!",
        "delay": 0
      }
    },
    {
      "id": "welcome_email_2",
      "type": "email",
      "position": {"x": 300, "y": 100},
      "data": {
        "template_id": 2,
        "subject": "Getting Started Guide",
        "delay": 86400
      }
    },
    {
      "id": "welcome_email_3",
      "type": "email",
      "position": {"x": 500, "y": 100},
      "data": {
        "template_id": 3,
        "subject": "Advanced Features",
        "delay": 259200
      }
    }
  ],
  "connections": [
    {
      "id": "conn_1",
      "source": "welcome_email_1",
      "target": "welcome_email_2",
      "sourceHandle": "output",
      "targetHandle": "input"
    },
    {
      "id": "conn_2",
      "source": "welcome_email_2",
      "target": "welcome_email_3",
      "sourceHandle": "output",
      "targetHandle": "input"
    }
  ]
}
```

### Creating via API

```bash
curl -X POST http://localhost:8080/api/v1/workflow \
  -H "Content-Type: application/json" \
  -d @welcome_workflow.json
```

### Triggering the Workflow

The workflow is automatically triggered when a new contact is created in the system. No manual triggering required.

## Example 2: Re-engagement for Inactive Users

### Workflow JSON Schema

```json
{
  "name": "Re-engagement Campaign",
  "description": "Re-engage users who haven't opened emails in 30 days",
  "status": 1,
  "version": 1,
  "trigger": "scheduled",
  "nodes": [
    {
      "id": "check_activity",
      "type": "condition",
      "position": {"x": 100, "y": 100},
      "data": {
        "condition": "last_activity > 30 days"
      }
    },
    {
      "id": "reengagement_email",
      "type": "email",
      "position": {"x": 300, "y": 100},
      "data": {
        "template_id": 4,
        "subject": "We Miss You! 🎉",
        "delay": 0
      }
    },
    {
      "id": "follow_up",
      "type": "email",
      "position": {"x": 500, "y": 100},
      "data": {
        "template_id": 5,
        "subject": "Last Chance - Special Offer",
        "delay": 604800,
        "condition": "not_opened"
      }
    }
  ],
  "connections": [
    {
      "id": "conn_1",
      "source": "check_activity",
      "target": "reengagement_email",
      "sourceHandle": "true",
      "targetHandle": "input"
    },
    {
      "id": "conn_2",
      "source": "reengagement_email",
      "target": "follow_up",
      "sourceHandle": "output",
      "targetHandle": "input"
    }
  ]
}
```

### Creating via API

```bash
curl -X POST http://localhost:8080/api/v1/workflow \
  -H "Content-Type: application/json" \
  -d @reengagement_workflow.json
```

### Triggering the Workflow

This workflow runs on a schedule (e.g., weekly) and checks for inactive users:

```bash
curl -X POST http://localhost:8080/api/v1/workflow/{workflow_id}/execute \
  -H "Content-Type: application/json" \
  -d '{"trigger": "scheduled", "parameters": {"inactive_days": 30}}'
```

## Example 3: Birthday Greetings

### Workflow JSON Schema

```json
{
  "name": "Birthday Greetings",
  "description": "Automated birthday emails for subscribers",
  "status": 1,
  "version": 1,
  "trigger": "scheduled",
  "nodes": [
    {
      "id": "birthday_check",
      "type": "condition",
      "position": {"x": 100, "y": 100},
      "data": {
        "condition": "birthday == today"
      }
    },
    {
      "id": "birthday_email",
      "type": "email",
      "position": {"x": 300, "y": 100},
      "data": {
        "template_id": 6,
        "subject": "Happy Birthday! 🎂",
        "delay": 0,
        "personalization": {
          "first_name": "{{contact.first_name}}",
          "age": "{{contact.age}}"
        }
      }
    },
    {
      "id": "birthday_discount",
      "type": "action",
      "position": {"x": 500, "y": 100},
      "data": {
        "action": "apply_discount",
        "discount_code": "BDAY{{contact.id}}",
        "percentage": 20,
        "delay": 3600
      }
    }
  ],
  "connections": [
    {
      "id": "conn_1",
      "source": "birthday_check",
      "target": "birthday_email",
      "sourceHandle": "true",
      "targetHandle": "input"
    },
    {
      "id": "conn_2",
      "source": "birthday_email",
      "target": "birthday_discount",
      "sourceHandle": "output",
      "targetHandle": "input"
    }
  ]
}
```

### Creating via API

```bash
curl -X POST http://localhost:8080/api/v1/workflow \
  -H "Content-Type: application/json" \
  -d @birthday_workflow.json
```

### Triggering the Workflow

This workflow runs daily and checks for birthdays:

```bash
curl -X POST http://localhost:8080/api/v1/workflow/{workflow_id}/execute \
  -H "Content-Type: application/json" \
  -d '{"trigger": "daily_birthday_check"}'
```

## Common Workflow Patterns

### Node Types
- **email**: Send email with template
- **condition**: Check conditions (time-based, user behavior, etc.)
- **action**: Perform actions (apply tags, update contact, etc.)
- **delay**: Wait for specified time
- **split**: Split workflow based on conditions

### Triggers
- **contact_created**: When new contact is added
- **contact_updated**: When contact data changes
- **scheduled**: Time-based execution
- **event**: Custom events from your application

### Conditions
- Time-based: `last_activity > 30 days`
- Behavior: `not_opened`, `clicked_link`
- Data: `birthday == today`, `tags contains 'vip'`

### Actions
- Send email
- Apply/remove tags
- Update contact fields
- Apply discounts/coupons
- Trigger webhooks