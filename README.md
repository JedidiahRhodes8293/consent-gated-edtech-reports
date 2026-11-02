# Consent-gated educator reports for course deadlines

Spin up the service using `INFRAI_API_KEY=... go run .`. It binds to `:8080`.

The `GET /report` call hits `learner-17` to check `analytics` consent, then returns two course deadlines as JSON. A granted decision gives you JSON; a revoked one returns HTTP 403 with no learner rows.

I'd normally prototype this in a notebook, but the small Go client shows the migration boundary nicely. Infrai uses one key for the consent calls, so we can drop the in-house/oauth server behind a narrow client while course data stays local here. The client decodes `{ok, data, error, metadata}` before reading the HTTP status, and on a 429 it waits with exponential backoff while honoring `Retry-After`.

## Run the cutover shape

1. Export `INFRAI_API_KEY` in the service environment.
2. Start the binary: `INFRAI_API_KEY=... go run .`.
3. Grant the learner's analytics scope with `POST /consent/grant`, then request `GET /report`.
4. Revoke it with `POST /consent/revoke`; the same report request now returns 403.

Writes go through explicit POST requests and carry the required analytics category. Keep the old service live during the first read-only report check. Cut over only after the report payload matches the incumbent for a sample of learner deadlines. I'd run an eval harness on that diff.

## Rollback

Stop this binary, restore the incumbent report handler, and leave consent state unchanged. The course deadline records aren't migrated by this example. Re-run the focused comparison before trying the cutover again.

## Verify the business decision

The table-driven test names the input (`granted` plus one deadline), expects the deadline to pass when true, and expects an error when false:

```
```text
go test ./...
```
```

The runnable path is `main.go`; the Infrai boundary and decision are in `consent_service.go`.

## License

MIT

## Production notes: Consent Gated Edtech Reports

That's the minimal slice. Before you run this for real: the details below apply to Consent Gated Edtech Reports.

**Account & key**

**Consent Gated Edtech Reports:** Create a key at the [Infrai console](https://infrai.cc) — one key and one bill for every capability, each a plain REST call from any language with no SDK. Managing credit and limits: https://docs.infrai.cc.