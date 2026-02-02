// FutBoss AI - Main TUI Model
// Author: Bruno Lucena (bruno@lucena.cloud)

package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type view int

const (
	viewDashboard view = iota
	viewTeam
	viewMarket
	viewMatch
	viewWallet
)

type MainModel struct {
	currentView  view
	width        int
	height       int
	teamName     string
	balance      int
	wins         int
	draws        int
	losses       int
	notification string
}

func NewMainModel() MainModel {
	return MainModel{
		currentView: viewDashboard,
		teamName:    "My Team FC",
		balance:     1000,
		wins:        0,
		draws:       0,
		losses:      0,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Help):
			m.notification = "Press: 1=Dashboard, 2=Team, 3=Market, 4=Match, 5=Wallet, q=Quit"
		case msg.String() == "1":
			m.currentView = viewDashboard
			m.notification = ""
		case msg.String() == "2":
			m.currentView = viewTeam
			m.notification = ""
		case msg.String() == "3":
			m.currentView = viewMarket
			m.notification = ""
		case msg.String() == "4":
			m.currentView = viewMatch
			m.notification = ""
		case msg.String() == "5":
			m.currentView = viewWallet
			m.notification = ""
		}
	}

	return m, nil
}

func (m MainModel) View() string {
	// Title bar
	title := TitleStyle.Render("⚽ FutBoss AI - " + m.getViewTitle())

	// Stats bar
	stats := m.renderStatsBar()

	// Main content
	content := m.renderContent()

	// Status bar
	statusBar := m.renderStatusBar()

	// Combine all
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		stats,
		content,
		statusBar,
	)
}

func (m MainModel) getViewTitle() string {
	switch m.currentView {
	case viewDashboard:
		return "Dashboard"
	case viewTeam:
		return "Team Management"
	case viewMarket:
		return "Transfer Market"
	case viewMatch:
		return "Match Center"
	case viewWallet:
		return "Wallet"
	default:
		return "FutBoss"
	}
}

func (m MainModel) renderStatsBar() string {
	teamBox := StatsBoxStyle.Render(fmt.Sprintf("🏟️ %s", m.teamName))
	balanceBox := StatsBoxStyle.Render(fmt.Sprintf("💰 %d FTC", m.balance))
	recordBox := StatsBoxStyle.Render(fmt.Sprintf("📊 %dW-%dD-%dL", m.wins, m.draws, m.losses))

	return lipgloss.JoinHorizontal(lipgloss.Top, teamBox, balanceBox, recordBox)
}

func (m MainModel) renderContent() string {
	switch m.currentView {
	case viewDashboard:
		return m.renderDashboard()
	case viewTeam:
		return m.renderTeamView()
	case viewMarket:
		return m.renderMarketView()
	case viewMatch:
		return m.renderMatchView()
	case viewWallet:
		return m.renderWalletView()
	default:
		return "Unknown view"
	}
}

func (m MainModel) renderDashboard() string {
	content := `
┌─────────────────────────────────────────────────────────────┐
│                    WELCOME TO FUTBOSS AI                     │
│                                                              │
│  Manage your football team with AI-powered players!         │
│                                                              │
│  Quick Actions:                                              │
│  [2] Team     - View and manage your squad                  │
│  [3] Market   - Buy and sell players                        │
│  [4] Match    - Start a new match                           │
│  [5] Wallet   - Manage your FutCoins                        │
│                                                              │
│  Press [?] for help                                         │
└─────────────────────────────────────────────────────────────┘

`
	return BoxStyle.Render(content)
}

func (m MainModel) renderTeamView() string {
	header := HeaderStyle.Render(fmt.Sprintf("%-20s %-5s %-8s %-6s", "NAME", "POS", "OVERALL", "VALUE"))

	players := []string{
		fmt.Sprintf("%-20s %-5s %-8d %-6d", "Loading...", "---", 0, 0),
	}

	content := header + "\n"
	for _, p := range players {
		content += ItemStyle.Render(p) + "\n"
	}

	return BoxStyle.Render(content)
}

func (m MainModel) renderMarketView() string {
	return BoxStyle.Render(`
Transfer Market
━━━━━━━━━━━━━━━

Browse available players and make transfers.

Use arrow keys to navigate, Enter to buy.

[Loading players from server...]
`)
}

func (m MainModel) renderMatchView() string {
	return BoxStyle.Render(`
Match Center
━━━━━━━━━━━━

Start a new match or view ongoing matches.

[N] New Match vs AI
[P] New Match vs Player
[H] Match History

`)
}

func (m MainModel) renderWalletView() string {
	return BoxStyle.Render(fmt.Sprintf(`
Wallet
━━━━━━

Balance: %d FutCoins

[B] Buy tokens (PIX)
[C] Buy tokens (Bitcoin)
[T] Transaction history

`, m.balance))
}

func (m MainModel) renderStatusBar() string {
	help := "1:Dashboard 2:Team 3:Market 4:Match 5:Wallet ?:Help q:Quit"
	if m.notification != "" {
		help = m.notification
	}
	return StatusBarStyle.Width(m.width).Render(help)
}

// Key bindings
type keyMap struct {
	Quit key.Binding
	Help key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

