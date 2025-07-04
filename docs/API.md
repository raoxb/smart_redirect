# Smart Redirect API Documentation

This document provides comprehensive documentation for the Smart Redirect API.

## Base URL

```
https://api.domain.com
```

## Authentication

Most API endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Rate Limiting

- **Redirect endpoints**: 1000 requests/hour per IP
- **API endpoints**: 100 requests/hour per IP
- **Admin endpoints**: 50 requests/hour per IP

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Request limit per window
- `X-RateLimit-Remaining`: Requests remaining in current window
- `X-RateLimit-Reset`: Time when window resets (Unix timestamp)

## Error Handling

All errors follow this format:

```json
{
  "error": "Description of the error",
  "code": "ERROR_CODE",
  "details": {}
}
```

### HTTP Status Codes

- `200`: Success
- `201`: Created
- `400`: Bad Request
- `401`: Unauthorized
- `403`: Forbidden
- `404`: Not Found
- `409`: Conflict
- `429`: Too Many Requests
- `500`: Internal Server Error

## Endpoints

### Health Check

#### GET /health

Check service health status.

**Response:**
```json
{
  "status": "ok",
  "time": 1640995200
}
```

---

## Redirect Service

### GET /v1/{bu}/{link_id}

Perform redirect to target URL.

**Parameters:**
- `bu` (path): Business unit (e.g., `bu01`, `bu02`)
- `link_id` (path): 6-character link identifier
- `network` (query): Channel identifier (e.g., `mi`, `google`, `fb`)

**Example:**
```
GET /v1/bu01/abc123?network=mi&kw=golang
```

**Responses:**
- `302`: Redirect to target URL
- `404`: Link not found
- `429`: Rate limit exceeded
- `503`: No available targets

---

## Authentication

### POST /api/v1/auth/register

Register a new user account.

**Request Body:**
```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "id": 1,
  "username": "newuser",
  "email": "user@example.com",
  "role": "user"
}
```

### POST /api/v1/auth/login

Authenticate user and get JWT token.

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": 1,
  "username": "testuser",
  "role": "user"
}
```

### GET /api/v1/auth/profile

Get current user profile. Requires authentication.

**Response:**
```json
{
  "id": 1,
  "username": "testuser",
  "email": "user@example.com",
  "role": "user"
}
```

---

## Link Management

### POST /api/v1/links

Create a new short link. Requires authentication.

**Request Body:**
```json
{
  "business_unit": "bu01",
  "network": "mi",
  "total_cap": 1000,
  "backup_url": "https://backup.example.com"
}
```

**Response:**
```json
{
  "id": 1,
  "link_id": "abc123",
  "business_unit": "bu01",
  "network": "mi",
  "total_cap": 1000,
  "current_hits": 0,
  "backup_url": "https://backup.example.com",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### GET /api/v1/links

List all links with pagination. Requires authentication.

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20, max: 100)

