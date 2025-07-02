package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samokw/ssl_tracker/internal/domain"
	"github.com/samokw/ssl_tracker/internal/types"
)

type App struct {
	domainService *domain.Service
	currentView   View
	home          HomeModel
	main          MainModel
	domain        DomainModel
	altScreen     bool
	width         int
	height        int
}

type View int

const (
	Home View = iota
	Main
	AddDomain
)

func NewApp(domainService *domain.Service) *App {
	return &App{
		domainService: domainService,
		currentView:   Home,
		home:          NewHomeModel(),
		main:          NewMainModel(),
		domain:        NewDomainModel(),
		altScreen:     true,
	}
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update app dimensions and propagate to views
		a.width = msg.Width
		a.height = msg.Height
		a.home.UpdateSize(msg.Width, msg.Height)
		a.main.UpdateSize(msg.Width, msg.Height)
		a.domain.UpdateSize(msg.Width, msg.Height)
		return a, nil
	case DomainsLoadedMsg:
		if msg.err != nil {
			a.main.err = msg.err
			a.main.loading = false
		} else {
			a.main.SetDomains(msg.domains)
		}
		return a, nil
	case SSLCheckStartedMsg:
		// Start SSL checking progress
		a.main.sslChecking = true
		a.main.sslProgress = 0.0
		return a, nil
	case SSLCheckCompletedMsg:
		// SSL check completed, stop progress and reload domains
		a.main.sslChecking = false
		a.main.sslProgress = 1.0
		return a, a.loadDomains()
	case SSLProgressMsg:
		// Update progress with real data
		a.main.sslProgress = msg.progress
		if msg.progress >= 1.0 {
			// Complete - transition to completion
			a.main.sslChecking = false
			return a, a.loadDomains()
		}
		return a, nil
	case ProgressTickMsg:
		// Update progress incrementally while SSL checking
		if a.main.sslChecking && a.main.sslProgress < 0.95 {
			a.main.sslProgress += 0.05
			return a, a.progressTicker()
		}
		return a, nil
	case AddDomainMsg:
		// Add a new domain
		return a, a.addDomain(msg.domain)
	case DomainAddedMsg:
		// Domain addition completed, delegate to domain view
		if a.currentView == AddDomain {
			var cmd tea.Cmd
			a.domain, cmd = a.domain.Update(msg)
			return a, cmd
		}
		return a, nil
	case DeleteDomainMsg:
		// Delete a domain
		return a, a.deleteDomain(msg.domainID)
	case DomainDeletedMsg:
		// Domain deletion completed, reload domains
		if msg.err != nil {
			a.main.err = msg.err
		}
		return a, a.loadDomains()
	case CheckSingleDomainMsg:
		// Check SSL for a single domain
		return a, a.checkSingleDomain(msg.domainID)
	case SingleDomainCheckCompletedMsg:
		// Single domain SSL check completed, reload domains
		if msg.err != nil {
			a.main.err = msg.err
		}
		return a, a.loadDomains()
	case string:
		switch msg {
		case "refresh_domains":
			// Trigger SSL check for all domains
			return a, a.checkAllSSL()
		case "show_add_domain":
			// Switch to add domain view
			a.currentView = AddDomain
			a.domain = NewDomainModel()            // Reset the form
			a.domain.UpdateSize(a.width, a.height) // Apply current window size
			return a, nil
		case "back_to_main":
			// Switch back to main view and reload domains
			a.currentView = Main
			return a, a.loadDomains()
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "alt+enter", "f11":
			// Toggle alt screen mode
			a.altScreen = !a.altScreen
			if a.altScreen {
				return a, tea.EnterAltScreen
			} else {
				return a, tea.ExitAltScreen
			}
		default:
			// If we're on home screen, any key moves to main
			if a.currentView == Home {
				a.currentView = Main
				// Load domains when transitioning to main view
				return a, a.loadDomains()
			} else if a.currentView == Main {
				// Delegate to main view and handle special commands
				var cmd tea.Cmd
				a.main, cmd = a.main.Update(msg)
				return a, cmd
			} else if a.currentView == AddDomain {
				// Delegate to add domain view
				var cmd tea.Cmd
				a.domain, cmd = a.domain.Update(msg)
				return a, cmd
			}
		}
	}

	// Delegate to current view
	switch a.currentView {
	case Home:
		// Handle home view updates
		return a, nil
	case Main:
		// Handle main view updates
		return a, nil
	case AddDomain:
		// Handle add domain view updates
		return a, nil
	}

	return a, nil
}

