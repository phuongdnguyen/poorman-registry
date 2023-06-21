# Reverse registry

This is Chainguard reverse registry that redirect to Chainguard's `cgr.dev/chainguard/*` public registry. The public, free tier of Chainguard Images only serves `latest` tag. This could be of inconvenience so we wrote this reverse registry to continously watching Chainguard registry for digest changes and extract the package version via SBOM. We then tag the image according with the packaged software version and serve via this reverse registry.

By default, in-mem sqlite is used but MySQL is recommended for production setup.
## Usage

To be update

How to run this locally/production.
## How it works

Just an example. Need to update this accordingly.

```mermaid
sequenceDiagram
actor U as User
participant RS as Reverse Registry
participant CG as Chainguard Images
U->>RS: Pull command `nginx:1.0.0`
RS->>U: Check and return digest if found
RS->>CG: Periodically checking `latest` tag for digest change
RS->>CG: Proxied every other APIs
```

## Deploy with Cloud Run

[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)
