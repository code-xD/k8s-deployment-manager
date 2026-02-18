# Schema

This document describes the database schema for the K8s Deployment Manager system.

## Tables

### users

Stores user information. Users are auto-created when a new `user_external_id` is encountered in requests.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique identifier |
| user_external_id | VARCHAR | UNIQUE, NOT NULL | External user identifier |
| created_on | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_on | TIMESTAMP | NULLABLE | Last update timestamp |

**Indexes:**
- `user_external_id` - Unique index for user lookup

### deployment_requests

Stores deployment requests raised by users. Multiple requests can exist for the same deployment.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique identifier |
| request_id | VARCHAR | UNIQUE, NOT NULL | Request ID for idempotency |
| identifier | VARCHAR(63) | NOT NULL | Deployment identifier (used as deployment name) |
| name | VARCHAR(255) | NOT NULL | Deployment name |
| namespace | VARCHAR(255) | NOT NULL | Kubernetes namespace |
| request_type | VARCHAR(50) | NOT NULL | Type: CREATE, UPDATE, DELETE |
| user_id | UUID | NOT NULL, FOREIGN KEY → users.id | Owner user |
| status | VARCHAR(50) | NOT NULL | Status: CREATED, SUCCESS, FAILURE |
| failure_reason | TEXT | NULLABLE | Failure reason if status is FAILURE |
| image | VARCHAR(255) | NOT NULL | Container image |
| metadata | JSONB | NULLABLE | Additional metadata |
| created_on | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_on | TIMESTAMP | NULLABLE | Last update timestamp |

**Indexes:**
- `request_id` - Unique index for idempotency
- `idx_deployment_request_name_namespace` - Composite index on (namespace, name)
- `idx_deployment_request_user_status` - Composite index on (user_id, status)

**Foreign Keys:**
- `user_id` → `users.id`

### deployments

Stores deployment information synced from Kubernetes. Kubernetes is the single source of truth.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique identifier |
| identifier | VARCHAR(63) | UNIQUE, NOT NULL | Deployment identifier (used as deployment name) |
| name | VARCHAR(255) | NULLABLE | Deployment name |
| namespace | VARCHAR(255) | NULLABLE | Kubernetes namespace |
| image | VARCHAR(255) | NULLABLE | Container image |
| status | VARCHAR(50) | NOT NULL | Status: INITIATED, CREATED, UPDATING, DELETED |
| user_id | UUID | NOT NULL, FOREIGN KEY → users.id | Owner user |
| resource_version | VARCHAR(255) | NULLABLE | Kubernetes resource version |
| metadata | JSONB | NULLABLE | Additional metadata |
| created_on | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_on | TIMESTAMP | NULLABLE | Last update timestamp |

**Indexes:**
- `identifier` - Unique index (critical for avoiding name conflicts)
- `idx_deployment_user_status` - Composite index on (user_id, status)

**Foreign Keys:**
- `user_id` → `users.id`

## Key Design Decisions

### 1. Identifier (Unique) in Deployment Table

**Decision:** The `identifier` field in the `deployments` table is unique, and the deployment name/app name is derived from this identifier.

**Rationale:** This avoids edge cases wherein:
- A deployment with a name is created but not propagated to the database
- A deployment is deleted in Kubernetes but that information is not propagated, preventing users from creating a new deployment with the same name

**Benefits:**
- Prevents naming conflicts even if synchronization fails
- Allows users to retry deployments with the same identifier
- Ensures consistent naming across the system

**Example Scenario:**
1. User creates deployment with identifier `my-app-v1`
2. Deployment is created in Kubernetes but watcher fails before syncing to DB
3. User manually deletes deployment from Kubernetes
4. User can still create a new deployment with identifier `my-app-v1` because the identifier is unique and not tied to Kubernetes name conflicts

### 2. Request ID (Unique) for Idempotency

**Decision:** The `request_id` field in the `deployment_requests` table is unique to ensure idempotency of requests.

**Rationale:** Any retry on the client side with the same `request_id` doesn't lead to creation of duplicate entries.

**Benefits:**
- Prevents duplicate deployment requests from client retries
- Enables safe retry mechanisms
- Simplifies error handling and recovery

**Example Scenario:**
1. Client sends request with `request_id: "req-123"`
2. Request is processed and stored
3. Client retries with same `request_id: "req-123"` due to network timeout
4. System recognizes duplicate and returns existing request instead of creating a new one

## Common Fields

All tables include common fields from the `Common` struct:

- `id` - UUID primary key
- `created_on` - Timestamp of creation
- `updated_on` - Timestamp of last update (nullable)

These fields are automatically managed by GORM hooks:
- `BeforeCreate` - Sets ID and `created_on` timestamp
- `BeforeUpdate` - Updates `updated_on` timestamp

## Relationships

```
users (1) ──< (many) deployment_requests
users (1) ──< (many) deployments
```

- One user can have many deployment requests
- One user can have many deployments
- Deployment requests and deployments are linked to users via `user_id` foreign key
