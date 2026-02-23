package views

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"git.sr.ht/~rockorager/vaxis/vxfw/list"
	"git.sr.ht/~rockorager/vaxis/vxfw/richtext"
	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-tui/widgets"
	"github.com/dustin/go-humanize"
	"golang.org/x/sync/errgroup"
)

// DashboardViewParams holds configuration for creating a DashboardView.
type DashboardViewParams struct {
	System     truenas.SystemServiceAPI
	Reporting  truenas.ReportingServiceAPI
	Interfaces truenas.InterfaceServiceAPI
	Apps       truenas.AppServiceAPI
	PostEvent  func(vaxis.Event)
}

// DashboardView displays a real-time monitoring dashboard.
type DashboardView struct {
	// Services
	systemSvc truenas.SystemServiceAPI
	reportSvc truenas.ReportingServiceAPI
	ifaceSvc  truenas.InterfaceServiceAPI
	appsSvc   truenas.AppServiceAPI

	// One-time data (from Load)
	sysInfo    *truenas.SystemInfo
	sysVersion string
	interfaces []truenas.NetworkInterface
	apps       []truenas.App

	// Streaming state (protected by mu)
	mu       sync.Mutex
	realtime *truenas.RealtimeUpdate
	appStats map[string]truenas.AppStats
	cpuSpark *widgets.Sparkline

	// Subscriptions
	realtimeSub *truenas.Subscription[truenas.RealtimeUpdate]
	statsSub    *truenas.Subscription[[]truenas.AppStats]
	cancelSubs  context.CancelFunc

	// UI
	appList   list.Dynamic
	appRows   []appRow
	loaded    bool
	postEvent func(vaxis.Event)

	// RetryBaseDelay is the base delay for subscription retry backoff.
	// Defaults to 1s; tests can set to a small value.
	RetryBaseDelay time.Duration
}

type appRow struct {
	Name     string
	State    string
	CPUUsage float64
	Memory   int64
}

// NewDashboardView creates a DashboardView backed by the given services.
func NewDashboardView(p DashboardViewParams) *DashboardView {
	dv := &DashboardView{
		systemSvc: p.System,
		reportSvc: p.Reporting,
		ifaceSvc:  p.Interfaces,
		appsSvc:   p.Apps,
		postEvent: p.PostEvent,
		cpuSpark:  widgets.NewSparkline(60),
		appStats:  make(map[string]truenas.AppStats),
	}
	dv.appList.DrawCursor = true
	dv.appList.Builder = dv.buildAppItem
	return dv
}

