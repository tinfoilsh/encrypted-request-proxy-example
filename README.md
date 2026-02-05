# Encrypted Request Proxying with Tinfoil

[![Docs](https://img.shields.io/badge/docs-tinfoil.sh-blue)](https://docs.tinfoil.sh/guides/proxy-server)

This example shows how to proxy inference requests through your own servers while preserving end-to-end encryption using the [Encrypted HTTP Body Protocol (EHBP)](https://github.com/tinfoilsh/encrypted-http-body-protocol).

EHBP encrypts HTTP message bodies using HPKE while keeping headers in plaintext. This lets your proxy inspect headers and route requests without being able to read the actual content.

## Project Structure

```
├── server/
│   └── main.go              # Go proxy server
└── clients/
    ├── typescript/          # Browser example
    └── swift/               # macOS/iOS example
```

## Quick Start

```bash
# Terminal 1: Start the proxy
export TINFOIL_API_KEY=tk_...
cd server && go run main.go

# Terminal 2: Start the TypeScript client
cd clients/typescript && npm install && npx vite
```

Open http://localhost:5173 and send a message.

## Prerequisites

- Go 1.21+
- Node.js 18+ (for TypeScript example)
- Swift 5.9+ (for Swift example)
- A Tinfoil API key

## What the Proxy Does

1. **Serves `/attestation`** — Proxies attestation bundles from Tinfoil so clients can verify enclaves
2. **Forwards `/v1/chat/completions` and `/v1/responses`** — Routes encrypted requests to the enclave specified in the `X-Tinfoil-Enclave-Url` header
3. **Adds authentication** — Injects your `TINFOIL_API_KEY` as the Bearer token
4. **Preserves encryption headers** — Copies `Ehbp-Encapsulated-Key` (request) and `Ehbp-Response-Nonce` (response)

The proxy sees routing metadata but **cannot decrypt** the request or response bodies.

---

## Implementing Your Own Proxy

### Endpoints

| Path | Method | Description |
|------|--------|-------------|
| `/attestation` | GET | Proxy to `https://atc.tinfoil.sh/attestation` |
| `/v1/chat/completions` | POST | Forward to the URL in `X-Tinfoil-Enclave-Url` header |
| `/v1/responses` | POST | Forward to the URL in `X-Tinfoil-Enclave-Url` header |

### Required Headers

| Direction | Header | Purpose |
|-----------|--------|---------|
| Request | `X-Tinfoil-Enclave-Url` | The enclave URL the client verified — use as upstream |
| Request | `Ehbp-Encapsulated-Key` | HPKE key for the enclave to decrypt the request |
| Response | `Ehbp-Response-Nonce` | Nonce for the client to decrypt the response |

### CORS (Browser Clients)

```
Access-Control-Allow-Headers: Ehbp-Encapsulated-Key, X-Tinfoil-Enclave-Url
Access-Control-Expose-Headers: Ehbp-Response-Nonce
```

---

## Usage Metrics for Billing

Since EHBP encrypts request/response bodies, your proxy cannot see token counts in the JSON. Tinfoil provides usage metrics via HTTP headers so you can bill your users.

### How It Works

1. **Request usage** — Add this header when forwarding to Tinfoil:
   ```
   X-Tinfoil-Request-Usage-Metrics: true
   ```

2. **Read the response** — Tinfoil returns usage in the `X-Tinfoil-Usage-Metrics` header:
   ```
   X-Tinfoil-Usage-Metrics: prompt=67,completion=42,total=109
   ```

3. **Streaming responses** — For streaming (`text/event-stream`), usage arrives as an HTTP trailer after the body completes. Read it from `resp.Trailer` after consuming the response body.

### Example (Go)

```go
// When building the upstream request:
req.Header.Set("X-Tinfoil-Request-Usage-Metrics", "true")

// After receiving the response:
// Non-streaming: read from response header
if usage := resp.Header.Get("X-Tinfoil-Usage-Metrics"); usage != "" {
    log.Printf("Usage: %s", usage)  // "prompt=67,completion=42,total=109"
}

// Streaming: read from trailer after body is consumed
io.Copy(w, resp.Body)
if usage := resp.Trailer.Get("X-Tinfoil-Usage-Metrics"); usage != "" {
    log.Printf("Usage: %s", usage)
}
```

### Format

```
prompt=<prompt_tokens>,completion=<completion_tokens>,total=<total_tokens>
```

Parse with a simple split on `,` and `=`.

---

## Request Flow

```
Client                    Proxy                     Tinfoil Enclave
  │                         │                              │
  │ GET /attestation        │                              │
  │────────────────────────>│ GET /attestation             │
  │                         │─────────────────────────────>│
  │                         │<─────────────────────────────│
  │<────────────────────────│ attestation bundle           │
  │                         │                              │
  │ (verify attestation)    │                              │
  │                         │                              │
  │ POST /v1/chat/completions                              │
  │ X-Tinfoil-Enclave-Url: https://...                     │
  │ Ehbp-Encapsulated-Key: <key>                           │
  │ Body: <encrypted>       │                              │
  │────────────────────────>│                              │
  │                         │ POST /v1/chat/completions    │
  │                         │ Authorization: Bearer <key>  │
  │                         │ Ehbp-Encapsulated-Key: <key> │
  │                         │ X-Tinfoil-Request-Usage-Metrics: true
  │                         │ Body: <encrypted>            │
  │                         │─────────────────────────────>│
  │                         │<─────────────────────────────│
  │                         │ Ehbp-Response-Nonce: <nonce> │
  │                         │ X-Tinfoil-Usage-Metrics: ... │
  │                         │ Body: <encrypted>            │
  │<────────────────────────│                              │
  │ Ehbp-Response-Nonce     │                              │
  │ Body: <encrypted>       │                              │
  │                         │                              │
  │ (decrypt response)      │                              │
```

The proxy adds authentication and reads usage metrics, but cannot decrypt the bodies.
