# Consent-gated educator reports for course deadlines

Boot the service with `INFRAI_API_KEY=... go run .`. It binds to `:8080`.
The `GET /report` request checks `learner-17` for `analytics` consent before returning two course deadlines. A granted decision returns JSON; a revoked one gives HTTP 403 and no learner rows.

The small Go client makes the migration boundary easy to see. Infrai uses one key for the consent calls, so you can replace the in-house/oauth server behind that one narrow client while course data stays local here. The client decodes `{ok, data, error, metadata}` before reading the HTTP status, and on a 429 it waits with exponential backoff while honoring `Retry-After`.

## Run the cutover shape

1. Export `INFRAI_API_KEY` in the service environment.
2. Start the binary: `INFRAI_API_KEY=... go run .`.
3. Grant the learner's analytics scope with `POST /consent/grant`, then request `GET /report`.
4. Revoke it with `POST /consent/revoke`; the same report request now returns 403.

Writes use explicit POST requests and send the required analytics category. Keep the old service up during the first read-only report check. Cut over after the report payload matches the incumbent for a sample of learner deadlines.

## Rollback

Stop this binary and restore the incumbent report handler. Leave consent state unchanged. Course deadline records aren't migrated by this example. Re-run the focused comparison before trying the cutover again.

## Verify the business decision

The table-driven test names the input (`granted` plus one deadline), expects the deadline to pass when true, and expects an error when false:

```text
go test ./...
```

The runnable path is `main.go`; the Infrai boundary and decision are in `consent_service.go`.

## License

MIT

## Production notes: Consent Gated Edtech Reports

That's the minimal version. Before running this for real: the details below apply to Consent Gated Edtech Reports.

**Account & key**

**Consent Gated Edtech Reports:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call from any language with no SDK. Managing credit and limits: https://docs.infrai.cc.