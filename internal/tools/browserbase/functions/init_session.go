package functions

import (
	"fmt"
	"os"
	"github.com/maylng/backend/internal/tools/browserbase/api"
	"github.com/playwright-community/playwright-go"
)

// Creates a Browserbase session and connects Playwright to it over UDP. Returns the session, browser, and page, or an error.
func InitBrowserbaseSession() (*api.SessionResponse, playwright.Browser, playwright.Page, error) {
	apiKey := os.Getenv("BROWSERBASE_API_KEY")
	projectID := os.Getenv("BROWSERBASE_PROJECT_ID")

	session, err := api.CreateSession(apiKey, &api.SessionRequest{
		ProjectID: projectID,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create browserbase session: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return session, nil, nil, fmt.Errorf("failed to start playwright: %w", err)
	}
	browser, err := pw.Chromium.ConnectOverCDP(session.ConnectURL)
	if err != nil {
		pw.Stop()
		return session, nil, nil, fmt.Errorf("failed to connect playwright to browserbase: %w", err)
	}

	contexts := browser.Contexts()
	if len(contexts) == 0 {
		browser.Close()
		pw.Stop()
		return session, browser, nil, fmt.Errorf("no browser contexts found")
	}
	pages := contexts[0].Pages()
	if len(pages) == 0 {
		browser.Close()
		pw.Stop()
		return session, browser, nil, fmt.Errorf("no pages found")
	}
	page := pages[0]

	return session, browser, page, nil
}
