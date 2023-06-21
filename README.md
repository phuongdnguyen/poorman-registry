# Reverse registry


This is a simple registry redirect service that redirect our.domain.com/* to cgr.dev/chainguard/*. It also run a background process to periodically hash image index to sha256 checksum and save to a local database (default in-mem sqlite is used. For production purpose, mysql is the recommended choice)

Deploying

[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)