**Response:**
```json
{
  "total": 100,
  "page": 1,
  "size": 20,
  "data": [
    {
      "id": 1,
      "link_id": "abc123",
      "business_unit": "bu01",
      "network": "mi",
      "total_cap": 1000,
      "current_hits": 50,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/links/{link_id}

Get link details by ID. Requires authentication.

**Response:**
```json
{
  "id": 1,
  "link_id": "abc123",
  "business_unit": "bu01",
  "network": "mi",
  "total_cap": 1000,
  "current_hits": 50,
  "backup_url": "https://backup.example.com",
  "is_active": true,
  "targets": [
    {
      "id": 1,
      "url": "https://target1.example.com",
      "weight": 70,
      "cap": 500,
      "current_hits": 25,
      "countries": ["US", "CA"],
      "is_active": true
    }
  ]
}
```

### PUT /api/v1/links/{link_id}

Update link configuration. Requires authentication.

**Request Body:**
```json
{
  "business_unit": "bu02",
  "network": "google",
  "total_cap": 2000,
  "backup_url": "https://new-backup.example.com"
}
```

### DELETE /api/v1/links/{link_id}

Delete a link. Requires authentication.

**Response:**
```json
{
  "message": "link deleted successfully"
}
```

---

## Target Management

### POST /api/v1/links/{link_id}/targets

Add a target to a link. Requires authentication.

**Request Body:**
```json
{
  "url": "https://target.example.com",
  "weight": 50,
  "cap": 300,
  "countries": ["US", "CA", "UK"],
  "param_mapping": {
    "kw": "q",
    "src": "source"
  },
  "static_params": {
    "ref": "campaign1",
    "utm_source": "redirect"
  }
}
```

**Response:**
```json
{
  "id": 1,
  "link_id": 1,
  "url": "https://target.example.com",
  "weight": 50,
  "cap": 300,
  "current_hits": 0,
  "countries": ["US", "CA", "UK"],
  "param_mapping": "{\"kw\":\"q\",\"src\":\"source\"}",
  "static_params": "{\"ref\":\"campaign1\",\"utm_source\":\"redirect\"}",
  "is_active": true
}
```

### GET /api/v1/links/{link_id}/targets

Get all targets for a link. Requires authentication.

### PUT /api/v1/targets/{target_id}

Update target configuration. Requires authentication.

### DELETE /api/v1/targets/{target_id}

Delete a target. Requires authentication.

---

## Batch Operations

### POST /api/v1/batch/links

Create multiple links in batch. Requires authentication.

**Request Body:**
```json
{
  "links": [
    {
      "business_unit": "bu01",
      "network": "mi",
      "total_cap": 1000,
      "backup_url": "https://backup1.example.com",
      "targets": [
        {
          "url": "https://target1.example.com",
          "weight": 100,
          "cap": 500,
          "countries": ["US"]
        }
      ]
    }
  ]
}
```

**Response:**
```json
{
  "success": [
    {
      "index": 0,
      "link_id": "abc123",
      "link_url": "api.domain.com/v1/bu01/abc123?network=mi"
    }
  ],
  "errors": []
}
```

### PUT /api/v1/batch/links

Update multiple links in batch. Requires authentication.

### DELETE /api/v1/batch/links

Delete multiple links in batch. Requires authentication.

**Request Body:**
```json
{
  "link_ids": ["abc123", "def456", "ghi789"]
}
```

### POST /api/v1/batch/import

Import links from CSV file. Requires authentication.

**Request:** Multipart form with `file` field containing CSV.

**CSV Format:**
```csv
business_unit,network,total_cap,backup_url,target_url,weight,cap,countries
bu01,mi,1000,https://backup.com,https://target.com,100,500,US;CA
```

### GET /api/v1/batch/export

Export all links to CSV format. Requires authentication.

---

## Templates

### POST /api/v1/templates

Create a link template. Requires authentication.

**Request Body:**
```json
{
  "name": "E-commerce Template",
  "description": "Template for e-commerce campaigns",
  "business_unit": "bu01",
  "network": "mi",
  "total_cap": 1000,
  "backup_url": "https://backup.example.com",
  "targets": [
    {
      "url": "https://shop.example.com",
      "weight": 70,
      "cap": 700,
      "countries": ["US", "CA"],
      "param_mapping": {"kw": "q"},
      "static_params": {"ref": "template"}
    }
  ]
}
```

### GET /api/v1/templates

List all templates with pagination. Requires authentication.

### GET /api/v1/templates/{id}

Get template details by ID. Requires authentication.

### PUT /api/v1/templates/{id}

Update template configuration. Requires authentication.

### DELETE /api/v1/templates/{id}

Delete a template. Requires authentication.

### POST /api/v1/templates/create-links

Create multiple links from a template. Requires authentication.

**Request Body:**
```json
{
  "template_id": 1,
  "count": 5,
  "overrides": {
    "network": "google",
    "total_cap": 2000
  }
}
```

---

## Statistics

### GET /api/v1/stats/links/{link_id}

Get comprehensive statistics for a link. Requires authentication.

**Response:**
```json
{
  "link_id": "abc123",
  "business_unit": "bu01",
  "total_hits": 1250,
  "today_hits": 45,
  "unique_ips": 320,
  "countries": [
    {"country": "US", "hits": 800},
    {"country": "CA", "hits": 300},
    {"country": "UK", "hits": 150}
  ],
  "targets": [
    {"target_id": 1, "url": "https://target1.com", "hits": 875},
    {"target_id": 2, "url": "https://target2.com", "hits": 375}
  ]
}
```

### GET /api/v1/stats/links/{link_id}/hourly

Get hourly statistics for a link. Requires authentication.

**Query Parameters:**
- `hours` (optional): Number of hours to include (default: 24, max: 168)

**Response:**
```json
[
  {"hour": "2024-01-01T00:00:00Z", "hits": 45},
  {"hour": "2024-01-01T01:00:00Z", "hits": 38},
  {"hour": "2024-01-01T02:00:00Z", "hits": 52}
]
```

### GET /api/v1/stats/system

Get system-wide statistics. Requires authentication.

**Response:**
```json
{
  "total_links": 150,
  "total_hits": 50000,
  "today_hits": 1200,
  "unique_ips": 8500,
  "top_countries": [
    {"country": "US", "hits": 25000},
    {"country": "CA", "hits": 12000},
    {"country": "UK", "hits": 8000}
  ]
}
```

---

## User Management (Admin Only)

### POST /api/v1/users

Create a new user. Requires admin authentication.

**Request Body:**
```json
{
  "username": "newuser",
  "email": "user@example.com",
  "password": "password123",
  "role": "user"
}
```

### GET /api/v1/users

List all users with pagination. Requires admin authentication.

### GET /api/v1/users/{id}

Get user details by ID. Requires admin authentication.

### PUT /api/v1/users/{id}

Update user information. Requires admin authentication.

**Request Body:**
```json
{
  "email": "newemail@example.com",
  "role": "admin",
  "is_active": false
}
```

### DELETE /api/v1/users/{id}

Delete a user. Requires admin authentication.

### POST /api/v1/users/{id}/links

Assign link permissions to a user. Requires admin authentication.

**Request Body:**
```json
{
  "link_id": 1,
  "can_edit": true,
  "can_delete": false
}
```

### GET /api/v1/users/{id}/links

Get link permissions for a user. Requires admin authentication.

---

## IP Management (Admin Only)

### GET /api/v1/stats/ip/{ip}

Get detailed information about an IP address. Requires admin authentication.

**Response:**
```json
{
  "ip": "192.168.1.100",
  "access_count": 150,
  "last_access": "2024-01-01T12:00:00Z",
  "country": "US",
  "is_blocked": false,
  "block_reason": "",
  "recent_logs": [
    {
      "id": 1,
      "link_id": 1,
      "target_id": 1,
      "created_at": "2024-01-01T11:55:00Z"
    }
  ]
}
```

### POST /api/v1/stats/ip/{ip}/block

Block an IP address. Requires admin authentication.

**Request Body:**
```json
{
  "reason": "Suspicious activity detected",
  "duration": 24
}
```

### DELETE /api/v1/stats/ip/{ip}/block

Unblock an IP address. Requires admin authentication.

---

## Webhooks (Optional)

Configure webhooks to receive real-time notifications:

### Link Events
```json
{
  "event": "link.redirect",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "link_id": "abc123",
    "target_url": "https://target.example.com",
    "ip": "192.168.1.100",
    "country": "US",
    "user_agent": "Mozilla/5.0..."
  }
}
```

### Rate Limit Events
```json
{
  "event": "rate_limit.exceeded",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "ip": "192.168.1.100",
    "limit_type": "ip_hourly",
    "current_count": 101,
    "limit": 100
  }
}
```

## SDKs and Examples

### cURL Examples

**Create a link:**
```bash
curl -X POST https://api.domain.com/api/v1/links \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "business_unit": "bu01",
    "network": "mi",
    "total_cap": 1000
  }'
```

**Batch import:**
```bash
curl -X POST https://api.domain.com/api/v1/batch/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@links.csv"
```

### JavaScript SDK

```javascript
const SmartRedirect = require('smart-redirect-sdk');

const client = new SmartRedirect({
  baseURL: 'https://api.domain.com',
  token: 'your-jwt-token'
});

// Create a link
const link = await client.links.create({
  business_unit: 'bu01',
  network: 'mi',
  total_cap: 1000
});

// Get statistics
const stats = await client.stats.getLink(link.link_id);
```

### Python SDK

```python
from smart_redirect import SmartRedirectClient

client = SmartRedirectClient(
    base_url='https://api.domain.com',
    token='your-jwt-token'
)

# Create a link
link = client.links.create(
    business_unit='bu01',
    network='mi',
    total_cap=1000
)

# Get statistics
stats = client.stats.get_link(link['link_id'])
```