// View renders the current view
func (a *App) View() string {
	switch a.currentView {
	case Home:
		return a.renderHomeView()
	case Main:
		return a.renderMainView()
	case AddDomain:
		return a.renderAddDomainView()
	default:
		return "Unknown view"
	}
}

// Placeholder render methods - we'll implement these next
func (a *App) renderHomeView() string {
	return a.home.RenderSplash()
}

func (a *App) renderMainView() string {
	return a.main.View()
}

func (a *App) renderAddDomainView() string {
	return a.domain.View()
}

// loadDomains loads domains from the service
func (a *App) loadDomains() tea.Cmd {
	return func() tea.Msg {
		domains, err := a.domainService.GetUsersDomains(types.UserID(1)) // Use default user
		if err != nil {
			return DomainsLoadedMsg{err: err}
		}
		return DomainsLoadedMsg{domains: domains}
	}
}

// checkAllSSL performs SSL checks on all domains with progress reporting
func (a *App) checkAllSSL() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg { return SSLCheckStartedMsg{} },
		a.progressTicker(),
		a.checkDomainsWithProgress(),
	)
}

// progressTicker creates a ticker to simulate progress updates
func (a *App) progressTicker() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return ProgressTickMsg{}
	})
}

// checkDomainsWithProgress checks domains concurrently using the worker pool
func (a *App) checkDomainsWithProgress() tea.Cmd {
	return func() tea.Msg {
		// Use the synchronous version that waits for completion
		err := a.domainService.CheckAllDomainsSSLSync(types.UserID(1))
		return SSLCheckCompletedMsg{err: err}
	}
}

// addDomain adds a new domain to the system
func (a *App) addDomain(domainName string) tea.Cmd {
	return func() tea.Msg {
		_, err := a.domainService.AddDomain(types.UserID(1), domainName)
		if err != nil {
			return DomainAddedMsg{err: err}
		}

		// Also perform an initial SSL check
		domains, err := a.domainService.GetUsersDomains(types.UserID(1))
		if err == nil {
			for _, d := range domains {
				if d.DomainName.String() == domainName {
					_ = a.domainService.CheckDomainSSL(d.DomainID)
					break
				}
			}
		}

		return DomainAddedMsg{err: nil}
	}
}

// deleteDomain removes a domain from the system
func (a *App) deleteDomain(domainID types.DomainID) tea.Cmd {
	return func() tea.Msg {
		err := a.domainService.RemoveDomain(domainID)
		return DomainDeletedMsg{err: err}
	}
}

// checkSingleDomain checks SSL for a single domain
func (a *App) checkSingleDomain(domainID types.DomainID) tea.Cmd {
	return func() tea.Msg {
		err := a.domainService.CheckDomainSSL(domainID)
		return SingleDomainCheckCompletedMsg{domainID: domainID, err: err}
	}
}

// DomainsLoadedMsg represents the result of loading domains
type DomainsLoadedMsg struct {
	domains []domain.Domain
	err     error
}

// Add SSL checking message types
type SSLCheckStartedMsg struct{}

type SSLCheckCompletedMsg struct {
	err error
}

// Progress message types
type SSLProgressMsg struct {
	progress     float64
	domainName   string
	totalDomains int
	completed    int
}

type ProgressTickMsg struct{}

// Domain management message types (defined in add_domain.go)
type DeleteDomainMsg struct {
	domainID types.DomainID
}

type DomainDeletedMsg struct {
	err error
}

// Single domain SSL check message types
type CheckSingleDomainMsg struct {
	domainID types.DomainID
}

type SingleDomainCheckCompletedMsg struct {
	domainID types.DomainID
	err      error
}

// Screen toggle message types
type ToggleAltScreenMsg struct{}
