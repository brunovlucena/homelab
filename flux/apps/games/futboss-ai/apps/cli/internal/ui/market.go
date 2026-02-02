// FutBoss AI - Market View
// Author: Bruno Lucena (bruno@lucena.cloud)

package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MarketModel struct {
	table    table.Model
	balance  int
	message  string
	width    int
	height   int
}

func NewMarketModel() MarketModel {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Pos", Width: 5},
		{Title: "OVR", Width: 5},
		{Title: "SPD", Width: 5},
		{Title: "FIN", Width: 5},
		{Title: "PAS", Width: 5},
		{Title: "DEF", Width: 5},
		{Title: "Price", Width: 8},
	}

	rows := []table.Row{
		{"Ronaldo Silva", "ST", "86", "88", "92", "70", "30", "1,500"},
		{"Messi Santos", "RW", "89", "85", "88", "90", "35", "1,800"},
		{"Casemiro Jr", "CDM", "84", "65", "55", "75", "88", "1,200"},
		{"Alisson Becker", "GK", "88", "50", "25", "45", "90", "1,400"},
		{"Marquinhos", "CB", "85", "70", "35", "60", "89", "1,100"},
		{"Raphinha Costa", "LW", "82", "88", "78", "75", "40", "950"},
		{"Fred Lima", "CM", "78", "70", "60", "80", "65", "600"},
		{"Endrick Felipe", "ST", "79", "85", "82", "55", "25", "800"},
		{"Militão", "CB", "83", "75", "30", "55", "87", "1,000"},
		{"Paquetá", "CAM", "81", "72", "75", "82", "50", "850"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(12),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#000")).
		Background(primaryColor).
		Bold(true)
	t.SetStyles(s)

	return MarketModel{
		table:   t,
		balance: 1000,
	}
}

func (m MarketModel) Init() tea.Cmd {
	return nil
}

func (m MarketModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "b"))):
			selected := m.table.SelectedRow()
			if len(selected) > 0 {
				m.message = fmt.Sprintf("Attempting to buy %s...", selected[0])
				// TODO: Call API to buy player
			}
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m MarketModel) View() string {
	title := TitleStyle.Render("🏪 Transfer Market")
	balanceDisplay := StatsBoxStyle.Render(fmt.Sprintf("💰 Balance: %d FTC", m.balance))

	help := HelpStyle.Render("↑/↓: Navigate | Enter/B: Buy | /: Filter | Q: Quit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		balanceDisplay,
		"",
		m.table.View(),
		"",
		help,
	)

	if m.message != "" {
		content += "\n" + SuccessStyle.Render(m.message)
	}

	return BoxStyle.Render(content)
}

