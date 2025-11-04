# Research & Technical Decisions: Kainuguru MVP

**Date**: 2025-11-04
**Feature**: Kainuguru Grocery Flyer Aggregation System

## Executive Summary

Research conducted to resolve technical decisions for building a Go monolith with PostgreSQL for Lithuanian grocery flyer aggregation. All decisions align with the constitution's "Simplicity First" and "PostgreSQL Everything" principles while ensuring Lithuanian language support and cost-effective operations.

## Technical Decisions

### 1. Database Access Layer

**Decision**: Bun ORM
**Rationale**:
- Near-raw-SQL performance (within 1.5x of sqlx) while providing ORM conveniences
- Excellent bulk operations support critical for inserting thousands of products weekly
- Thin abstraction over database/sql maintains control and visibility
- Modern design created specifically for performance

**Alternatives Considered**:
- GORM: Rejected due to significant reflection overhead that degrades performance with large datasets
- sqlx: Would work but requires more boilerplate for common operations
- Raw database/sql: Too low-level for rapid MVP development

### 2. PostgreSQL Full-Text Search Configuration

**Decision**: Built-in PostgreSQL Lithuanian Configuration with trigram indexes
**Rationale**:
- PostgreSQL has native Lithuanian language support with proper stemming
- Trigram indexes (pg_trgm) provide fuzzy matching for spelling variations
- Single database solution aligns with "PostgreSQL Everything" principle
- No additional infrastructure required

**Implementation**:
```sql
CREATE INDEX idx_products_fts ON products
USING gin(to_tsvector('lithuanian', product_name || ' ' || description));

CREATE INDEX idx_products_trigram ON products
USING gin(product_name gin_trgm_ops);
```

**Alternatives Considered**:
- Elasticsearch: Rejected - adds complexity, violates "PostgreSQL Everything"
- PGroonga extension: Optional enhancement for post-MVP if needed

### 3. Job Queue Implementation

**Decision**: PostgreSQL-based queue with SELECT FOR UPDATE SKIP LOCKED
**Rationale**:
- Leverages existing PostgreSQL infrastructure
- SKIP LOCKED prevents worker contention without external coordination
- Provides exactly-once processing guarantees
- Proven pattern used successfully in production systems

**Implementation Pattern**:
```sql
SELECT * FROM jobs
WHERE status = 'pending'
FOR UPDATE SKIP LOCKED
LIMIT 1;
```

**Alternatives Considered**:
- NATS/RabbitMQ: Rejected - additional infrastructure complexity
- Redis queues: Would work but PostgreSQL solution is simpler

### 4. Product Table Partitioning Strategy

**Decision**: Range partitioning by week aligned with flyer cycles
**Rationale**:
- Weekly partitions match flyer update frequency
- Enables efficient querying of current products via partition pruning
- Easy archival of old data (drop old partitions)
- Supports price history requirement from clarifications

**Alternatives Considered**:
- No partitioning: Would work initially but degrades over time
- Monthly partitions: Too coarse for weekly flyer updates
- Daily partitions: Too fine-grained, excessive partition overhead

### 5. Connection Pooling Library

**Decision**: pgx with pgxpool
**Rationale**:
- Native PostgreSQL driver with best performance
- Built-in connection pooling optimized for PostgreSQL
- Binary protocol reduces parsing overhead
- Supports all PostgreSQL-specific features we need

**Configuration**:
- Max connections: 25 (as specified in requirements)
- Min connections: 5
- Connection lifetime: 1 hour

**Alternatives Considered**:
- database/sql with lib/pq: Slower, text protocol
- GORM's built-in pooling: Tied to ORM choice

### 6. Bulk Insert Strategy

**Decision**: Multi-valued INSERT with ON CONFLICT and UNNEST for batches
**Rationale**:
- UNNEST provides optimal performance for bulk operations
- ON CONFLICT handles upserts efficiently
- Batch size of 1000 products balances memory and performance
- Can insert 10,000+ products in seconds

**Alternatives Considered**:
- Individual INSERTs: Too slow for thousands of products
- COPY command: Faster but doesn't handle upserts well

### 7. ChatGPT Vision API Integration

**Decision**: GPT-4o with structured JSON output and Lithuanian-optimized prompts
**Rationale**:
- Best accuracy for Lithuanian text extraction
- Structured output ensures consistent data format
- Cost acceptable for MVP without optimization

**Cost Optimization Strategy**:
- Image compression to 1024x1024 before API calls
- Redis caching of extraction results (24-hour TTL)
- Batch processing during off-peak hours
- TODO: Implement cost monitoring for post-MVP

