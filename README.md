## MCP Tools

### Deal Tools

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-deals` | List deals with flexible filters: search, status, owner, stage, pipeline, value range, date range | `searchTitle`, `status`, `ownerId`, `stageId`, `pipelineId`, `minValue`, `maxValue`, `daysBack`, `limit` |
| `get-deal` | Get a single deal by ID, including all custom fields | `dealId` (required) |
| `get-deal-notes` | Get notes and booking details for a specific deal | `dealId` (required), `limit` |
| `search-deals` | Quick search deals by term | `term` (required) |
| `search-leads` | Search leads by term | `term` (required) |

### Contact Tools
| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-persons` | List all persons (contacts) | — |
| `get-person` | Get a single person by ID, including custom fields | `personId` (required) |
| `search-persons` | Search persons by name or email | `term` (required) |

### Organization Tools
| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-organizations` | List all organizations | — |
| `get-organization` | Get a single organization by ID, including custom fields | `organizationId` (required) |
| `search-organizations` | Search organizations by name | `term` (required) |

### Activity Tools
| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-activities` | List activities with filters: owner, deal, person, org, done status. Cursor pagination | `ownerId`, `dealId`, `personId`, `orgId`, `done`, `sortBy`, `add_time`, `update_time`, `due_date`, `limit`, `cursor` |
| `get-activity` | Get a single activity by ID with full details | `activityId` (required) |
| `add-activity` | Create a new activity. Requires `subject`. Link to deal, person, org, lead. Set due date/duration/type | `subject`, `type`, `dueDate`, `dueTime`, `duration`, `done`, `busy`, `note`, `priority` |
| `update-activity` | Update an existing activity (mark done, change subject/type, due date/duration/ assignee, other fields | `id` (required), plus fields to update |
| `delete-activity` | Delete an activity by ID. Activity is soft-deleted and permanently removed after 30 days | `activityId` (required) |

### Note Tools
| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-notes` | List notes with filters: deal, person, org, user, lead. Supports pagination | `dealId`, `personId`, `orgId`, `leadId`, `userId`, `limit`, `start` |
| `get-note` | Get a single note by ID with full details | `noteId` (required) |
| `add-note` | Create a new note attached to a deal, person, or organization, or lead. Requires `content` and at least one associated entity | `content` (required), `dealId`, `personId`, `orgId`, or `leadId` |
| `update-note` | Update note content or reassign to different entity | `id` (required), `content` |
| `delete-note` | Delete a note by ID | `noteId` (required) |

### Pipeline & Stage Tools
| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-pipelines` | List all pipelines | — |
| `get-stages` | List all stages across all pipelines, enriched with `pipeline_name` | — |

### Other Tools
| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-users` | List all users/owners (useful for filtering deals by `ownerId`) | — |
| `search-leads` | Search leads by term | `term` (required) |
| `search-all` | Cross-entity search across deals, persons, organizations, leads, and more | `term` (required), `itemTypes` |

### The `get-deals` Tool in Detail

This is the most powerful tool. Here's how the filtering works:

- **With `searchTitle`**: Uses Pipedrive's search endpoint, then applies client-side filters (owner, status, stage, pipeline, value range).
- **Without `searchTitle`**: Lists deals sorted by last activity date, filtered by `status` (default: `open`), then applies date range (`daysBack`) and value range filters client-side.
- Results are capped at 30 summarized deals per response to keep payloads manageable.

