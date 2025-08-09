package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type SessionRequest struct {
	ProjectID       string                 `json:"projectId"`
	ExtensionID     string                 `json:"extensionId,omitempty"`
	BrowserSettings *BrowserSettings       `json:"browserSettings,omitempty"`
	Timeout         int                    `json:"timeout,omitempty"`
	KeepAlive       *bool                  `json:"keepAlive,omitempty"`
	Proxies         interface{}            `json:"proxies,omitempty"` // bool or []ProxyConfig
	Region          string                 `json:"region,omitempty"`
	UserMetadata    map[string]interface{} `json:"userMetadata,omitempty"`
}

type BrowserSettings struct {
	Context              *BrowserContext `json:"context,omitempty"`
	ExtensionID          string          `json:"extensionId,omitempty"`
	Fingerprint          *Fingerprint    `json:"fingerprint,omitempty"`
	Viewport             *Viewport       `json:"viewport,omitempty"`
	BlockAds             *bool           `json:"blockAds,omitempty"`
	SolveCaptchas        *bool           `json:"solveCaptchas,omitempty"`
	RecordSession        *bool           `json:"recordSession,omitempty"`
	LogSession           *bool           `json:"logSession,omitempty"`
	AdvancedStealth      *bool           `json:"advancedStealth,omitempty"`
	CaptchaImageSelector string          `json:"captchaImageSelector,omitempty"`
	CaptchaInputSelector string          `json:"captchaInputSelector,omitempty"`
}

type BrowserContext struct {
	ID      string `json:"id"`
	Persist *bool  `json:"persist,omitempty"`
}

type Fingerprint struct {
	HTTPVersion      string   `json:"httpVersion,omitempty"`
	Browsers         []string `json:"browsers,omitempty"`
	Devices          []string `json:"devices,omitempty"`
	Locales          []string `json:"locales,omitempty"`
	OperatingSystems []string `json:"operatingSystems,omitempty"`
	Screen           *Screen  `json:"screen,omitempty"`
}

type Screen struct {
	MaxHeight int `json:"maxHeight,omitempty"`
	MaxWidth  int `json:"maxWidth,omitempty"`
	MinHeight int `json:"minHeight,omitempty"`
	MinWidth  int `json:"minWidth,omitempty"`
}

type Viewport struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

type ProxyConfig struct {
	Type          string       `json:"type"`
	Geolocation   *Geolocation `json:"geolocation,omitempty"`
	DomainPattern string       `json:"domainPattern,omitempty"`
	// External proxy fields
	Server   string `json:"server,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type Geolocation struct {
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	Country string `json:"country,omitempty"`
}

type SessionResponse struct {
	ID                string                 `json:"id"`
	CreatedAt         time.Time              `json:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt"`
	ProjectID         string                 `json:"projectId"`
	StartedAt         *time.Time             `json:"startedAt,omitempty"`
	EndedAt           *time.Time             `json:"endedAt,omitempty"`
	ExpiresAt         *time.Time             `json:"expiresAt,omitempty"`
	Status            string                 `json:"status"`
	ProxyBytes        int64                  `json:"proxyBytes"`
	AvgCPUUsage       int64                  `json:"avgCpuUsage"`
	MemoryUsage       int64                  `json:"memoryUsage"`
	KeepAlive         bool                   `json:"keepAlive"`
	ContextID         string                 `json:"contextId"`
	Region            string                 `json:"region"`
	UserMetadata      map[string]interface{} `json:"userMetadata"`
	ConnectURL        string                 `json:"connectUrl"`
	SeleniumRemoteURL string                 `json:"seleniumRemoteUrl"`
	SigningKey        string                 `json:"signingKey"`
}

type UpdateSessionRequest struct {
	ProjectID string `json:"projectId"`
	Status    string `json:"status"` // Only allowed value: REQUEST_RELEASE
}

type SessionDebugResponse struct {
	DebuggerFullscreenURL string             `json:"debuggerFullscreenUrl"`
	DebuggerURL           string             `json:"debuggerUrl"`
	Pages                 []SessionDebugPage `json:"pages"`
	WSUrl                 string             `json:"wsUrl"`
}

type SessionDebugPage struct {
	ID                    string `json:"id"`
	URL                   string `json:"url"`
	FaviconURL            string `json:"faviconUrl"`
	Title                 string `json:"title"`
	DebuggerURL           string `json:"debuggerUrl"`
	DebuggerFullscreenURL string `json:"debuggerFullscreenUrl"`
}

