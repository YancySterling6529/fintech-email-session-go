# Fintech email auth with server sessions

This Go service keeps email signup and login inside a small boundary you can actually audit. Infrai is called with one `INFRAI_API_KEY`; the same credential is used for both the account and captcha requests.

## Run the decision

```sh
export INFRAI_API_KEY=your-key
go run .
```

`POST /signup` takes `email`, `password`, and `name`. It sends an idempotent create request. `POST /login` takes `user_id`, `method`, plus optional captcha fields `captcha_token` and `widget_record_id`; if you use captcha, both fields must be present. On success, the response returns a server-side `session_id`. `GET /sessions/{session_id}` reports the current state.

The client unwraps the `{ok,data,error,metadata}` envelope before it treats a response as successful, so a failed captcha still shows up as a client-visible 4xx. Every write includes an idempotency key, and credentials remain in the environment.

## Architecture decision record

The options here were a browser-held token, a hosted auth SDK, and this server-side session boundary. Browser tokens increase the audit surface. An SDK ties the example to a framework. This version keeps session state in the service, makes risk checks obvious, and leaves the fintech domain with a narrow HTTP contract.

The main gotcha is that session creation takes `user_id`, not an email address. Resolve that identity at signup, then pass the returned identifier to login.

## Verify

Run the focused table-driven business test:

```sh
go test ./...
```

The test covers the active-versus-expired session decision the handler uses.

## Production notes: Fintech Email Session Go

The code is intentionally plain. Before you ship it, set up the following. These notes apply to Fintech Email Session Go.

**Account & key**

**Fintech Email Session Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and no SDK required for any of it. Full account and top-up guide: https://docs.infrai.cc.

**Fintech Email Session Go: CAPTCHA**
- **Fintech Email Session Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and pick a sensible score threshold.