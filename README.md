# Chirpy

Chirpy is a small social posting API written in Go. Users can create accounts, sign in, and publish short posts called chirps. PostgreSQL stores users, posts, and refresh tokens.

Built by following Boot.dev’s **Learn HTTP Servers in Go** course, this project puts HTTP routing, JSON APIs, SQL, authentication, and webhooks together in a working backend. It is a useful reference for learning how those pieces connect using Go’s standard `net/http` package.

## Features

- Create accounts and update email addresses and passwords.
- Sign in with hashed passwords, one-hour JWT access tokens, and revocable refresh tokens.
- Create, read, and delete chirps, with a 140-byte limit and basic word filtering.
- Filter chirps by author and sort them by creation time.
- Upgrade users to Chirpy Red through an API-key-protected Polka webhook.
- Serve a basic welcome page and track file-server requests.

## Install and run

You need Go **1.26.2 or newer**, a running PostgreSQL server, and the Goose migration CLI.

### 1. Get the code and dependencies

```bash
git clone https://github.com/benjnicholls/chirpy.git
cd chirpy
go mod download
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Make sure your Go binary directory (usually `$(go env GOPATH)/bin`) is on your `PATH` so you can run `goose`.

### 2. Create the database

Using a PostgreSQL role that can create databases:

```bash
sudo -u postgres psql
CREATE DATABASE chirpy;
\c chirpy # Test connection
```

Use your own PostgreSQL role if it differs from `postgres`.

### 3. Configure the environment

Create a `.env` file in the project root:

```dotenv
DB_URL="postgres://postgres:YOUR_PASSWORD@localhost:5432/chirpy?sslmode=disable"
PLATFORM="dev"
SECRET="REPLACE_WITH_RANDOM_SECRET"
POLKA_KEY="REPLACE_WITH_WEBHOOK_API_KEY"
```

- `DB_URL`: your PostgreSQL connection string. The example disables TLS for local development.
- `PLATFORM`: use `prod` for normal use. Setting it to `dev` enables `POST /admin/reset`, which deletes all users and their associated chirps and refresh tokens.
- `SECRET`: the JWT signing secret. Generate one with `openssl rand -base64 32`.
- `POLKA_KEY`: the shared webhook API key. A nonempty local test value is enough to start the server.

All four values are required, and the server expects `.env` to exist. This file is ignored by Git; keep credentials out of commits.

### 4. Apply migrations

Run from the project root, using the same connection string as `DB_URL`:

```bash
goose -dir sql/schema postgres 'postgres://postgres:YOUR_PASSWORD@localhost:5432/chirpy?sslmode=disable' up
```

Generated database code is already included, so sqlc is not required to run the project.

### 5. Start the server

```bash
go run .
```

The server listens on **http://localhost:8080**. Visit **http://localhost:8080/app/** for the welcome page, or check the health endpoint:

```bash
curl http://localhost:8080/api/healthz
# OK
```

## Try the API

Create an account:

```bash
curl -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"example-password"}'
```

Sign in:

```bash
curl -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"example-password"}'
```

Copy the response’s `token` into the request below to publish a chirp:

```bash
curl -X POST http://localhost:8080/api/chirps \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{"body":"Hello, Chirpy!"}'
```

Read chirps without signing in:

```bash
curl 'http://localhost:8080/api/chirps?sort=desc'
```

## Tests

```bash
go test ./...
```

The included tests cover authentication helpers and chirp word filtering.
