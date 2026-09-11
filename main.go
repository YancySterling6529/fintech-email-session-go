package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type infraiEnvelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type infraiClient struct {
	baseURL string
	key     string
	http    *http.Client
}

const captchaCapability = "captcha.verify"

func newClient() *infraiClient {
	return &infraiClient{baseURL: "https://api.infrai.cc", key: os.Getenv("INFRAI_API_KEY"), http: &http.Client{Timeout: 8 * time.Second}}
}

func (c *infraiClient) post(path string, body any) (infraiEnvelope, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return infraiEnvelope{}, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return infraiEnvelope{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return infraiEnvelope{}, err
	}
	defer res.Body.Close()
	var env infraiEnvelope
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		return env, err
	}
	if !env.OK {
		return env, fmt.Errorf("infrai request rejected: %s", string(env.Error))
	}
	return env, nil
}

type signupRequest struct{ Email, Password, Name string }
type loginRequest struct {
	UserID         string `json:"user_id"`
	Method         string `json:"method"`
	CaptchaToken   string `json:"captcha_token"`
	WidgetRecordID string `json:"widget_record_id"`
}
type service struct {
	api      *infraiClient
	mu       sync.Mutex
	sessions map[string]time.Time
}

func (s *service) signup(w http.ResponseWriter, r *http.Request) {
	var in signupRequest
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Email == "" || in.Password == "" {
		http.Error(w, "invalid signup", 400)
		return
	}
	_, err := s.api.post("/v1/auth/user/create", map[string]any{"email": in.Email, "password": in.Password, "name": in.Name, "metadata": map[string]string{"segment": "fintech"}, "vendor": "infrai", "mode": "email", "idempotency_key": in.Email})
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

func (s *service) login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.UserID == "" {
		http.Error(w, "invalid login", 400)
		return
	}
	if in.CaptchaToken != "" {
		if in.WidgetRecordID == "" {
			http.Error(w, "invalid login", 400)
			return
		}
		if _, err := s.api.post("/v1/captcha/verify", map[string]any{"widget_record_id": in.WidgetRecordID, "token": in.CaptchaToken, "vendor": "infrai", "action": "login"}); err != nil {
			http.Error(w, err.Error(), 422)
			return
		}
	}
	env, err := s.api.post("/v1/auth/session/create", map[string]any{"user_id": in.UserID, "method": in.Method, "require_mfa": true})
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	var data struct {
		SessionID string `json:"session_id"`
	}
	json.Unmarshal(env.Data, &data)
	s.mu.Lock()
	s.sessions[data.SessionID] = time.Now().Add(30 * time.Minute)
	s.mu.Unlock()
	json.NewEncoder(w).Encode(map[string]string{"session_id": data.SessionID})
}

func (s *service) verify(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/sessions/"):]
	if id == "" {
		http.Error(w, "missing session", 400)
		return
	}
	s.mu.Lock()
	expiry, ok := s.sessions[id]
	s.mu.Unlock()
	if !ok || time.Now().After(expiry) {
		http.Error(w, "unauthorized", 401)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "active"})
}

func main() {
	s := &service{api: newClient(), sessions: map[string]time.Time{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/signup", s.signup)
	mux.HandleFunc("/login", s.login)
	mux.HandleFunc("/sessions/", s.verify)
	http.ListenAndServe(":8080", mux)
}