type SessionLogEntry struct {
	Method    string             `json:"method"`
	PageID    int                `json:"pageId"`
	SessionID string             `json:"sessionId"`
	Request   SessionLogRequest  `json:"request"`
	Response  SessionLogResponse `json:"response"`
	Timestamp int64              `json:"timestamp"`
	FrameID   string             `json:"frameId"`
	LoaderID  string             `json:"loaderId"`
}

type SessionLogRequest struct {
	Timestamp int64                  `json:"timestamp"`
	Params    map[string]interface{} `json:"params"`
	RawBody   string                 `json:"rawBody"`
}

type SessionLogResponse struct {
	Timestamp int64                  `json:"timestamp"`
	Result    map[string]interface{} `json:"result"`
	RawBody   string                 `json:"rawBody"`
}

type SessionRecordingEvent struct {
	Data      map[string]interface{} `json:"data"`
	SessionID string                 `json:"sessionId"`
	Timestamp int64                  `json:"timestamp"`
	Type      int                    `json:"type"`
}

type SessionUploadResponse struct {
	Message string `json:"message"`
}

// CreateSession creates a Browserbase session with all supported options.
func CreateSession(apiKey string, reqBody *SessionRequest) (*SessionResponse, error) {
	url := "https://api.browserbase.com/v1/sessions"
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-BB-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var sessionResp SessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &sessionResp, nil
}

// UpdateSession updates a Browserbase session (e.g., to request release).
func UpdateSession(apiKey, sessionID string, reqBody *UpdateSessionRequest) (*SessionResponse, error) {
	url := "https://api.browserbase.com/v1/sessions/" + sessionID
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-BB-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var session SessionResponse
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &session, nil
}

// ListSessionsOptions defines query params for listing sessions
// Status: optional filter (RUNNING, ERROR, TIMED_OUT, COMPLETED)
// Q: optional user metadata query string
type ListSessionsOptions struct {
	Status string // optional
	Q      string // optional
}

// ListSessions fetches a list of Browserbase sessions with optional filters.
func ListSessions(apiKey string, opts *ListSessionsOptions) ([]SessionResponse, error) {
	url := "https://api.browserbase.com/v1/sessions"
	params := []string{}
	if opts != nil {
		if opts.Status != "" {
			params = append(params, "status="+opts.Status)
		}
		if opts.Q != "" {
			params = append(params, "q="+opts.Q)
		}
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-BB-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var sessions []SessionResponse
	if err := json.Unmarshal(body, &sessions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return sessions, nil
}

// GetSession fetches a single Browserbase session by ID.
func GetSession(apiKey, sessionID string) (*SessionResponse, error) {
	url := "https://api.browserbase.com/v1/sessions/" + sessionID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-BB-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var session SessionResponse
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &session, nil
}

// GetSessionDebug fetches live debug URLs and page info for a session.
func GetSessionDebug(apiKey, sessionID string) (*SessionDebugResponse, error) {
	url := "https://api.browserbase.com/v1/sessions/" + sessionID + "/debug"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-BB-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var debugResp SessionDebugResponse
	if err := json.Unmarshal(body, &debugResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &debugResp, nil
}

// GetSessionDownloads downloads the session's ZIP file from Browserbase.
func GetSessionDownloads(apiKey, sessionID string) ([]byte, error) {
	url := "https://api.browserbase.com/v1/sessions/" + sessionID + "/downloads"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-BB-API-Key", apiKey)
	// Content-Type not needed for downloads

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	zipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read zip response: %w", err)
	}

	return zipBytes, nil
}

// GetSessionLogs fetches logs for a Browserbase session.
func GetSessionLogs(apiKey, sessionID string) ([]SessionLogEntry, error) {
	url := "https://api.browserbase.com/v1/sessions/" + sessionID + "/logs"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-BB-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var logs []SessionLogEntry
	if err := json.Unmarshal(body, &logs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal logs: %w", err)
	}

	return logs, nil
}

// GetSessionRecording fetches recording events for a Browserbase session.
func GetSessionRecording(apiKey, sessionID string) ([]SessionRecordingEvent, error) {
	url := "https://api.browserbase.com/v1/sessions/" + sessionID + "/recording"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-BB-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var events []SessionRecordingEvent
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recording events: %w", err)
	}

	return events, nil
}

// CreateSessionUpload uploads a file to a Browserbase session.
func CreateSessionUpload(apiKey, sessionID, fieldName, fileName string, fileData []byte) (*SessionUploadResponse, error) {
	url := "https://api.browserbase.com/v1/sessions/" + sessionID + "/uploads"

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-BB-API-Key", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var uploadResp SessionUploadResponse
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &uploadResp, nil
}
