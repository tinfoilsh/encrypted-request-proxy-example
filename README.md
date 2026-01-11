# Encrypted Request Proxying with Tinfoil

A small example that demonstrates how to proxy inference requests through third-party servers while preserving end-to-end encryption using the [Encrypted HTTP Body Protocol](https://github.com/tinfoilsh/encrypted-http-body-protocol).

The protocol encrypts HTTP message bodies using Hybrid Public Key Encryption (HPKE) while preserving routing metadata, allowing proxies to inspect headers and route requests while keeping the actual payload encrypted end-to-end.

## Project Structure

```text
├── server/
│   └── main.go          # Go proxy that adds API key and forwards requests
├── examples/
│   ├── typescript/      # Browser-based TypeScript example
│   │   ├── main.ts
│   │   ├── index.html
│   │   └── styles.css
│   └── swift/           # macOS/iOS Swift example
│       ├── Package.swift
│       └── Sources/main.swift
```

> **Note:** The proxy can be implemented in any language (Go, Rust, Python, etc.) with no special dependencies - it only requires basic HTTP and header parsing.

In this example, the Go proxy intercepts `/v1/chat/completions` requests to:
- Read the `X-Tinfoil-Enclave-Url` header to determine which enclave the client verified
- Inspect and preserve EHBP-specific headers (`Ehbp-Encapsulated-Key` for requests, `Ehbp-Response-Nonce` for responses)
- Add your `TINFOIL_API_KEY` as the Authorization header
- Forward the encrypted request body to the verified Tinfoil enclave

The proxy can see routing metadata but cannot decrypt the request/response bodies, which remain encrypted end-to-end between the client and the Tinfoil enclave.

## Prerequisites

- Go 1.21+ (for the proxy server)
- Node.js 18+ (for the TypeScript example)
- Swift 5.9+ / Xcode 15+ (for the Swift example)
- `TINFOIL_API_KEY` exported in your shell

## Running the TypeScript Example

```bash
# Terminal 1 – start the proxy on http://localhost:8080
export TINFOIL_API_KEY=tk-...
cd server && go run main.go

# Terminal 2 – serve the static files with Vite
cd examples/typescript
npm install
npx vite
```

Open the printed Vite URL (typically http://localhost:5173), type a message, and
watch the assistant stream its reply.

### Tweaks

- Change the `baseURL` or model in `main.ts` if you want to point at a different
  proxy server or model. Defaults are `http://localhost:8080` and `gpt-oss-120b`.
- The SDK automatically fetches enclave configuration from the router and sends
  the enclave URL to the proxy via the `X-Tinfoil-Enclave-Url` header.

## Running the Swift Example

```bash
# Terminal 1 – start the proxy on http://localhost:8080
export TINFOIL_API_KEY=tk-...
cd server && go run main.go

# Terminal 2 – run the Swift example
cd examples/swift
swift run
```

The Swift SDK will:

1. Fetch available routers from Tinfoil
2. Perform remote attestation to verify the enclave
3. Set up EHBP encryption using the verified public key
4. Send requests to the proxy at `http://localhost:8080` with the `X-Tinfoil-Enclave-Url` header

## Advanced Usage

### Unverified Client (TypeScript)

**Caution:** This should only be used if you have alternate ways of performing client verification. This bypasses verification checks.

In `main.ts`, replace `SecureClient` with `UnverifiedClient`. Run as usual.
