package views

// ViewLoaded is a custom vaxis event posted when a view finishes loading data.
// It is sent from background goroutines via PostEvent to notify the UI.
type ViewLoaded struct {
	Tab int
	Err error
}

// DashboardUpdated is posted by subscription goroutines when new realtime
// or app stats data arrives, triggering a redraw.
type DashboardUpdated struct{}