// Load fetches initial data for the dashboard (system info, version, interfaces, apps).
func (dv *DashboardView) Load(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	var sysInfo *truenas.SystemInfo
	var version string
	var ifaces []truenas.NetworkInterface
	var apps []truenas.App

	g.Go(func() error {
		info, err := dv.systemSvc.GetInfo(gctx)
		if err != nil {
			return fmt.Errorf("system.info: %w", err)
		}
		sysInfo = info
		return nil
	})

	g.Go(func() error {
		v, err := dv.systemSvc.GetVersion(gctx)
		if err != nil {
			return fmt.Errorf("system.version: %w", err)
		}
		version = v
		return nil
	})

	g.Go(func() error {
		list, err := dv.ifaceSvc.List(gctx)
		if err != nil {
			return fmt.Errorf("interface.list: %w", err)
		}
		ifaces = list
		return nil
	})

	g.Go(func() error {
		list, err := dv.appsSvc.ListApps(gctx)
		if err != nil {
			return fmt.Errorf("app.list: %w", err)
		}
		apps = list
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	dv.sysInfo = sysInfo
	dv.sysVersion = version
	dv.interfaces = ifaces
	dv.apps = apps
	dv.rebuildAppRows()
	dv.loaded = true
	return nil
}

// Loaded reports whether data has been successfully fetched.
func (dv *DashboardView) Loaded() bool {
	return dv.loaded
}

// StartSubscriptions begins streaming realtime and app stats data.
func (dv *DashboardView) StartSubscriptions(ctx context.Context) {
	subCtx, cancel := context.WithCancel(ctx)
	dv.cancelSubs = cancel

	go dv.runRealtimeSub(subCtx)
	go dv.runStatsSub(subCtx)
}

// StopSubscriptions terminates all active subscriptions.
func (dv *DashboardView) StopSubscriptions() {
	if dv.cancelSubs != nil {
		dv.cancelSubs()
	}
	if dv.realtimeSub != nil {
		dv.realtimeSub.Close()
	}
	if dv.statsSub != nil {
		dv.statsSub.Close()
	}
}

// retryBackoff sleeps with exponential backoff, returning false if ctx is cancelled.
func (dv *DashboardView) retryBackoff(ctx context.Context, attempt int) bool {
	base := dv.RetryBaseDelay
	if base == 0 {
		base = time.Second
	}
	delay := base * time.Duration(1<<min(attempt, 5)) // base*1, base*2, base*4, ... base*32 max
	select {
	case <-ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}

func (dv *DashboardView) runRealtimeSub(ctx context.Context) {
	for attempt := 0; ; attempt++ {
		sub, err := dv.reportSvc.SubscribeRealtime(ctx)
		if err != nil {
			log.Printf("realtime subscription failed: %v (attempt %d)", err, attempt+1)
			if !dv.retryBackoff(ctx, attempt) {
				return
			}
			continue
		}
		dv.mu.Lock()
		dv.realtimeSub = sub
		dv.mu.Unlock()
		attempt = 0

		for {
			select {
			case <-ctx.Done():
				return
			case update, ok := <-sub.C:
				if !ok {
					log.Printf("realtime subscription closed, reconnecting...")
					break
				}
				dv.mu.Lock()
				dv.realtime = &update

				// Compute average CPU usage across all cores
				if len(update.CPU) > 0 {
					var total float64
					for _, cpu := range update.CPU {
						total += cpu.Usage
					}
					dv.cpuSpark.Push(total / float64(len(update.CPU)))
				}

				dv.mu.Unlock()
				if dv.postEvent != nil {
					dv.postEvent(DashboardUpdated{})
				}
				continue
			}
			break
		}
	}
}

func (dv *DashboardView) runStatsSub(ctx context.Context) {
	for attempt := 0; ; attempt++ {
		sub, err := dv.appsSvc.SubscribeStats(ctx)
		if err != nil {
			log.Printf("stats subscription failed: %v (attempt %d)", err, attempt+1)
			if !dv.retryBackoff(ctx, attempt) {
				return
			}
			continue
		}
		dv.mu.Lock()
		dv.statsSub = sub
		dv.mu.Unlock()
		attempt = 0

		for {
			select {
			case <-ctx.Done():
				return
			case stats, ok := <-sub.C:
				if !ok {
					log.Printf("stats subscription closed, reconnecting...")
					break
				}
				dv.mu.Lock()
				for _, s := range stats {
					dv.appStats[s.AppName] = s
				}
				dv.rebuildAppRows()
				dv.mu.Unlock()
				if dv.postEvent != nil {
					dv.postEvent(DashboardUpdated{})
				}
				continue
			}
			break
		}
	}
}

func (dv *DashboardView) rebuildAppRows() {
	rows := make([]appRow, 0, len(dv.apps))
	for _, a := range dv.apps {
		row := appRow{
			Name:  a.Name,
			State: a.State,
		}
		if stats, ok := dv.appStats[a.Name]; ok {
			row.CPUUsage = stats.CPUUsage
			row.Memory = stats.Memory
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].CPUUsage > rows[j].CPUUsage
	})
	dv.appRows = rows
}

// Fixed-width columns for the apps table (CPU%, MEM, STATE).
// The name column width is computed dynamically to fill remaining space.
const (
	appColCPUWidth   = 8
	appColMemWidth   = 10
	appColStateWidth = 10
	appColGap        = 2
	// Total fixed portion: CPU + gap + MEM + gap + STATE
	appFixedWidth = appColCPUWidth + appColGap + appColMemWidth + appColGap + appColStateWidth
)

// appCols returns column definitions with the name column sized to fill width.
func appCols(totalWidth int) []widgets.TableColumn {
	nameWidth := totalWidth - appFixedWidth - appColGap // subtract gap after name
	if nameWidth < 12 {
		nameWidth = 12
	}
	return []widgets.TableColumn{
		{Width: nameWidth},                            // Name (dynamic)
		{Width: appColCPUWidth, AlignRight: true},     // CPU%
		{Width: appColMemWidth, AlignRight: true},     // MEM
		{Width: appColStateWidth},                     // STATE
	}
}

// appRowWidget renders a single app row using WriteCell for exact column alignment.
type appRowWidget struct {
	cells  []string
	styles []vaxis.Style
}

func (w *appRowWidget) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	s := vxfw.NewSurface(ctx.Max.Width, 1, w)
	cols := appCols(int(ctx.Max.Width))
	col := 0
	for i, c := range cols {
		if col >= int(ctx.Max.Width) {
			break
		}
		text := ""
		if i < len(w.cells) {
			text = w.cells[i]
		}
		style := vaxis.Style{}
		if i < len(w.styles) {
			style = w.styles[i]
		}
		writeCell(&s, uint16(col), 0, c.Width, text, style, c.AlignRight)
		col += c.Width + appColGap
	}
	return s, nil
}

func (w *appRowWidget) HandleEvent(ev vaxis.Event, phase vxfw.EventPhase) (vxfw.Command, error) {
	return nil, nil
}

// writeCell writes text into surf at (col, row) within maxWidth, right-aligning if requested.
func writeCell(surf *vxfw.Surface, col, row uint16, maxWidth int, s string, style vaxis.Style, alignRight bool) {
	chars := vaxis.Characters(s)
	displayWidth := 0
	for _, ch := range chars {
		displayWidth += ch.Width
	}
	offset := 0
	if alignRight && displayWidth < maxWidth {
		offset = maxWidth - displayWidth
	}
	pos := offset
	for _, ch := range chars {
		if pos+ch.Width > maxWidth {
			break
		}
		surf.WriteCell(col+uint16(pos), row, vaxis.Cell{
			Character: ch,
			Style:     style,
		})
		pos += ch.Width
	}
}

func (dv *DashboardView) buildAppItem(i uint, cursor uint) vxfw.Widget {
	dv.mu.Lock()
	defer dv.mu.Unlock()

	if int(i) >= len(dv.appRows) {
		return nil
	}
	row := dv.appRows[i]

	stateColor := vaxis.IndexColor(2) // green
	if row.State != "RUNNING" {
		stateColor = vaxis.IndexColor(1) // red
	}

	memStr := ""
	if row.Memory > 0 {
		memStr = humanize.Bytes(uint64(row.Memory))
	}

	return &appRowWidget{
		cells: []string{
			" " + row.Name,
			fmt.Sprintf("%.2f%%", row.CPUUsage),
			memStr,
			row.State,
		},
		styles: []vaxis.Style{
			{},
			{},
			{},
			{Foreground: stateColor},
		},
	}
}

// Draw renders the dashboard.
func (dv *DashboardView) Draw(ctx vxfw.DrawContext) (vxfw.Surface, error) {
	if !dv.loaded {
		return drawLoadingState(ctx, dv)
	}

	dv.mu.Lock()
	rt := dv.realtime
	sparkCount := dv.cpuSpark.Count()
	dv.mu.Unlock()

	s := vxfw.NewSurface(ctx.Max.Width, ctx.Max.Height, dv)
	row := 0
	barWidth := 20

	// === Header row ===
	uptimeStr := ""
	if dv.sysInfo != nil && dv.sysInfo.UptimeSeconds > 0 {
		uptimeStr = FormatUptime(dv.sysInfo.UptimeSeconds)
	}
	headerSegments := []vaxis.Segment{
		{Text: " " + dv.sysInfo.Hostname + "  ", Style: vaxis.Style{Attribute: vaxis.AttrBold}},
		{Text: dv.sysVersion + "  ", Style: vaxis.Style{Attribute: vaxis.AttrDim}},
		{Text: dv.sysInfo.Model + "  "},
	}
	if uptimeStr != "" {
		headerSegments = append(headerSegments, vaxis.Segment{
			Text: "Up " + uptimeStr, Style: vaxis.Style{Attribute: vaxis.AttrDim},
		})
	}
	header := richtext.New(headerSegments)
	headerSurf, err := header.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, row, headerSurf)
	row++

	// === CPU gauge + sparkline ===
	cpuVal := 0.0
	cpuSuffix := ""
	if rt != nil && len(rt.CPU) > 0 {
		var totalUsage, maxTemp float64
		for _, cpu := range rt.CPU {
			totalUsage += cpu.Usage
			if cpu.Temperature > maxTemp {
				maxTemp = cpu.Temperature
			}
		}
		cpuVal = totalUsage / float64(len(rt.CPU))
		if maxTemp > 0 {
			cpuSuffix = fmt.Sprintf("%.0f°C", maxTemp)
		}
	}
	cpuGauge := &widgets.BarGauge{Label: "CPU", Value: cpuVal, Suffix: cpuSuffix, BarWidth: barWidth}
	cpuSurf, err := cpuGauge.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, row, cpuSurf)

	// Sparkline after the gauge
	if sparkCount > 0 {
		gaugeWidth := 5 + 1 + barWidth + 1 + 7 // "LABL [bars] XX.X%"
		if cpuSuffix != "" {
			gaugeWidth += len(cpuSuffix) + 2
		}
		sparkWidth := int(ctx.Max.Width) - gaugeWidth - 2
		if sparkWidth > 0 {
			dv.mu.Lock()
			sparkSurf, sparkErr := dv.cpuSpark.Draw(ctx.WithMax(vxfw.Size{Width: uint16(sparkWidth), Height: 1}))
			dv.mu.Unlock()
			if sparkErr == nil {
				s.AddChild(gaugeWidth+2, row, sparkSurf)
			}
		}
	}
	row++

	// === MEM gauge ===
	memVal := 0.0
	memSuffix := ""
	if rt != nil && rt.Memory.PhysicalTotal > 0 {
		used := rt.Memory.PhysicalTotal - rt.Memory.PhysicalAvailable
		memVal = float64(used) / float64(rt.Memory.PhysicalTotal) * 100
		memSuffix = fmt.Sprintf("%s/%s",
			humanize.IBytes(uint64(used)),
			humanize.IBytes(uint64(rt.Memory.PhysicalTotal)))
	}
	memGauge := &widgets.BarGauge{Label: "MEM", Value: memVal, Suffix: memSuffix, BarWidth: barWidth}
	memSurf, err := memGauge.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, row, memSurf)
	row++

	// === ARC gauge ===
	arcVal := 0.0
	arcSuffix := ""
	if rt != nil && rt.Memory.PhysicalTotal > 0 && rt.Memory.ArcSize > 0 {
		arcVal = float64(rt.Memory.ArcSize) / float64(rt.Memory.PhysicalTotal) * 100
		arcSuffix = humanize.IBytes(uint64(rt.Memory.ArcSize))
	}
	arcGauge := &widgets.BarGauge{Label: "ARC", Value: arcVal, Suffix: arcSuffix, BarWidth: barWidth}
	arcSurf, err := arcGauge.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, row, arcSurf)
	row++

	// === DISK gauge ===
	diskVal := 0.0
	diskSuffix := ""
	if rt != nil {
		diskVal = rt.Disks.BusyPercent
		diskSuffix = fmt.Sprintf("R:%s/s W:%s/s",
			humanize.Bytes(uint64(rt.Disks.ReadBytes)),
			humanize.Bytes(uint64(rt.Disks.WriteBytes)))
	}
	diskGauge := &widgets.BarGauge{Label: "DISK", Value: diskVal, Suffix: diskSuffix, BarWidth: barWidth}
	diskSurf, err := diskGauge.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
	if err != nil {
		return vxfw.Surface{}, err
	}
	s.AddChild(0, row, diskSurf)
	row++

	// === Blank separator ===
	row++

	// === Network section ===
	dv.mu.Lock()
	rtIfaces := make(map[string]truenas.RealtimeInterface)
	if rt != nil {
		for k, v := range rt.Interfaces {
			rtIfaces[k] = v
		}
	}
	dv.mu.Unlock()

	for _, iface := range dv.interfaces {
		if iface.State.LinkState != truenas.LinkStateUp {
			continue
		}
		rxRate := 0.0
		txRate := 0.0
		speedStr := ""
		if ri, ok := rtIfaces[iface.ID]; ok {
			rxRate = ri.ReceivedBytesRate
			txRate = ri.SentBytesRate
			if ri.Speed > 0 {
				if ri.Speed >= 1000 {
					speedStr = fmt.Sprintf("(%d Gbps)", ri.Speed/1000)
				} else {
					speedStr = fmt.Sprintf("(%d Mbps)", ri.Speed)
				}
			}
		}

		segments := []vaxis.Segment{
			{Text: fmt.Sprintf(" NET  %-12s", iface.ID), Style: vaxis.Style{Attribute: vaxis.AttrBold}},
			{Text: fmt.Sprintf("▼ %8s/s", humanize.Bytes(uint64(rxRate))), Style: vaxis.Style{Foreground: vaxis.IndexColor(2)}},
			{Text: fmt.Sprintf("  ▲ %8s/s", humanize.Bytes(uint64(txRate))), Style: vaxis.Style{Foreground: vaxis.IndexColor(3)}},
		}
		if speedStr != "" {
			segments = append(segments, vaxis.Segment{
				Text: "  " + speedStr, Style: vaxis.Style{Attribute: vaxis.AttrDim},
			})
		}
		netLabel := richtext.New(segments)
		netSurf, err := netLabel.Draw(ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: 1}))
		if err != nil {
			return vxfw.Surface{}, err
		}
		s.AddChild(0, row, netSurf)
		row++
	}

	// === Blank separator ===
	row++

	// === APPS header ===
	running := 0
	for _, a := range dv.apps {
		if a.State == "RUNNING" {
			running++
		}
	}
	appsTitle := fmt.Sprintf(" APPS (%d running / %d total)", running, len(dv.apps))
	cols := appCols(int(ctx.Max.Width))
	colHeaders := []string{appsTitle, "CPU%", "MEM", "STATE"}
	colHeaderStyles := []vaxis.Style{
		{Attribute: vaxis.AttrBold},
		{Attribute: vaxis.AttrDim},
		{Attribute: vaxis.AttrDim},
		{Attribute: vaxis.AttrDim},
	}

	colHeaderSurf := vxfw.NewSurface(ctx.Max.Width, 1, dv)
	colPos := 0
	for i, c := range cols {
		if i < len(colHeaders) {
			writeCell(&colHeaderSurf, uint16(colPos), 0, c.Width, colHeaders[i], colHeaderStyles[i], c.AlignRight)
		}
		colPos += c.Width + appColGap
	}
	s.AddChild(0, row, colHeaderSurf)
	row++

	// === App list ===
	remaining := int(ctx.Max.Height) - row
	if remaining > 0 {
		listCtx := ctx.WithMax(vxfw.Size{Width: ctx.Max.Width, Height: uint16(remaining)})
		listSurf, err := dv.appList.Draw(listCtx)
		if err != nil {
			return vxfw.Surface{}, err
		}
		s.AddChild(0, row, listSurf)
	}

	return s, nil
}

// HandleEvent delegates navigation keys to the app list.
func (dv *DashboardView) HandleEvent(ev vaxis.Event, phase vxfw.EventPhase) (vxfw.Command, error) {
	return dv.appList.HandleEvent(ev, phase)
}

// FormatUptime converts seconds to a human-readable duration.
func FormatUptime(seconds float64) string {
	d := time.Duration(seconds) * time.Second
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
