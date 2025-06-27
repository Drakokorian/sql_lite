Phase 7 — Release & Supply-Chain Hardening (Pure Go)
Goal: ship v1.0.0 binaries whose provenance, reproducibility, and security can be verified without introducing a single third-party package.
Everything below is implemented with the stock Go tool-chain, ­crypto/*, os/exec, hash, encoding, and built-in POSIX/Windows sys-calls.
CI jobs may run auxiliary commands that are checked into the repo as Go source; no external binaries are fetched at build time.

7.1 · Build & Release Pipeline — cmd/release/main.go
Capability	Pure-Go realisation	Validation
Reproducible build	go run cmd/release	

Creates a throw-away $TMPDIR/buildenv GOPATH.

Runs go clean -cache -modcache.

Invokes GOOS/GOARCH matrix using go build -trimpath -buildvcs=false -ldflags "-s -w -X main.commit=$GIT_SHA".

Captures entire go version and go env -json into buildinfo.json. | CI builds same commit twice → identical sha256sum. |
| Cross-compilation | Loops over hard-coded target list:
linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64. Uses only the stock linker. | All six binaries spawn and print Hello gosqlite smoke test. |
| SBOM generation | In-tree Go code runs go list -m -json all, marshals to SPDX-lite JSON (ID, version, checksum, licence from go list -m -json). | jq diff on two runs of same commit → identical SBOM. |
| Vulnerability scan | Pipeline executes go vuln ./... (part of Go tool-chain since 1.22) and fails if any GO- advisory is in critical, high. | CI log shows zero findings on green build. |
| Static analysis | go vet ./... and go test -run=^$ -race ./... executed inside pipeline; no external linters. | CI fails on first non-zero output. |
| Artifact signing | crypto/ed25519 keypair generated once and stored as base64 in repo’s release-keys/maintainer.ed25519.pub.
Pipeline signs each sha256sum file; signature emitted as *.sig. Verification script in Go checks both hash and signature. | go run verify.go gosqlite_linux_amd64.tar.gz exits 0. |
| Tamper detection | Every tar/zip contains: checksums.txt, SIGNATURE, buildinfo.json, SBOM.json. The verify.go script (std-lib only) re-hashes files, validates signature, and prints “OK”. | Manual unpack + verify passes. |
| Artifact storage | Release tool pushes to S3/Artifactory only via HTTPS PUT with presigned URL read from env; no SDK import. | Dry-run mode prints curl command for auditors. |

7.2 · API Stability & Versioning
Semantic Versioning:

v1.0.0 locks every exported name in gosqlite/.

v1.x.y never changes behaviour in a way that breaks existing callers as verified by an in-repo compatibility test (go vet -test).

Compatibility harness: cmd/compatcheck builds a small user program against previous tag and current HEAD; must compile and run.

7.3 · Release Documentation Generator
cmd/mkrelease (all std-lib):

Reads Git log since last tag → RELEASE_NOTES.md sections Features, Fixes, Breaking.

Inserts SBOM digest, buildinfo hash, and vuln report summary.

Emits skeleton security advisory if go vuln flagged anything between last and new tag.

7.4 · Quality Gates for Release Pipeline
Gate	Tool (std-lib / go cmd)	Pass condition
Rebuild reproducibility	go run cmd/release twice	identical SHA-256 for every artifact
SBOM consistency	internal Go diff of two SBOM JSONs	zero line differences
Signature validity	go run verify.go artifact	prints OK
Vuln scan	go vuln ./...	no critical/high advisories
API compat	go run cmd/compatcheck	exits 0

Documentation shipped with each tag
RELEASE_NOTES.md – automatic changelog.

INSTALL.md – one-liner go get plus checksum verification steps (all std-lib commands).

SECURITY.md – responsible disclosure, key fingerprints, signature-verification guide.

✅ Phase 7 Exit Checklist
cmd/release produces identical artifacts on two clean machines.

verify.go passes for every platform archive.

Vuln scan, race tests, vet all green.

SBOM present, signature valid, SHA-256 printed on binary --version output.

Tag v1.0.0 pushed; binaries uploaded; release notes published.