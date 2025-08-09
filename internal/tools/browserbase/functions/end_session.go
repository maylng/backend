package functions

import (
	"fmt"
	"os"

	"github.com/maylng/backend/internal/tools/browserbase/api"
	"github.com/playwright-community/playwright-go"
)

// Ends a Browserbase session and disconnects Playwright from it
func EndBrowserbaseSession(session *api.SessionResponse, browser playwright.Browser, page playwright.Page, pw *playwright.Playwright) error {
	// Close the page if open
	if page != nil {
		_ = page.Close()
	}
	// Close the browser if open
	if browser != nil {
		_ = browser.Close()
	}
	// Stop Playwright
	if pw != nil {
		_ = pw.Stop()
	}

	// Request session release from Browserbase
	apiKey := os.Getenv("BROWSERBASE_API_KEY")
	if session == nil {
		return fmt.Errorf("session is nil")
	}
	req := &api.UpdateSessionRequest{
		ProjectID: session.ProjectID,
		Status:    "REQUEST_RELEASE",
	}
	_, err := api.UpdateSession(apiKey, session.ID, req)
	if err != nil {
		return fmt.Errorf("failed to end browserbase session: %w", err)
	}

	return nil
}
