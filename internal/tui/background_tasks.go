package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TaskStatus represents the status of a background task
type TaskStatus int

const (
	TaskRunning TaskStatus = iota
	TaskCompleted
	TaskFailed
	TaskCancelled
)

func (s TaskStatus) String() string {
	switch s {
	case TaskRunning:
		return "Running"
	case TaskCompleted:
		return "Completed"
	case TaskFailed:
		return "Failed"
	case TaskCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// BackgroundTask represents a single background task
type BackgroundTask struct {
	ID          string
	Name        string
	Description string
	Status      TaskStatus
	StartTime   time.Time
	EndTime     *time.Time
	Progress    float64 // 0.0 to 1.0
	Output      string
	Error       error
	PID         int // Process ID if applicable
}

// Duration returns the task duration
func (t *BackgroundTask) Duration() time.Duration {
	if t.EndTime != nil {
		return t.EndTime.Sub(t.StartTime)
	}
	return time.Since(t.StartTime)
}

// IsActive returns true if the task is still running
func (t *BackgroundTask) IsActive() bool {
	return t.Status == TaskRunning
}

// BackgroundTaskPanel manages the display of background tasks
type BackgroundTaskPanel struct {
	tasks   []*BackgroundTask
	visible bool
	cursor  int
	width   int
	height  int
}

// NewBackgroundTaskPanel creates a new background task panel
func NewBackgroundTaskPanel() *BackgroundTaskPanel {
	return &BackgroundTaskPanel{
		tasks:   []*BackgroundTask{},
		visible: false,
		cursor:  0,
		width:   60,
		height:  20,
	}
}

// AddTask adds a new background task
func (p *BackgroundTaskPanel) AddTask(task *BackgroundTask) {
	p.tasks = append(p.tasks, task)
}

// RemoveTask removes a task by ID
func (p *BackgroundTaskPanel) RemoveTask(id string) {
	filtered := []*BackgroundTask{}
	for _, task := range p.tasks {
		if task.ID != id {
			filtered = append(filtered, task)
		}
	}
	p.tasks = filtered
}

// UpdateTask updates an existing task
func (p *BackgroundTaskPanel) UpdateTask(id string, updater func(*BackgroundTask)) {
	for _, task := range p.tasks {
		if task.ID == id {
			updater(task)
			break
		}
	}
}

// GetTask retrieves a task by ID
func (p *BackgroundTaskPanel) GetTask(id string) *BackgroundTask {
	for _, task := range p.tasks {
		if task.ID == id {
			return task
		}
	}
	return nil
}

// GetRunningTasks returns all running tasks
func (p *BackgroundTaskPanel) GetRunningTasks() []*BackgroundTask {
	var running []*BackgroundTask
	for _, task := range p.tasks {
		if task.Status == TaskRunning {
			running = append(running, task)
		}
	}
	return running
}

// GetCompletedTasks returns all completed tasks
func (p *BackgroundTaskPanel) GetCompletedTasks() []*BackgroundTask {
	var completed []*BackgroundTask
	for _, task := range p.tasks {
		if task.Status == TaskCompleted {
			completed = append(completed, task)
		}
	}
	return completed
}

// GetFailedTasks returns all failed tasks
func (p *BackgroundTaskPanel) GetFailedTasks() []*BackgroundTask {
	var failed []*BackgroundTask
	for _, task := range p.tasks {
		if task.Status == TaskFailed {
			failed = append(failed, task)
		}
	}
	return failed
}

// HasActiveTasks returns true if there are any running tasks
func (p *BackgroundTaskPanel) HasActiveTasks() bool {
	return len(p.GetRunningTasks()) > 0
}

// TaskCount returns the total number of tasks
func (p *BackgroundTaskPanel) TaskCount() int {
	return len(p.tasks)
}

// Toggle toggles the panel visibility
func (p *BackgroundTaskPanel) Toggle() {
	p.visible = !p.visible
}

// Show shows the panel
func (p *BackgroundTaskPanel) Show() {
	p.visible = true
}

// Hide hides the panel
func (p *BackgroundTaskPanel) Hide() {
	p.visible = false
}

// IsVisible returns true if the panel is visible
func (p *BackgroundTaskPanel) IsVisible() bool {
	return p.visible
}

// MoveCursor moves the cursor up or down
func (p *BackgroundTaskPanel) MoveCursor(delta int) {
	p.cursor += delta
	if p.cursor < 0 {
		p.cursor = 0
	}
	if p.cursor >= len(p.tasks) {
		p.cursor = len(p.tasks) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

// GetSelectedTask returns the currently selected task
func (p *BackgroundTaskPanel) GetSelectedTask() *BackgroundTask {
	if p.cursor >= 0 && p.cursor < len(p.tasks) {
		return p.tasks[p.cursor]
	}
	return nil
}

// ClearCompleted removes all completed tasks
func (p *BackgroundTaskPanel) ClearCompleted() {
	filtered := []*BackgroundTask{}
	for _, task := range p.tasks {
		if task.Status != TaskCompleted {
			filtered = append(filtered, task)
		}
	}
	p.tasks = filtered
	if p.cursor >= len(p.tasks) {
		p.cursor = len(p.tasks) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

// CleanupOldTasks removes old completed tasks, keeping only the last N
func (p *BackgroundTaskPanel) CleanupOldTasks(keepLast int) {
	completed := []*BackgroundTask{}
	active := []*BackgroundTask{}

	for _, task := range p.tasks {
		if task.Status == TaskCompleted {
			completed = append(completed, task)
		} else {
			active = append(active, task)
		}
	}

	// Keep only last N completed tasks
	if len(completed) > keepLast {
		completed = completed[len(completed)-keepLast:]
	}

	// Rebuild task list: active tasks + recent completed
	p.tasks = append(active, completed...)

	if p.cursor >= len(p.tasks) {
		p.cursor = len(p.tasks) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

// View renders the background task panel
func (p *BackgroundTaskPanel) View(width, height int) string {
	if !p.visible {
		return ""
	}

	p.width = min(width-4, 80)
	p.height = min(height-4, 30)

	var b strings.Builder

	// Header
	header := taskHeaderStyle.Render("⚙️  Background Tasks")
	b.WriteString(header + "\n")
	b.WriteString(strings.Repeat("─", p.width) + "\n\n")

	// Running tasks
	running := p.GetRunningTasks()
	if len(running) > 0 {
		b.WriteString(taskSectionStyle.Render(fmt.Sprintf("⚡ Running (%d)", len(running))) + "\n")
		for _, task := range running {
			b.WriteString(p.renderTask(task, false) + "\n")
		}
		b.WriteString("\n")
	}

	// Completed tasks (last 5)
	completed := p.GetCompletedTasks()
	if len(completed) > 0 {
		displayCount := min(len(completed), 5)
		b.WriteString(taskSectionStyle.Render(fmt.Sprintf("✅ Completed (%d)", len(completed))) + "\n")
		for i := len(completed) - displayCount; i < len(completed); i++ {
			b.WriteString(p.renderTask(completed[i], false) + "\n")
		}
		b.WriteString("\n")
	}

	// Failed tasks
	failed := p.GetFailedTasks()
	if len(failed) > 0 {
		b.WriteString(taskSectionStyle.Render(fmt.Sprintf("❌ Failed (%d)", len(failed))) + "\n")
		for _, task := range failed {
			b.WriteString(p.renderTask(task, false) + "\n")
		}
		b.WriteString("\n")
	}

	// Empty state
	if len(p.tasks) == 0 {
		b.WriteString(statusStyle.Render("  No background tasks\n\n"))
	}

	// Footer with controls
	controls := statusStyle.Render("[Ctrl+B] Close  [c] Clear completed  [↑↓] Navigate  [Enter] Details")
	b.WriteString("\n" + controls)

	// Wrap in a box
	content := b.String()
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("33")).
		Padding(1, 2).
		Width(p.width)

	box := boxStyle.Render(content)

	// Center on screen
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// renderTask renders a single task
func (p *BackgroundTaskPanel) renderTask(task *BackgroundTask, detailed bool) string {
	var b strings.Builder

	// Status icon
	icon := "⏳"
	iconStyle := statusStyle
	switch task.Status {
	case TaskRunning:
		icon = "🔄"
		iconStyle = toolRunningStyle
	case TaskCompleted:
		icon = "✅"
		iconStyle = toolDoneStyle
	case TaskFailed:
		icon = "❌"
		iconStyle = toolFailedStyle
	case TaskCancelled:
		icon = "⏹️"
		iconStyle = statusStyle
	}

	// Task name and duration
	duration := formatDuration(task.Duration())
	line := fmt.Sprintf("  %s %s", iconStyle.Render(icon), task.Name)

	if task.Status == TaskRunning {
		line += fmt.Sprintf(" (%s)", duration)
	} else if task.EndTime != nil {
		ago := time.Since(*task.EndTime)
		line += fmt.Sprintf(" (%s, %s ago)", duration, formatDuration(ago))
	}

	b.WriteString(line + "\n")

	// Progress bar for running tasks
	if task.Status == TaskRunning && task.Progress > 0 {
		progressBar := renderProgressBar(task.Progress, 30)
		b.WriteString("    " + progressBar + "\n")
	}

	// Description
	if task.Description != "" && detailed {
		b.WriteString("    " + statusStyle.Render(task.Description) + "\n")
	}

	// Error message for failed tasks
	if task.Status == TaskFailed && task.Error != nil {
		errMsg := truncate(task.Error.Error(), 60)
		b.WriteString("    " + errorStyle.Render("Error: "+errMsg) + "\n")
	}

	return b.String()
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// Task panel styles
var (
	taskHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true)

	taskSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250")).
				Bold(true)
)

// renderProgressBar creates a simple text-based progress bar
func renderProgressBar(progress float64, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	filled := int((progress / 100.0) * float64(width))
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("[%s] %.0f%%", bar, progress)
}