**Alternatives Considered**:
- GPT-4o-mini: Cheaper but lower accuracy for Lithuanian
- Traditional OCR (Tesseract): Poor Lithuanian support, requires training

### 8. Error Handling Strategy

**Decision**: Graceful degradation with circuit breakers
**Rationale**:
- Failed extractions display page as-is (per clarification)
- Circuit breakers prevent cascade failures
- Exponential backoff with jitter for retries
- Aligns with "Fail Gracefully" principle

**Implementation**:
- Max 3 retries with exponential backoff
- Circuit breaker opens after 5 consecutive failures
- Manual review queue for persistent failures

### 9. Rate Limiting Approach

**Decision**: Token bucket algorithm with per-user limits
**Rationale**:
- Smooth request distribution prevents API throttling
- Per-user limits ensure fair usage
- Global daily quota prevents budget overruns
- Simple to implement and understand

**Configuration**:
- 100 requests/minute global (as specified)
- 10 requests/minute per user
- 50,000 tokens/day budget limit

### 10. Lithuanian Text Normalization

**Decision**: Dual storage - original and normalized text
**Rationale**:
- Preserves original for display
- Normalized version for search (ą→a, č→c, etc.)
- Supports both exact and fuzzy matching
- Handles mixed Lithuanian/English products

**Implementation**:
```go
// Store both versions
product.Name = "Žemaitijos pienas"
product.NormalizedName = "zemaitijos pienas"
```

## Infrastructure Decisions

### Docker Configuration

**Decision**: Multi-stage builds with Alpine Linux
**Rationale**:
- Minimal image size (~20MB for Go binary)
- Security benefits from minimal attack surface
- Fast builds with layer caching

### Monitoring Approach

**Decision**: Structured logging with zerolog, no metrics initially
**Rationale**:
- Aligns with "Basic Logging" principle
- Zerolog provides fast, structured JSON logs
- Can query logs for metrics if needed
- TODO: Add Prometheus/Grafana post-MVP

### Deployment Strategy

**Decision**: DigitalOcean Droplet + Managed PostgreSQL
**Rationale**:
- €80/month budget fits 4GB droplet + managed DB
- Managed PostgreSQL reduces operational burden
- Simple deployment with docker-compose
- Can scale vertically initially

## Development Timeline Validation

### Week 1-2: Foundation ✅
- Feasible with Bun ORM and gqlgen code generation
- Database schema straightforward with partitioning

### Week 3: Extraction ✅
- ChatGPT integration well-documented
- PDF processing with pdftoppm is standard

### Week 4: Features ✅
- PostgreSQL FTS configuration is simple
- Shopping list logic clear from clarifications

### Week 5: Production ✅
- Standard Go patterns for pooling and rate limiting
- BDD tests with Ginkgo established practice

### Week 6: Deployment ✅
- Docker and DigitalOcean deployment straightforward
- Documentation can be generated from code

## Risk Mitigation

### Technical Risks

1. **ChatGPT API Costs**
   - Mitigation: Caching, compression, budget limits
   - Monitoring: Daily cost tracking
   - Fallback: Display unprocessed flyers

2. **Lithuanian Text Extraction Accuracy**
   - Mitigation: Optimized prompts, confidence scoring
   - Monitoring: Track extraction success rates
   - Fallback: Manual review queue

3. **PostgreSQL Performance**
   - Mitigation: Partitioning, indexes, connection pooling
   - Monitoring: Query performance logs
   - Fallback: Read replicas if needed

### Operational Risks

1. **Scraper Blocking**
   - Mitigation: Rate limiting, user-agent rotation
   - Monitoring: Success rate tracking
   - Fallback: Manual flyer upload

2. **Data Volume Growth**
   - Mitigation: Weekly partitions, image archival
   - Monitoring: Disk usage alerts
   - Fallback: Older partition cleanup

## Post-MVP Enhancements

Priority order for future improvements:

1. **Cost Optimization**
   - Implement ChatGPT cost monitoring dashboard
   - Test GPT-4o-mini for non-critical pages
   - Batch processing optimizations

2. **Search Enhancement**
   - Add PGroonga for better Lithuanian support
   - Implement search suggestions
   - Category-based filtering

3. **Monitoring**
   - Add Prometheus metrics
   - Grafana dashboards
   - Alerting for critical failures

4. **Performance**
   - Redis caching for popular searches
   - CDN for flyer images
   - Database read replicas

5. **Features**
   - Email notifications for price drops
   - Mobile app API support
   - Store location-based filtering

## Conclusion

All technical decisions have been validated against the constitution and proven to be implementable within the 6-week timeline. The chosen stack (Go + PostgreSQL + ChatGPT) provides a solid foundation for the MVP while maintaining simplicity and allowing for future optimizations.