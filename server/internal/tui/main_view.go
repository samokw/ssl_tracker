package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samokw/ssl_tracker/internal/domain"
)

type MainModel struct {
	table       table.Model
	domains     []domain.Domain
	loading     bool
	err         error
	sslChecking bool
	progress    progress.Model
	sslProgress float64
	width       int
	height      int
}

func NewMainModel() MainModel {
	columns := []table.Column{
		{Title: "Domain", Width: 25},
		{Title: "Status", Width: 12},
		{Title: "Expires", Width: 15},
		{Title: "Last Check", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	prog := progress.New(progress.WithDefaultGradient())
	prog.ShowPercentage = true
	prog.Width = 60

	return MainModel{
		table:       t,
		domains:     []domain.Domain{},
		loading:     true,
		sslChecking: false,
		progress:    prog,
		sslProgress: 0.0,
		width:       80,
		height:      24,
	}
}

func (m MainModel) Update(msg tea.Msg) (MainModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(m.domains) > 0 && m.table.Cursor() < len(m.domains) {
				selectedDomain := m.domains[m.table.Cursor()]
				return m, func() tea.Msg {
					return CheckSingleDomainMsg{domainID: selectedDomain.DomainID}
				}
			}
		case "a":
			return m, func() tea.Msg { return "show_add_domain" }
		case "d":
			if len(m.domains) > 0 && m.table.Cursor() < len(m.domains) {
				selectedDomain := m.domains[m.table.Cursor()]
				return m, func() tea.Msg {
					return DeleteDomainMsg{domainID: selectedDomain.DomainID}
				}
			}
		case "r":
			return m, func() tea.Msg { return "refresh_domains" }
		}
	}

	// Update table
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m MainModel) View() string {
	var b strings.Builder

	// Calculate responsive layout dimensions
	separatorWidth := m.width - 4 // Leave 2 chars padding on each side
	if separatorWidth < 20 {
		separatorWidth = 20 // Minimum separator width
	}
	if separatorWidth > 80 {
		separatorWidth = 80 // Maximum separator width
	}

	b.WriteString("\n\n")

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ff88")).
		Bold(true).
		Width(m.width).
		Align(lipgloss.Center)

	b.WriteString(headerStyle.Render("sslcerttop 🔒 SSL Certificate Monitor"))
	b.WriteString("\n")

	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc")).
		Width(m.width).
		Align(lipgloss.Center)

	domainCount := len(m.domains)
	b.WriteString(statsStyle.Render(fmt.Sprintf("[%d domains tracked]", domainCount)))
	b.WriteString("\n")

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Width(m.width).
		Align(lipgloss.Center)

	if m.width < 84 {
		b.WriteString(separatorStyle.Render("- - - - - - - - - - - - - - - -"))
	} else {
		b.WriteString(separatorStyle.Render(strings.Repeat("═", separatorWidth)))
	}
	b.WriteString("\n\n")

	if m.sslChecking {
		statusStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center)
		b.WriteString(statusStyle.Render("🔍 Checking SSL certificates..."))
		b.WriteString("\n\n")

		progressStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center)
		b.WriteString(progressStyle.Render(m.progress.ViewAs(m.sslProgress)))
		b.WriteString("\n\n")
	} else if m.loading {
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00bfff")).
			Width(m.width).
			Align(lipgloss.Center)
		b.WriteString(loadingStyle.Render("Loading domains..."))
		b.WriteString("\n")
		b.WriteString(loadingStyle.Render("⣾⣽⣻⢿⡿⣟⣯⣟"))
		b.WriteString("\n")
	} else if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff4444")).
			Bold(true).
			Width(m.width).
			Align(lipgloss.Center)
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	} else if len(m.domains) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cccccc")).
			Width(m.width).
			Align(lipgloss.Center)
		b.WriteString(emptyStyle.Render("No domains found. Press 'a' to add your first domain."))
		b.WriteString("\n")
	} else {
		listHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00bfff")).
			Bold(true).
			Width(m.width).
			Align(lipgloss.Center)
		b.WriteString(listHeaderStyle.Render("📋 Your SSL Certificates"))
		b.WriteString("\n\n")

		tableStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center)
		b.WriteString(tableStyle.Render(m.table.View()))
	}

	b.WriteString("\n\n")

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Width(m.width).
		Align(lipgloss.Center)

	footerText := "[Enter] Check SSL  [a] Add Domain  [d] Delete  [r] Refresh  [Alt+Enter] Toggle Screen  [q] Quit"
	if m.width < 80 {
		footerText = "[Enter] Check  [a] Add  [d] Del  [r] Refresh  [q] Quit"
	}
	b.WriteString(footerStyle.Render(footerText))

	return b.String()
}

