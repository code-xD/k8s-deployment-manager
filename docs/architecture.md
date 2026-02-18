# Architecture

## High-Level Design (HLD)

The K8s Deployment Manager follows an event-driven architecture with clear separation of concerns. The system consists of the following components:

### Stack Components

#### Third-Party Components (External Dependencies)
- **DB (Postgres)** - PostgreSQL database for persistent storage
- **Queue (NATS)** - NATS JetStream for message queuing and event streaming
- **K8sClient** - Kubernetes client library for cluster interactions

#### Own Stack Components (Application Services)
- **API** - REST API server for handling user requests
- **Worker** - Background worker for processing deployment requests and updates
- **Watcher** - Kubernetes informer that watches for deployment changes

### Data Objects

#### 1. User
A simple user entity. For simplicity, any new user-id present in the request will automatically create a user record.

**Key Characteristics:**
- Auto-created on first request with a new user-id
- Identified by `user_external_id`
- Used for authentication and authorization

#### 2. Deployment Request
A request raised by a user which is to be fulfilled in an asynchronous manner. Multiple deployment requests can be raised for the same deployment.

**Key Characteristics:**
- Represents user intent (CREATE, UPDATE, DELETE)
- Tracks request status (CREATED, SUCCESS, FAILURE)
- Contains deployment configuration (image, metadata, namespace, etc.)
- Unique `request_id` for idempotency
- Linked to a user via `user_id`

#### 3. Deployment
The deployment object which is synced from Kubernetes with the help of watcher, NATS queue, and worker in an asynchronous manner. The deployment object's single source of truth is the Kubernetes cluster.

**Key Characteristics:**
- Represents actual Kubernetes deployment state
- Synced from Kubernetes cluster via watcher
- Unique `identifier` used as deployment name/app name
- Tracks deployment status (INITIATED, CREATED, UPDATING, DELETED)
- Contains Kubernetes resource version for conflict detection

### Data Flow

#### 1. Create Deployment Request Flow

```
User → API (POST /deployments/requests/create) → NATS Queue → Worker → Kubernetes API
                                                                    ↓
                                                              Store Result in DB
```

**Steps:**
1. New/existing user calls `POST /api/v1/deployments/requests/create` API to create a deployment request
2. API validates request, creates user if needed, and stores deployment request in database
3. Deployment request is published to NATS queue
4. Worker receives the request from the queue
5. Worker calls Kubernetes API to execute the deployment
6. Success or failure of the Kubernetes call is stored in the deployment request
7. Response is returned to the user

#### 2. Update/Delete Deployment Flow

```
User → API (PATCH/DELETE /deployments/requests/:id) → NATS Queue → Worker → Kubernetes API
                                                                         ↓
                                                                   Update Request Status
```

**Steps:**
1. User calls API to update or delete an existing deployment
2. API validates request and updates deployment request in database
3. Message is published to NATS queue
4. Worker consumes the message and calls Kubernetes API
5. Result is stored back in the deployment request

#### 3. Watcher Flow (Kubernetes → Database Sync)

```
Kubernetes Cluster → Watcher (Informer) → NATS Queue → Worker → Pull from K8s → Update DB
```

**Steps:**
1. Watcher runs as a Kubernetes informer that listens to deployment changes
2. Watcher filters deployments managed by `k8s-deployment-manager` (via label selector)
3. When a deployment change is detected, watcher pushes the deployment name to NATS queue
4. Worker receives the deployment name from the queue
5. Worker pulls the full deployment data from Kubernetes API
6. Worker updates the deployment object in the database with current Kubernetes state

**Note:** The watcher ensures that the database stays synchronized with the actual Kubernetes cluster state, making Kubernetes the single source of truth for deployment objects.
