//go:build cgo && (linux || windows || darwin)

package tray

import (
	_ "embed"

	"github.com/getlantern/systray"

	"github.com/DSerejo/lichess-puzzle-mixer/internal/browser"
)

//go:embed icon.png
var iconPNG []byte

func available() bool { return true }

func run(appURL string, onExit func()) {
	systray.Run(func() {
		systray.SetIcon(iconPNG)
		systray.SetTooltip("Lichess Puzzle Mixer")

		openItem := systray.AddMenuItem("Open in browser", "Open the app in your default browser")
		systray.AddSeparator()
		quitItem := systray.AddMenuItem("Quit", "Stop the server and exit")

		go func() {
			for {
				select {
				case <-openItem.ClickedCh:
					_ = browser.Open(appURL)
				case <-quitItem.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}, onExit)
}

func quit() {
	systray.Quit()
}
