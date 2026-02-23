//go:build cgo

package app

import (
	"fmt"
	"image/color"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"port_sentinel/internal/ports"
	"port_sentinel/internal/store"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func Run() error {
	cfg, _ := store.LoadConfig()
	state := NewState(cfg)
	svc := NewDefaultService(state)

	fyneApp := app.NewWithID("portsentinel")
	w := fyneApp.NewWindow("Port Sentinel")
	w.Resize(fyne.NewSize(980, 600))

	status := widget.NewLabel("Ready.")
	status.Wrapping = fyne.TextWrapWord

	refresher := &AutoRefresher{}

	refreshAll := func() {
		go func() {
			_, err := svc.RefreshAll()
			fyne.Do(func() {
				if err != nil {
					status.SetText(fmt.Sprintf("Refresh failed: %v", err))
				} else {
					status.SetText("Refreshed.")
				}
			})
		}()
	}

	rowHeader := container.NewGridWithColumns(8,
		widget.NewLabelWithStyle("Port", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Pinned", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("PID", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Process", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Command", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Updated", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	var list *widget.List
	list = widget.NewList(
		func() int {
			return len(state.GetPorts())
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
			port := widget.NewLabel("")
			pin := widget.NewCheck("", nil)
			status := widget.NewLabel("")
			pid := widget.NewLabel("")
			proc := widget.NewLabel("")
			cmd := widget.NewLabel("")
			updated := widget.NewLabel("")
			refreshBtn := widget.NewButton("Refresh", nil)
			killBtn := widget.NewButton("Terminate", nil)
			actions := container.NewHBox(refreshBtn, killBtn)
			grid := container.NewGridWithColumns(8, port, pin, status, pid, proc, cmd, updated, actions)
			return container.NewMax(bg, grid)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container)
			bg := row.Objects[0].(*canvas.Rectangle)
			grid := row.Objects[1].(*fyne.Container)
			portsList := state.GetPorts()
			if i >= len(portsList) {
				return
			}
			port := portsList[i]
			result := getResult(state, port)

			portLabel := grid.Objects[0].(*widget.Label)
			pinCheck := grid.Objects[1].(*widget.Check)
			statusLabel := grid.Objects[2].(*widget.Label)
			pidLabel := grid.Objects[3].(*widget.Label)
			procLabel := grid.Objects[4].(*widget.Label)
			cmdLabel := grid.Objects[5].(*widget.Label)
			updatedLabel := grid.Objects[6].(*widget.Label)
			actions := grid.Objects[7].(*fyne.Container)
			refreshBtn := actions.Objects[0].(*widget.Button)
			killBtn := actions.Objects[1].(*widget.Button)

			portLabel.SetText(strconv.Itoa(port))
			pinned := state.IsPinned(port)
			pinCheck.OnChanged = nil
			pinCheck.SetChecked(pinned)
			pinCheck.OnChanged = func(val bool) {
				if err := svc.TogglePinAndSave(port, val); err != nil {
					status.SetText(fmt.Sprintf("Pin update failed: %v", err))
					return
				}
				if canvas := fyneApp.Driver().CanvasForObject(pinCheck); canvas != nil {
					canvas.Focus(nil)
				}
				list.UnselectAll()
				list.Refresh()
			}
			if pinned {
				bg.FillColor = color.NRGBA{R: 220, G: 235, B: 250, A: 255}
			} else {
				bg.FillColor = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
			}
			bg.Refresh()
			statusLabel.SetText(string(result.Status))
			if result.PID > 0 {
				pidLabel.SetText(strconv.Itoa(result.PID))
			} else {
				pidLabel.SetText("-")
			}
			procLabel.SetText(ellipsis(result.ProcessName, 18))
			cmdLabel.SetText(ellipsis(maskSensitiveArgs(firstNonEmpty(result.CommandLine, result.ExePath)), 32))
			if !result.UpdatedAt.IsZero() {
				updatedLabel.SetText(result.UpdatedAt.Local().Format("15:04:05"))
			} else {
				updatedLabel.SetText("-")
			}

			refreshBtn.OnTapped = func() {
				go func() {
					_, err := svc.RefreshOne(port)
					fyne.Do(func() {
						if err != nil {
							status.SetText(fmt.Sprintf("Refresh %d failed: %v", port, err))
						} else {
							status.SetText(fmt.Sprintf("Port %d refreshed.", port))
						}
						list.Refresh()
					})
				}()
			}

			killBtn.Disable()
			if result.Status == ports.StatusInUse && result.PID > 0 {
				killBtn.Enable()
				killBtn.OnTapped = func() {
					showKillDialog(fyneApp, w, svc, state, result, status, list)
				}
			}
		},
	)
	for i := 0; i < len(state.GetPorts()); i++ {
		list.SetItemHeight(widget.ListItemID(i), 36)
	}

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("Port (1-65535)")
	minSize := portEntry.MinSize()
	portEntryWrap := container.NewGridWrap(fyne.NewSize(180, minSize.Height), portEntry)
	addBtn := widget.NewButton("Add", func() {
		port, err := strconv.Atoi(strings.TrimSpace(portEntry.Text))
		if err != nil {
			status.SetText("Invalid port.")
			return
		}
		if err := svc.AddCustomPortAndSave(port); err != nil {
			status.SetText(fmt.Sprintf("Add failed: %v", err))
			return
		}
		portEntry.SetText("")
		list.Refresh()
		status.SetText(fmt.Sprintf("Port %d added.", port))
	})

	refreshAllBtn := widget.NewButtonWithIcon("Refresh All", theme.ViewRefreshIcon(), func() {
		refreshAll()
		list.Refresh()
	})

	intervalSelect := widget.NewSelect([]string{"2s", "5s", "10s"}, nil)
	intervalSelect.SetSelected(fmt.Sprintf("%ds", cfg.UI.AutoRefreshIntervalMs/1000))

	autoRefresh := widget.NewCheck("Auto Refresh", func(checked bool) {
		if err := svc.UpdateUIConfig(func(c *store.Config) error {
			c.UI.AutoRefreshEnabled = checked
			return nil
		}); err != nil {
			status.SetText(fmt.Sprintf("Auto refresh update failed: %v", err))
			return
		}
		applyAutoRefresh(fyneApp, refresher, svc, intervalSelect.Selected, checked, list, status)
	})
	autoRefresh.SetChecked(cfg.UI.AutoRefreshEnabled)

	intervalSelect.OnChanged = func(value string) {
		if err := svc.UpdateUIConfig(func(c *store.Config) error {
			c.UI.AutoRefreshIntervalMs = parseIntervalMs(value)
			return nil
		}); err != nil {
			status.SetText(fmt.Sprintf("Interval update failed: %v", err))
			return
		}
		applyAutoRefresh(fyneApp, refresher, svc, value, autoRefresh.Checked, list, status)
	}

	settingsBtn := widget.NewButton("Ports & Settings", func() {
		showSettingsDialog(fyneApp, w, svc, state, list, status)
	})

	top := container.NewHBox(portEntryWrap, addBtn, refreshAllBtn, autoRefresh, intervalSelect, settingsBtn)
	content := container.NewBorder(top, status, nil, nil, container.NewBorder(rowHeader, nil, nil, nil, list))
	w.SetContent(content)

	applyAutoRefresh(fyneApp, refresher, svc, intervalSelect.Selected, autoRefresh.Checked, list, status)
	w.SetOnClosed(func() {
		refresher.Stop()
	})

	status.SetText("Loading ports...")
	if _, err := svc.RefreshAll(); err != nil {
		status.SetText(fmt.Sprintf("Initial refresh failed: %v", err))
	} else {
		status.SetText("Ready.")
	}
	list.Refresh()
	w.ShowAndRun()
	return nil
}

func applyAutoRefresh(app fyne.App, refresher *AutoRefresher, svc *Service, interval string, enabled bool, list *widget.List, status *widget.Label) {
	intervalMs := parseIntervalMs(interval)
	if enabled {
		refresher.Start(time.Duration(intervalMs)*time.Millisecond, func() {
			_, err := svc.RefreshAll()
			fyne.Do(func() {
				if err != nil {
					status.SetText(fmt.Sprintf("Auto refresh failed: %v", err))
				}
				list.Refresh()
			})
		})
	} else {
		refresher.Stop()
	}
}

func showSettingsDialog(app fyne.App, w fyne.Window, svc *Service, state *State, list *widget.List, status *widget.Label) {
	cfg := state.SnapshotConfig()
	presetBox := container.NewVBox()
	presetPorts := make([]int, 0, len(cfg.PresetPorts))
	for port := range cfg.PresetPorts {
		presetPorts = append(presetPorts, port)
	}
	sort.Ints(presetPorts)
	for _, port := range presetPorts {
		enabled := cfg.PresetPorts[port]
		p := port
		checked := enabled
		check := widget.NewCheck(fmt.Sprintf("%d", port), func(val bool) {
			if err := svc.TogglePresetAndSave(p, val); err != nil {
				status.SetText(fmt.Sprintf("Preset update failed: %v", err))
				return
			}
			list.Refresh()
		})
		check.SetChecked(checked)
		presetBox.Add(check)
	}

	customList := container.NewVBox()
	for _, port := range cfg.CustomPorts {
		p := port
		row := container.NewHBox(widget.NewLabel(strconv.Itoa(port)), widget.NewButton("Remove", func() {
			if err := svc.RemoveCustomPortAndSave(p); err != nil {
				status.SetText(fmt.Sprintf("Remove failed: %v", err))
				return
			}
			list.Refresh()
			status.SetText(fmt.Sprintf("Port %d removed.", p))
		}))
		customList.Add(row)
	}

	forceKill := widget.NewCheck("Force terminate by default", func(val bool) {
		if err := svc.UpdateUIConfig(func(c *store.Config) error {
			c.UI.ForceKillEnabled = val
			return nil
		}); err != nil {
			status.SetText(fmt.Sprintf("Force terminate update failed: %v", err))
		}
	})
	forceKill.SetChecked(cfg.UI.ForceKillEnabled)

	content := container.NewVBox(
		widget.NewLabelWithStyle("Preset Ports", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		presetBox,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Custom Ports", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		customList,
		widget.NewSeparator(),
		forceKill,
	)

	dialog.NewCustom("Ports & Settings", "Close", content, w).Show()
}

func showKillDialog(app fyne.App, w fyne.Window, svc *Service, state *State, result ports.PortScanResult, status *widget.Label, list *widget.List) {
	force := widget.NewCheck("Force terminate", nil)
	cfg := state.SnapshotConfig()
	force.SetChecked(cfg.UI.ForceKillEnabled)
	ack := widget.NewCheck("I understand this may terminate critical system/app processes", nil)

	message := fmt.Sprintf("Terminate PID %d (%s) on port %d?", result.PID, result.ProcessName, result.Port)
	exePath := firstNonEmpty(strings.TrimSpace(result.ExePath), "-")
	cmdPreview := ellipsis(maskSensitiveArgs(strings.TrimSpace(result.CommandLine)), 96)
	if cmdPreview == "" {
		cmdPreview = "-"
	}
	content := container.NewVBox(
		widget.NewLabel(message),
		widget.NewLabel("Executable: "+exePath),
		widget.NewLabel("Command: "+cmdPreview),
		force,
		ack,
	)
	dialog.NewCustomConfirm("Terminate Process", "Terminate", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}
		if !ack.Checked {
			status.SetText("Please acknowledge risk before terminating the process.")
			return
		}
		go func() {
			err := svc.KillProcess(result.PID, force.Checked)
			fyne.Do(func() {
				if err != nil {
					status.SetText(fmt.Sprintf("Terminate failed: %v", err))
				} else {
					status.SetText(fmt.Sprintf("Terminated PID %d.", result.PID))
				}
				_, _ = svc.RefreshAll()
				list.Refresh()
			})
		}()
	}, w).Show()
}

func getResult(state *State, port int) ports.PortScanResult {
	state.mu.Lock()
	defer state.mu.Unlock()
	if res, ok := state.Results[port]; ok {
		return res
	}
	return ports.PortScanResult{
		Port:      port,
		Status:    ports.StatusUnknown,
		Protocol:  ports.ProtocolTCP,
		UpdatedAt: ports.NowStamp(),
	}
}

func firstNonEmpty(values ...string) string {
	for _, val := range values {
		if strings.TrimSpace(val) != "" {
			return val
		}
	}
	return ""
}

func ellipsis(val string, max int) string {
	val = strings.TrimSpace(val)
	if max <= 0 {
		return ""
	}
	if len(val) <= max {
		return val
	}
	return val[:max-3] + "..."
}

func parseIntervalMs(s string) int {
	switch s {
	case "2s":
		return 2000
	case "10s":
		return 10000
	default:
		return 5000
	}
}

var sensitiveArgPattern = regexp.MustCompile(`(?i)\b(password|passwd|pwd|token|secret|api[_-]?key)\s*[:=]\s*([^\s]+)`)
var bearerTokenPattern = regexp.MustCompile(`(?i)\bbearer\s+([A-Za-z0-9\-._~+/]+=*)`)

func maskSensitiveArgs(input string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}
	masked := sensitiveArgPattern.ReplaceAllString(input, "$1=***")
	masked = bearerTokenPattern.ReplaceAllString(masked, "bearer ***")
	return masked
}
