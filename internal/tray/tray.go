package tray

// Available reports whether a system tray icon will be shown for this build.
func Available() bool {
	return available()
}

// Run shows the tray icon until Quit is called. onReady runs after the icon is
// visible; onExit runs when the user chooses Quit (call shutdown from onExit).
func Run(appURL string, onExit func()) {
	run(appURL, onExit)
}

// Quit removes the tray icon and unblocks Run.
func Quit() {
	quit()
}
