package functions

import (
	"fmt"
	"github.com/playwright-community/playwright-go"
)

type PageUser struct {
	Browser playwright.Browser
	Page    playwright.Page

}

type AgentCommand struct {
	Action    string  `json:"action"`
	Selector  string  `json:"selector,omitempty"`
	Text      string  `json:"text,omitempty"`
	URL       string  `json:"url,omitempty"`
	Direction string  `json:"direction,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	X         float64 `json:"x,omitempty"`
	Y         float64 `json:"y,omitempty"`
	Duration  float64 `json:"duration,omitempty"`
}

func NewPageUser() (*PageUser, error) {
	_, browser, page, err := InitBrowserbaseSession()
	if err != nil {
		return nil, err
	}
	return &PageUser{Browser: browser, Page: page}, nil
}

func (pu *PageUser) ExecuteCommand(cmd AgentCommand) error {
	switch cmd.Action {
	case "navigate":
		_, err := pu.Page.Goto(cmd.URL)
		return err
	case "type":
		locator := pu.Page.Locator(cmd.Selector)
		return locator.Fill(cmd.Text)
		// err := locator.Type("agent@example.com", playwright.LocatorTypeOptions{
		// 	Delay: playwright.Float(100), // milliseconds between keystrokes
		// })
	case "scroll":
		_, err := pu.Page.Evaluate(fmt.Sprintf(`window.scrollBy(0, %f)`, cmd.Amount))
		return err
	case "moveMouse":
		return pu.Page.Mouse().Move(cmd.X, cmd.Y, playwright.MouseMoveOptions{
			Steps: intPtr(int(cmd.Duration / 10)),
		})
	case "click":
		locator := pu.Page.Locator(cmd.Selector)
		box, err := locator.BoundingBox()
		if err != nil {
			return fmt.Errorf("failed to get bounding box for selector %s: %w", cmd.Selector, err)
		}
		mouseX := box.X + box.Width/2
		mouseY := box.Y + box.Height/2
		if cmd.X != 0 {
			mouseX = cmd.X
		}
		if cmd.Y != 0 {
			mouseY = cmd.Y
		}
		if err := pu.Page.Mouse().Move(mouseX, mouseY, playwright.MouseMoveOptions{
			Steps: intPtr(int(cmd.Duration / 10)),
		}); err != nil {
			return fmt.Errorf("failed to move mouse: %w", err)
		}
		if err := pu.Page.Mouse().Click(mouseX, mouseY); err != nil {
			return fmt.Errorf("failed to click: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown action: %s", cmd.Action)
	}
}

// Cleanup closes the page and browser.
func (pu *PageUser) Cleanup() {
	if pu.Page != nil {
		_ = pu.Page.Close()
	}
	if pu.Browser != nil {
		_ = pu.Browser.Close()
	}
}

func intPtr(i int) *int {
	return &i
}
