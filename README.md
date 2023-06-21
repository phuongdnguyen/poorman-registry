# Reverse registry

[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)

This is a Chainguard reverse registry that redirect to Chainguard's `cgr.dev/chainguard/*` public registry. The public, free tier of Chainguard Images only serves `latest` tag. This could be of inconvenience so we wrote this reverse registry to continously watching Chainguard registry for digest changes and extract the package version via SBOM. We then tag the image according with the packaged software version and serve via this reverse registry.

By default, in-mem sqlite is used but MySQL is recommended for production setup.
## Usage

```bash
go run main.go server --config=config.yaml
```
## How it works

```mermaid
sequenceDiagram
autonumber
actor U as User
participant RR as Reverse Registry
participant DB as Local Digest Database
participant CG as Chainguard Images
U->>+RR: Pull command `nginx:1.0.0`
RR->>+DB: Check if digest existed for `nginx:1.0.0`
DB-->>-RR: Found digest for `nginx:1.0.0`
RR-->>-U: Return digest if found

loop Every x Minutes
RR->>CG: Periodically checking `latest` tag for digest change
RR->>DB: Save digest for this tag to local db
end


RR->>CG: Proxied every other APIs

```
