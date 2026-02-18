# Design Decisions

This document outlines the key architectural and design decisions made in the K8s Deployment Manager system.

## Event-Driven Architecture

### Rationale
Apart from the requirement of the assignment, infrastructure provisioning is tedious in nature. In high-volume scenarios, provisioning of infrastructure can take between 5-10 minutes. In order to avoid any thread being hogged or manage complex states, it's better to communicate any action via queue and then handle it asynchronously.

### Benefits
- **Non-blocking Operations**: API endpoints return immediately after queuing the request, preventing thread blocking
- **Scalability**: Components can be scaled independently based on load
- **Resilience**: Failed requests can be retried via Dead Letter Queue (DLQ) or cron jobs
- **State Management**: Complex state transitions are handled asynchronously, reducing complexity

### Implementation
- All deployment operations (CREATE, UPDATE, DELETE) are queued via NATS JetStream
- Worker processes consume messages and execute Kubernetes operations
- Results are stored back in the database for tracking

## Single Source of Truth / Responsibility

### Principle
Whatever data object we have will always pertain to a single source of truth:
- **Deployment** → Kubernetes (single source of truth)
- **Deployment Request** → API/Database (single source of truth)

### Benefits

#### 1. Conflict Avoidance
This ensures that there is no conflict or complex state management. State transitions are ordered in nature and closer to real-time.

#### 2. Clear Responsibility Boundaries
This also ensures that responsibilities are constrained to a single component:
- **Deployment State Updates**: Any failure of deployment state update can be debugged at the consumer (worker) level
- **Request Fulfillment**: In case there's any issue in request fulfillment, we could debug the issue separately and rerun the request with the help of DLQ or cron jobs

#### 3. Data Consistency
- Kubernetes is the authoritative source for deployment state
- Database deployment records are synchronized from Kubernetes via the watcher
- Deployment requests track user intent and are managed by the API layer

## Request Model

### Rationale
Although the theme is repeated in the above points, to re-iterate, we follow a request model so that each component can be scaled separately and failure of any component doesn't propagate into other systems.

### Benefits
- **Independent Scaling**: API, Worker, and Watcher can be scaled independently based on their respective loads
- **Fault Isolation**: Failure of one component (e.g., worker) doesn't affect others (e.g., API)
- **Retry Capability**: Failed requests can be retried without affecting the API layer
- **Monitoring**: Each component can be monitored and debugged independently

### Implementation
- Deployment requests are stored in the database before being queued
- Each request has a unique `request_id` for idempotency
- Request status is tracked through the lifecycle (CREATED → SUCCESS/FAILURE)

## Caveats

### 1. Near-Realtime Updates
Any updates taking place in Kubernetes are not communicated in real-time. Although at low throughput it might not cause an issue, however in high throughput when the queue is at limit, the lag would be higher.

**Impact:**
- There may be a delay between when a deployment changes in Kubernetes and when it's reflected in the database
- Queue backpressure can increase this delay
- Users querying deployments may see slightly stale data

**Mitigation:**
- Monitor queue depth and scale workers accordingly
- Implement queue metrics and alerting
- Consider implementing direct Kubernetes queries for critical real-time operations

### 2. K8s-Informer Model Limitation
The informer API in k8sclient is pull-based in nature and always shows the current state instead of state transitions. Therefore, there are scenarios wherein the state transition might not be communicated.

**Example Edge Case:**
If the watcher is down for too long and a user creates a deployment, the deployment will be created. Now, in case the user decides to manually delete the deployment, the system might never come to know about the deployment object because when the watcher might come back up, the latest state would not contain that deployment.

**Impact:**
- State transitions that occur while the watcher is down may be missed
- Deployments created and deleted between watcher restarts may not be tracked
- The database may contain stale deployment records

**Mitigation:**
- Implement periodic reconciliation jobs to sync state
- Add health checks and alerting for watcher downtime
- Consider implementing event-based watchers if available in future Kubernetes versions
- Add manual reconciliation endpoints for administrative purposes
