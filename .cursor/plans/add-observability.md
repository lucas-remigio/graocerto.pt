# Observability Implementation Plan (Full Local Stack)

## Objective
Add robust, low-resource observability (Structured Logs + Metrics + Visualization) to the Wallet Tracker application while staying strictly within a ~150MB total RAM budget. We will migrate the Go backend to use structured `slog` logging and expose Prometheus metrics, and then deploy lightweight Prometheus and Grafana containers to scrape and visualize this data.

## Key Files & Context
- `backend/go.mod`: Add the `prometheus/client_golang` dependency.
- `backend/utils/logger.go`: (New) Configure global `slog` instances with JSON formatting.
- `backend/cmd/api/api.go`: Register new middlewares and the `/metrics` endpoint.
- `backend/cmd/api/middlewares/observability.go`: (New) Add `RequestID`, `StructuredLogger`, and `PrometheusMetrics` middlewares.
- `ops/prometheus/prometheus.yml`: (New) Configuration to scrape the backend.
- `docker-compose.yaml` / `docker-compose.prod.yaml`: Add `prometheus` and `grafana` services with strict memory limits.

## Implementation Steps

### Phase 1: Structured Logging & Tracing (`slog`)
1. **Initialize `slog`**: Create a setup function that configures the global logger to use `slog.JSONHandler` (especially in production).
2. **Correlation ID Middleware**: Create a middleware that generates a unique `X-Request-ID` for every incoming HTTP request and injects it into the request context.
3. **Access Log Middleware**: Create a middleware that logs every HTTP request (Method, Path, Status, Latency, Request ID) as a single JSON object.
4. **Refactor Existing Logs**: Replace basic `log.Printf` and `log.Fatal` calls in the application startup (`main.go`) and critical paths with `slog`.

### Phase 2: Lightweight Metrics
1. **Add Dependency**: Run `go get github.com/prometheus/client_golang/prometheus` in the backend.
2. **Expose Endpoint**: Register the `promhttp.Handler()` at the `/metrics` path in the Go router.
3. **Metrics Middleware**: Implement a middleware that increments a `http_requests_total` counter and observes `http_request_duration_seconds` for all API endpoints, labeled by HTTP method, path, and status code.

### Phase 3: Infrastructure (Prometheus & Grafana)
1. **Prometheus Config**: Create `ops/prometheus/prometheus.yml` to scrape the Go backend (`app:${BACKEND_PORT}`) every 15 seconds.
2. **Docker Compose Updates**: Add `prometheus` (image: `prom/prometheus`) and `grafana` (image: `grafana/grafana-oss`) services to the Docker Compose files.
3. **Strict Resource Constraints**: To respect the VPS headroom constraints, we will add strict `deploy.resources.limits.memory` constraints in the compose file:
   - Prometheus: `64M` max RAM.
   - Grafana: `80M` max RAM.
   (This ensures the total observability footprint stays < 150MB).

## Verification & Testing
1. **Logs**: Ensure terminal output (or `docker logs`) shows neatly formatted JSON logs containing `request_id` and `latency`.
2. **Metrics Endpoint**: Hit `http://localhost:8080/metrics` and verify standard Go and custom HTTP metrics are visible.
3. **Infrastructure**: Start the full stack using `docker-compose up`. Ensure Prometheus (usually port 9090) shows the `app` target as "UP" and Grafana (port 3001) is accessible and running within its memory limits.
