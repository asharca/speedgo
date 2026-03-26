# SpeedGo

A command-line network speed test tool written in Go, supporting both Chinese domestic and international servers.

## Features

- **Multi-region support** — test against Chinese (cn) or global servers, or auto-detect the best ones
- **Auto server selection** — latency-based probing picks the fastest servers automatically
- **Concurrent testing** — multiple parallel streams for accurate bandwidth measurement
- **Real-time progress** — live speed display with visual progress bar
- **Full test suite** — one command to run ping + download + upload with summary
- **ICMP Ping** — raw ICMP ping with min/avg/max/loss statistics

## Installation

```bash
go install github.com/asharca/speedgo@latest
```

Or build from source:

```bash
git clone https://github.com/asharca/speedgo.git
cd speedgo
go build -o speedgo .
```

## Usage

### Full Speed Test

```bash
# Auto-detect best servers and run all tests
speedgo all

# Test against Chinese servers only
speedgo all --region=cn

# Test against global servers only
speedgo all --region=global
```

### Ping

```bash
# Ping default targets (auto region)
speedgo ping

# Ping Chinese servers
speedgo ping --region=cn

# Ping custom targets
speedgo ping --targets=google.com,baidu.com --count=5

# Verbose output
speedgo ping --region=global --verbose
```

### Download

```bash
# Download test with auto server selection
speedgo download

# Short alias, Chinese servers, 15s duration
speedgo d --region=cn --duration=15s

# Custom concurrency
speedgo d --region=global --concurrency=8 --duration=20s
```

### Upload

```bash
# Upload test with auto server selection
speedgo upload

# Short alias, 20s duration
speedgo u --region=cn --duration=20
```

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `ping` | `p` | Test network latency to multiple targets |
| `download` | `d` | Test download speed |
| `upload` | `u` | Test upload speed |
| `all` | `a` | Run full speed test (ping + download + upload) |

## Flags

### Common Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--region` | `auto` | Server region: `cn`, `global`, `auto` |
| `--concurrency` | `4` | Number of concurrent streams |
| `--verbose` | `false` | Enable detailed output |

### Ping Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--targets` | (per region) | Comma-separated targets |
| `--count` | `4` | Pings per target |
| `--timeout` | `1s` | Timeout per ping |

### Download Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--duration` | `30s` | Test duration |
| `--url` | (auto) | Custom download URL |

### Upload Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--duration` | `10` | Test duration in seconds |
| `--url` | (auto) | Custom upload endpoint |

## Servers

### Chinese Servers (--region=cn)

| Type | Provider |
|------|----------|
| Download | USTC Mirror, Tsinghua Mirror, Aliyun Mirror |
| Ping | baidu.com, aliyun.com, tencent.com, qq.com, bilibili.com |

### Global Servers (--region=global)

| Type | Provider |
|------|----------|
| Download | Cloudflare, OVH, Hetzner |
| Upload | Cloudflare |
| Ping | cloudflare.com, google.com, amazon.com |

## How It Works

1. **Server Probing** — sends HTTP HEAD requests to all candidate servers in parallel, measures latency
2. **Server Selection** — sorts by latency, picks the fastest N servers
3. **Speed Test** — runs concurrent download/upload streams against selected servers
4. **Progress Display** — shows real-time speed bar updated every 200ms
5. **Results** — calculates and displays average speed in Mbps

## Note

The `ping` command requires raw socket access. On most systems, you need to either:

- Run with `sudo`
- Set the capability: `sudo setcap cap_net_raw+ep ./speedgo`

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./core/ -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## License

See [LICENSE](LICENSE) file.
