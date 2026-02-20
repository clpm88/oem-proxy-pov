# Neo4j License Proxy (proxy-test.go)

## Purpose

A lightweight TCP proxy that validates license expiration before forwarding connections to a Neo4j database. Acts as a "front door" and "bouncer" for OEM deployments.

## How It Works

1. **Listens** on port `:7688` for incoming client connections
2. **Validates** the license expiration date
   - If expired: Closes the connection immediately
   - If valid: Forwards the connection to Neo4j
3. **Proxies** all traffic between client and Neo4j on port `:7687` using Layer 4 bidirectional streaming

The proxy doesn't parse Bolt protocol or Cypher queries—it simply streams bytes transparently.

## Configuration

**License Expiration**: Hardcoded in the source (line 12):
```go
var licenseExpiration = time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
```

**Ports**:
- Proxy listens on: `:7688`
- Neo4j backend: `localhost:7687`

## Usage

### Run the proxy

```bash
go run proxy-test.go
```

### Connect clients

Point your Neo4j client to port `7688` instead of the default `7687`:

```bash
# Example with neo4j-shell or driver
neo4j://localhost:7688
```

### Monitor logs

The proxy logs connection events:
- Startup confirmation
- License validation results
- Connection forwarding status
- Connection closures

## Production Considerations

This is a PoC. For production use, you would need to:
- Read license data from an OEM license file or service (not hardcoded)
- Add proper Bolt protocol error responses for expired licenses
- Implement connection pooling and resource limits
- Add metrics and health checks
- Configure ports via environment variables or config files