// UpdateSize adjusts the model for new terminal dimensions
func (m *MainModel) UpdateSize(width, height int) {
	m.width = width
	m.height = height

	var columns []table.Column
	if width < 80 {
		columns = []table.Column{
			{Title: "Domain", Width: max(20, width/3)},
			{Title: "Status", Width: 8},
			{Title: "Expires", Width: 8},
		}
	} else if width < 120 {
		columns = []table.Column{
			{Title: "Domain", Width: 25},
			{Title: "Status", Width: 12},
			{Title: "Expires", Width: 15},
			{Title: "Last Check", Width: 12},
		}
	} else {
		columns = []table.Column{
			{Title: "Domain", Width: 35},
			{Title: "Status", Width: 15},
			{Title: "Expires", Width: 20},
			{Title: "Last Check", Width: 18},
			{Title: "Details", Width: 25},
		}
	}

	m.table.SetRows([]table.Row{})
	m.table.SetColumns(columns)

	if len(m.domains) > 0 {
		m.SetDomains(m.domains)
	}

	tableHeight := max(5, height-10)
	m.table.SetHeight(tableHeight)

	progressWidth := max(30, min(60, width-10))
	m.progress.Width = progressWidth
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to update table data
func (m *MainModel) SetDomains(domains []domain.Domain) {
	m.domains = domains
	m.loading = false

	// Convert domains to table rows based on current column layout
	rows := make([]table.Row, len(domains))
	columns := m.table.Columns()

	for i, d := range domains {
		status := m.getStatusDisplay(d)
		expires := m.getExpiryDisplay(d)
		lastCheck := m.getLastCheckDisplay(d)

		switch len(columns) {
		case 3: // Narrow layout
			rows[i] = table.Row{
				d.DomainName.String(),
				status,
				expires,
			}
		case 4: // Standard layout
			rows[i] = table.Row{
				d.DomainName.String(),
				status,
				expires,
				lastCheck,
			}
		case 5: // Wide layout
			details := m.getDetailsDisplay(d)
			rows[i] = table.Row{
				d.DomainName.String(),
				status,
				expires,
				lastCheck,
				details,
			}
		default: // Fallback to standard
			rows[i] = table.Row{
				d.DomainName.String(),
				status,
				expires,
				lastCheck,
			}
		}
	}

	m.table.SetRows(rows)
}

func (m MainModel) getStatusDisplay(d domain.Domain) string {
	if d.LastError != nil {
		return "❌ Error"
	}

	if d.ExpiryDate == nil {
		return "❓ Unknown"
	}

	daysLeft := time.Until(d.ExpiryDate.Time()).Hours() / 24

	if daysLeft < 0 {
		return "❌ Expired"
	} else if daysLeft < 7 {
		return "⚠️ Warning"
	} else if daysLeft < 30 {
		return "🟡 Soon"
	} else {
		return "✅ Valid"
	}
}

func (m MainModel) getExpiryDisplay(d domain.Domain) string {
	if d.ExpiryDate == nil {
		return "Unknown"
	}

	daysLeft := time.Until(d.ExpiryDate.Time()).Hours() / 24

	if daysLeft < 0 {
		return fmt.Sprintf("-%d days", int(-daysLeft))
	} else {
		return fmt.Sprintf("%d days", int(daysLeft))
	}
}

func (m MainModel) getLastCheckDisplay(d domain.Domain) string {
	if d.LastChecked == nil {
		return "Never"
	}

	duration := time.Since(d.LastChecked.Time())

	if duration.Hours() < 1 {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	} else if duration.Hours() < 24 {
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd ago", int(duration.Hours()/24))
	}
}

func (m MainModel) getDetailsDisplay(d domain.Domain) string {
	if d.LastError != nil {
		return "Check failed"
	}

	if d.ExpiryDate == nil {
		return "No cert data"
	}

	daysLeft := time.Until(d.ExpiryDate.Time()).Hours() / 24

	if daysLeft < 0 {
		return "Certificate expired"
	} else if daysLeft < 7 {
		return "Expires very soon!"
	} else if daysLeft < 30 {
		return "Renewal recommended"
	} else {
		return "Certificate healthy"
	}
}
