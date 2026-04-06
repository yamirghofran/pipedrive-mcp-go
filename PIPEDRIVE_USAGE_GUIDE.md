# Pipedrive MCP — Practical Usage Guide

## Quick Reference

| Tool | Purpose |
|------|---------|
| `get-deals` | **Primary tool** — list deals with filters (status, owner, pipeline, stage, value range, date range). Returns all non-deleted deals by default. |
| `get-deal` | Get a single deal by ID (includes custom fields) |
| `get-deal-notes` | Get notes/booking details for a specific deal |
| `search-deals` | Full-text search deals by keyword (when you need fuzzy/partial matching) |
| `get-pipelines` | List all pipelines |
| `get-stages` | List all stages across pipelines |
| `get-users` | List all users/owners (needed for `ownerId` filters) |
| `get-organizations` | List all orgs |
| `get-organization` | Get a single org by ID |
| `get-persons` | List all persons/contacts |
| `get-person` | Get a single person by ID |
| `search-all` | Global search across deals, persons, orgs, products, files, activities, leads |
| `search-leads` | Search **leads** by term (leads != deals — use `get-deals` for deals) |
| `search-organizations` | Search orgs by term |
| `search-persons` | Search persons by term |

## Key Concepts

### 1. `get-deals` vs `search-deals`

These serve **different purposes**:

- **`get-deals`** — Structured listing with filters. Use this for "show me deals", dashboards, reports, filtered lists. Returns all non-deleted deals by default.
- **`search-deals`** — Full-text keyword search. Use this when you need fuzzy/partial matching on deal titles, notes, or custom fields. Requires a `term` parameter.

**Rule of thumb:** Use `get-deals` unless you specifically need keyword search.

### 2. `status` parameter — defaults to all deals

The `status` filter on `get-deals` accepts:
- *(omitted)* — returns **all non-deleted deals** (open, won, lost). This is the default.
- `"open"` — active deals only
- `"won"` — closed-won
- `"lost"` — closed-lost
- `"deleted"` — soft-deleted

Multiple statuses can be comma-separated (e.g., `"open,won"`).

### 3. Filter IDs must come from lookup calls

`get-deals` accepts `ownerId`, `pipelineId`, and `stageId` as numeric IDs. You must first call:

- **`get-users`** → gives you `id` for each user (use as `ownerId`)
- **`get-pipelines`** → gives you `id` for each pipeline (use as `pipelineId`)
- **`get-stages`** → gives you `id` for each stage (use as `stageId`)

### 4. `get-deal` requires the numeric deal ID

The deal `id` is a plain integer (e.g., `131`, `210`), **not** a string. You can find IDs from `get-deals` or `search-deals` results.

### 5. `daysBack` filters by update time

The `daysBack` parameter filters based on **`update_time`** (when the deal was last modified), not creation date or activity date. Default is 365 days. A deal created 2 years ago but edited yesterday will appear with `daysBack=7`.

Set `daysBack=0` or a very large number (e.g., `daysBack=3650`) if you want to include very old deals.

### 6. Value filters accept numbers only

`minValue` and `maxValue` on `get-deals` are numeric. They match the deal's monetary value regardless of currency.

### 7. Search tools — when to use which

| Search Tool | Scope | Best For |
|-------------|-------|----------|
| `search-all` | Deals, persons, orgs, products, files, activities, leads | Broad discovery when you don't know the entity type |
| `search-deals` | Deals only | Fuzzy/keyword search on deal titles, notes, custom fields |
| `search-organizations` | Orgs only | Looking up a company by name |
| `search-persons` | Persons only | Looking up a contact by name |
| `search-leads` | Leads only | Searches **leads** (unconverted prospects), NOT deals |

### 8. Pagination

`get-deals` accepts a `limit` parameter (capped at 500). For pagination, use `daysBack` to narrow windows or `ownerId`/`pipelineId` to partition results.

## Typical Workflows

### "Show me all deals"

```
get-deals()                 # Returns all non-deleted deals by default
```

### "Show me only open deals"

```
get-deals(status="open")
```

### "Show me my deals"

```
get-users()                 # Find your user ID
get-deals(ownerId=<yourId>)
```

### "Get details + notes for a deal"

```
get-deal(dealId=131)
get-deal-notes(dealId=131)
```

### "Filter deals by pipeline stage"

```
get-stages()                # Find the stage ID
get-deals(stageId=<id>)
```

### "Find everything about an organization"

```
search-all(term="Movistar")  # Returns matching deals, persons, and orgs
```

### "High-value open deals"

```
get-deals(minValue=50000, status="open")
```

### "Deals closed this quarter"

```
get-deals(status="won", daysBack=90)
```

## Data Model Quick Reference

```
Pipeline 1:N Stage
Stage    1:N Deal
User     1:N Deal      (owner)
Org      1:N Deal
Org      1:N Person
Person   1:N Deal
Deal     1:N Note
```
