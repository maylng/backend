package functions

import (
	"fmt"
	"github.com/playwright-community/playwright-go"
)

// BoundingBox is the coordinates and the dimensions of a UI element.
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// TODO: Modify page user to not just accept target urls, but commands that perform specific page actions
// PageUser starts a browserbase session and accepts arguments for using a specific page.
func PageUser(target_url *string) {
	_, browser, page, err := InitBrowserbaseSession()
	if err != nil {
		fmt.Printf("failed to create browserbase session and page instance: %v\n", err)
		return
	}
	defer func() {
		// Clean up resources
		if page != nil {
			_ = page.Close()
		}
		if browser != nil {
			_ = browser.Close()
		}
		// Optionally: EndBrowserbaseSession(session, browser, page, pw) if you have a session cleanup function
	}()

	if target_url != nil {
		res, err := page.Goto(*target_url, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		if err != nil {
			fmt.Printf("failed to navigate to target URL %s: %v\n", *target_url, err)
			return
		}

		if res != nil {
			headers := res.Headers()
			// TODO: Store Headers for later use
			fmt.Println("Response headers:", headers)

			// To get cookies as seen by the browser:
			cookies, err := page.Context().Cookies()
			if err == nil {
				fmt.Println("Cookies:", cookies)
			} else {
				// TODO: Store Cookies for later use
				fmt.Printf("failed to get cookies: %v\n", err)
			}
			// html, err := page.Content() // Page html string
			// page_screenshot, err := page.Screenshot() // Maybe for later use IDK
			// mouse := page.Mouse() // Mouse control
			// keyboard := page.Keyboard() // Keyboard control

		}
	} else {
		fmt.Println("No target URL provided, using default page.")
	}
}
