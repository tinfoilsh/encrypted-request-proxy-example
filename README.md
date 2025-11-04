# Encrypted Request Proxying with Tinfoil

A small example that demonstrates how to proxy inference requests through third-party servers while preserving end-to-end encryption using the [Encrypted HTTP Body Protocol](https://github.com/tinfoilsh/encrypted-http-body-protocol).

The protocol encrypts HTTP message bodies using Hybrid Public Key Encryption (HPKE) while preserving routing metadata, allowing proxies to inspect headers and route requests while keeping the actual payload encrypted end-to-end.

This example contains two pieces:

- `main.go`: a Go proxy that adds your TINFOIL_API_KEY and forwards chat completions to
the Tinfoil enclaves for inference.
- `main.ts`: a few lines of TypeScript that instantiates the Tinfoil `SecureClient`, sends
  `/v1/chat/completions`, and streams the response into the page.

> **Note:** The proxy can be implemented in any language (Go, Rust, Python, etc.) with no special dependencies - it only requires basic HTTP and header parsing.

In this example, the Go proxy intercepts `/v1/chat/completions` requests to:
- Inspect and preserve EHBP-specific headers (`Ehbp-Client-Public-Key`, `Ehbp-Encapsulated-Key`, `Ehbp-Fallback`)
- Add your `TINFOIL_API_KEY` as the Authorization header
- Forward the encrypted request body to the Tinfoil enclave at `https://ehbp.inf6.tinfoil.sh/v1/chat/completions`

The proxy can see routing metadata but cannot decrypt the request/response bodies, which remain encrypted end-to-end between the browser and the Tinfoil enclave.

## Prerequisites

- `npm install` already run in this directory
- `TINFOIL_API_KEY` exported in your shell

## Running the demo

```bash
# Terminal 1 – start the proxy on http://localhost:8080
export TINFOIL_API_KEY=tk-...
go run main.go

# Terminal 2 – serve the static files with Vite
npx vite
```

Open the printed Vite URL (typically http://localhost:5173), type a message, and
watch the assistant stream its reply. The demo intentionally keeps the UI and
error handling minimal so it is easy to read and adapt.

### Tweaks

- Change the `baseURL` or model in `main.ts` if you want to point at a different
  proxy server or model. Defaults are `http://localhost:8080` and `gpt-oss-120b`.

### Advanced usage with unverified client

Caution: This should only be used if you have alternate ways of performing client verification. This bypasses verification checks.

In `main.ts`, replace `SecureClient` with `UnverifiedClient`. Run as usual.
