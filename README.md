# RSS gRPC Service

This service provides a robust interface for working with RSS feeds through gRPC. It can fetch feed contents, validate feed URLs, and process multiple feeds concurrently. The service handles various RSS feed formats and includes proper error handling for invalid or inaccessible feeds.

**Prerequisites:** the service requires latest stable version of Go and protobuf compiler to be installed on your system.

**Usage:** the service exposes three main RPC endpoints:

- `GetRssFeed` accepts a single RSS feed URL and returns its contents, including titles, descriptions, and items.
- `GetRssFeeds` handles multiple RSS feed URLs simultaneously. It processes the feeds in parallel and returns all successfully parsed feeds. Failed feeds are not returned.
- `ValidateRssFeed` checks if a given URL points to a valid RSS feed.

**Development:** the project uses a Makefile to simplify common development tasks.

**Deployment:** the service is deployed on Fly.io. If needed, replace the default `fly.toml` with the following contents:

```toml
app = 'XXX'
primary_region = 'YYY'

[build]

[[services]]
  internal_port = 50051
  protocol = "tcp"

  [[services.ports]]
    handlers = ["tls"]
    port = "443"

  [services.ports.tls_options]
    alpn = ["h2"]

[[vm]]
  size = 'shared-cpu-1x'
```