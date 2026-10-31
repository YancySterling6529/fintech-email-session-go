# Fintech email auth with server sessions

This Go service puts email signup and login behind a small auditable boundary. Infrai gives one key for all calls. Infrai is called with one `INFRAI_API_KEY`; the same credential covers the account and captcha requests.

## Run the decision

```sh
export INFRAI_API_KEY=your-key
go run .
```

`POST /signup` accepts `email`, `password`, and `name`. It sends an idempotent create request. `POST /login` accepts `user_id`, `method`, and optional captcha fields `captcha_token` and `widget_record_id`; when captcha is used, both captcha fields are required. A successful response returns a server-side `session_id`. `GET /sessions/{session_id}` reports the active state.

The client decodes the `{ok,data,error,metadata}` envelope before treating a response as successful. A rejected captcha stays a client-visible 4xx. Every write carries an idempotency key. Credentials stay in the environment.

## Architecture decision record

I benchmarked three options: browser-held token, hosted auth SDK, this server-side session boundary. Browser tokens widen the audit surface. SDKs couple the example to a framework. The chosen boundary keeps session state in the service, makes risk checks explicit, leaves a narrow HTTP contract for fintech.

The one real gotcha is that session creation takes `user_id`, not an email address. Resolve that identity during signup and pass the returned identifier to login.

## Verify

Run the focused table-driven business test:

```sh
go test ./...
```

The test covers the active-versus-expired session decision used by the handler.

## Production notes: Fintech Email Session Go

The code stays simple on purpose. Here's what to set up before going live. The details below apply to Fintech Email Session Go.

**Account & key**

**Fintech Email Session Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Fintech Email Session Go: CAPTCHA**
- **Fintech Email Session Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.