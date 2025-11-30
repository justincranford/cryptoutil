# Queue Listeners Decision - Task 10.3

## Decision: Queue Listeners Not Required (Current Implementation)

**Date**: November 10, 2025
**Status**: Deferred for future consideration

## Context

Task 10 (Integration Layer Completion) originally mentioned implementing queue listeners for async operations. After reviewing the current identity service architecture, the need for message queues has been evaluated.

## Current Architecture

The identity service stack consists of:

- **AuthZ Server**: OAuth 2.1 Authorization Server (synchronous HTTP API)
- **IdP Server**: OIDC Identity Provider (synchronous HTTP API)
- **RS Server**: Resource Server with OAuth 2.0 Bearer token validation (synchronous HTTP API)
- **Background Jobs**: Periodic cleanup jobs for expired tokens/sessions (time-based triggers)
- **PostgreSQL Database**: Persistent storage for tokens, sessions, clients
- **OTEL Collector**: Telemetry aggregation (push-based telemetry)

## Analysis

### Current Async Operations

1. **Background Cleanup Jobs**: Implemented using Go time.Ticker for periodic execution
   - Token expiration cleanup
   - Session expiration cleanup
   - No external queue needed - simple time-based scheduling

2. **Telemetry Export**: Handled by OTEL Collector
   - Push-based from services to collector
   - No queuing needed at application layer

3. **HTTP Request/Response**: All OAuth/OIDC flows are synchronous
   - Authorization requests: immediate redirects
   - Token exchange: immediate responses
   - Token introspection: immediate responses
   - UserInfo requests: immediate responses

### When Message Queues Would Be Beneficial

Message queues (RabbitMQ, Kafka, NATS, etc.) become valuable when:

1. **Decoupling Services**: Multiple services need to react to events asynchronously
2. **Load Buffering**: Need to handle traffic spikes by queuing requests
3. **Guaranteed Delivery**: Critical operations must not be lost
4. **Event Sourcing**: Building audit trail or event history
5. **Long-Running Operations**: Background processing that takes minutes/hours

### Current Use Cases - No Queue Needed

| Use Case | Current Solution | Queue Alternative | Decision |
|----------|------------------|-------------------|----------|
| Token Cleanup | time.Ticker background job | Queue-based job scheduler | **Not needed** - ticker sufficient |
| Session Cleanup | time.Ticker background job | Queue-based job scheduler | **Not needed** - ticker sufficient |
| OAuth Authorization | Synchronous HTTP | Event-driven async | **Not needed** - must be synchronous per spec |
| Token Exchange | Synchronous HTTP | Event-driven async | **Not needed** - must be synchronous per spec |
| Telemetry | OTEL Collector (push) | Message queue | **Not needed** - OTEL handles buffering |

## Future Considerations

Message queues SHOULD be considered if/when we add:

1. **Email/SMS Notifications**
   - OTP delivery
   - Magic link delivery
   - Account alerts
   - **Recommendation**: Use queue to decouple from external providers

2. **Webhook Delivery**
   - Event notifications to external systems
   - Retry logic for failed deliveries
   - **Recommendation**: Use queue for reliability

3. **Audit Event Processing**
   - Complex audit analysis
   - Compliance report generation
   - **Recommendation**: Use queue for buffering and replay

4. **Multi-Step Workflows**
   - MFA enrollment flows with multiple async steps
   - Account recovery workflows
   - **Recommendation**: Use queue for state management

5. **Integration with External Systems**
   - CRM synchronization
   - Analytics pipelines
   - **Recommendation**: Use queue for decoupling

## Implementation Readiness

If message queues become needed, the codebase is structured to add them easily:

```go
// Future message queue integration point
type QueueListener struct {
    queue    MessageQueue
    handlers map[string]EventHandler
}

// Hook into existing background job infrastructure
func (q *QueueListener) Start(ctx context.Context) error {
    // Subscribe to queue topics
    // Process messages
    // Call event handlers
}
```

**Integration points**:

- Add to `jobs/` package alongside `cleanup.go`
- Wire into ServerManager lifecycle (Start/Stop methods)
- Add queue configuration to `config/config.go`
- Add Docker Compose services (RabbitMQ, Kafka, NATS)

## Recommendation

**DO NOT implement message queues now**:

- Current architecture is simple and sufficient
- No identified use cases requiring queues
- Premature optimization adds complexity without benefit
- Easy to add when actually needed

**REVISIT when**:

- Adding Task 12 (OTP/Magic Link Services) - email/SMS delivery
- Adding webhooks or external integrations
- Adding complex multi-step workflows
- Observing performance bottlenecks requiring buffering

## References

- Task 10 requirements: `docs/identityV2/task-10-integration-layer-completion.md`
- Background jobs implementation: `internal/identity/jobs/cleanup.go`
- Server lifecycle management: `internal/identity/server/server_manager.go`
- Docker orchestration: `deployments/compose/identity-compose.yml`

## Sign-off

**Decision**: Defer queue implementation until specific use case identified
**Rationale**: Current synchronous architecture is appropriate for OAuth/OIDC specifications
**Next Review**: During Task 12 (OTP/Magic Link Services) implementation
