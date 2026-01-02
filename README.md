# external-secrets-passwork

A lightweight HTTP proxy that forwards requests to a Passwork instance and returns the password entry as JSON.
The binary authenticates once on start‑up, then serves two simple endpoints.

It is meant to be used with external-secrets generic webhook to fetch passwords.

The proxy reads its configuration from environment variables, making it easy to run in any CI/CD pipeline or container orchestration platform.

---

## Prerequisites

- **Docker**
- **Passwork** instance reachable from the container
- **Passwork API key** with read permissions for the vault(s) you want to query

---

## Configuration

All values are taken from environment variables at container start‑up:

| Variable          | Required? | Description                                                                                              |
|-------------------|-----------|----------------------------------------------------------------------------------------------------------|
| `PASSWORK_HOST`   | yes        | Base URL of the Passwork API (e.g. `https://my-passwork.example.com/api/v4`).                           |
| `PASSWORK_API_KEY`| yes       | API key for authenticating against Passwork.                                                            |

If any required variable is missing, the container will fail to start.

---

## Build the Docker image

The repository contains a multi‑stage `Dockerfile` that builds a static binary and copies it into a minimal Alpine image.

```bash
# From the repository root
docker build -t passwork-proxy .
```

---

## Run the container

```bash
docker run -d \
  --name external-secrets-passwork \
  -p 8080:8080 \
  -e PASSWORK_HOST="https://my-passwork.example.com/api/v4" \
  -e PASSWORK_API_KEY="YOUR_API_KEY" \
  external-secrets-passwork
```

---

## API usage examples

### Fetch a secret

```bash
curl -s http://localhost:8080/passwords/<secret-id>
```

Successful response (example):

```json
{
  "VaultId": "xxx",
  "FolderId": "",
  "Custom": [
    {
      "name": "xxx",
      "value": "xxx",
      "type": "xxx"
    },
    {
      "name": "xxx",
      "value": "xxx",
      "type": "xxx"
    }
  ],
  "Id": "xxx",
  "Name": "xxx",
  "Login": "",
  "CryptedPassword": "",
  "CryptedKey": "xxx",
  "Description": "",
  "Url": "",
  "Color": 0,
  "Attachments": [],
  "Tags": null,
  "Path": [
    {
      "Order": 0,
      "Name": "xxx",
      "Type": "xxx",
      "Id": "xxx"
    }
  ],
  "Access": "xxx",
  "AccessCode": 4,
  "Shortcut": {
    "Id": "",
    "PasswordId": "",
    "VaultId": "",
    "FolderId": "",
    "Access": "",
    "AccessCode": 0,
    "CryptedKey": ""
  },
  "LastPasswordUpdate": xxx,
  "UpdatedAt": "xxx",
  "IsFavorite": false
}
```

If the secret cannot be found:

```json
{
  "error": "secret not found"
}
```

### Health‑check

```bash
curl -s http://localhost:8080/healthz
# => {"status":"ok"}
```

---

## Logging

All logs are written to **STDOUT**:

---

## Development workflow

1. **Clone the repo**

   ```bash
   git clone https://github.com/lupa95/external-secrets-passwork.git
   cd external-secrets-passwork
   ```

2. **Update dependencies** (optional)

   ```bash
   go get -u github.com/lupa95/passwork-client-go
   go mod tidy
   ```

3. **Build the Docker image**

    ```bash
    docker build -t external-secrets-passwork:dev ./
    ```

4. **Run container**

   ```bash
   docker run -d \
    --name external-secrets-passwork \
    -p 8080:8080 \
    -e PASSWORK_HOST="https://my-passwork.example.com/api/v4" \
    -e PASSWORK_API_KEY="YOUR_API_KEY" \
    test:test
   ```
