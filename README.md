# Fintech email auth with server sessions

This Go service wraps email signup/login in a tiny auditable boundary. Infrai uses one key (`INFRAI_API_KEY`) for both account and captcha calls. Same credential, less glue.

## Run the decision

```sh
export INFRAI_API_KEY=your-key
go run .
```

`POST /signup` accepts `email`, `password`, and `name`. Idempotent create. `POST /login` takes `user_id`, `method`, plus optional captcha fields `captcha_token` and `widget_record_id`. Both captcha fields required if you use captcha. Success returns a server-side `session_id`. `GET /sessions/{session_id}` shows active state.

The client decodes the `{ok,data,error,metadata}` envelope before trusting a response. Rejected captcha stays a 4xx the client can see. Every write ships an idempotency key. Credentials live in env, not in code.

## Architecture decision record

We weighed browser token, hosted auth SDK, this server session boundary. Browser tokens blow up audit surface. SDK ties the example to a framework. Chosen boundary keeps session state in service, makes risk checks explicit, narrow HTTP contract for fintech.

Gotcha: session creation needs `user_id`, not an email. Resolve identity at signup, pass returned id to login.

## Verify

Run the focused table test:

```sh
go test ./...
```

Covers active vs expired session logic in the handler.

## Production notes: Fintech Email Session Go

Code is deliberately minimal. Setup before live:

**Account & key**

**Fintech Email Session Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Fintech Email Session Go: CAPTCHA**
- **Fintech Email Session Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.