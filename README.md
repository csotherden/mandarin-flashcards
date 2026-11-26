# Mandarin Cards

A mobile-first flashcard app for drilling Mandarin vocabulary with tone-level pronunciation feedback.

The Go API follows the what-iff shape: Ent on Postgres, local JWT auth (argon2id with a per-user salt), one file per handler function, and a local file store. The UI is React. Docker Compose runs both on a local server. There is no Cognito and no S3.

## Run

```bash
docker compose up --build
```

- Web: http://localhost:8091
- API: http://localhost:8090/api/health/live

The first boot generates `JWT_SECRET` and `JWT_REFRESH_SECRET` into a Docker volume. Postgres is published on port 5433 so it does not collide with another local database.

Create an account in the web UI, then add cards. Auto-fill, photo search, and pronunciation grading stay dark until you set the keys in `.env` (see `.env.example`) and recreate the API container:

- `TONEPERFECT_API_KEY` — `POST /v1/assess` on TonePerfect
- `UNSPLASH_ACCESS_KEY` — photo search
- `LLM_API_KEY` — card auto-fill and coaching tips (`LLM_BASE_URL` / `LLM_MODEL` for an OpenAI-compatible endpoint, including Gemini)

Recordings are normalized to 16 kHz mono WAV with ffmpeg before they are sent to TonePerfect. Speaking drills call `POST /api/assess`.

## Native development

Start Postgres with `docker compose up db`, copy `.env.example` to `.env`, set the two JWT secrets, then:

```bash
go run ./cmd/api-server
cd web && npm install && npm run dev
```

The Vite dev server proxies `/api` to port 8090.
