// FutBoss AI - Match View
// Author: Bruno Lucena (bruno@lucena.cloud)

package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type MatchState int

const (
	MatchStateLobby MatchState = iota
	MatchStatePlaying
	MatchStateFinished
)

type MatchEvent struct {
	Minute      int
	Type        string
	PlayerName  string
	Team        string
	Description string
}

type MatchModel struct {
	opponent   string
	state      MatchState
	minute     int
	homeScore  int
	awayScore  int
	homeTeam   string
	awayTeam   string
	possession int
	events     []MatchEvent
	width      int
	height     int
}

func NewMatchModel(opponent string) MatchModel {
	return MatchModel{
		opponent:   opponent,
		state:      MatchStateLobby,
		minute:     0,
		homeScore:  0,
		awayScore:  0,
		homeTeam:   "My Team FC",
		awayTeam:   opponent,
		possession: 50,
		events:     []MatchEvent{},
	}
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m MatchModel) Init() tea.Cmd {
	return nil
}

func (m MatchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		if m.state == MatchStatePlaying {
			m.minute++
			if m.minute >= 90 {
				m.state = MatchStateFinished
				return m, nil
			}
			// Simulate match events
			m.simulateMinute()
			return m, tickCmd()
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
			if m.state == MatchStateLobby {
				m.state = MatchStatePlaying
				return m, tickCmd()
			}
		}
	}

	return m, nil
}

func (m *MatchModel) simulateMinute() {
	// Simple random simulation - in real version this uses AI agents
	// This is where Ollama would be called for AI decisions
}

func (m MatchModel) View() string {
	switch m.state {
	case MatchStateLobby:
		return m.renderLobby()
	case MatchStatePlaying:
		return m.renderPlaying()
	case MatchStateFinished:
		return m.renderFinished()
	}
	return ""
}

func (m MatchModel) renderLobby() string {
	content := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════╗
║                      MATCH PREVIEW                        ║
╠══════════════════════════════════════════════════════════╣
║                                                          ║
║     %s                                                   
║                                                          ║
║                        VS                                ║
║                                                          ║
║     %s                                                   
║                                                          ║
╠══════════════════════════════════════════════════════════╣
║                                                          ║
║            Press [ENTER] or [SPACE] to start             ║
║                    Press [Q] to quit                     ║
║                                                          ║
╚══════════════════════════════════════════════════════════╝
`,
		SelectedStyle.Render(m.homeTeam),
		ItemStyle.Render(m.awayTeam),
	)

	return BoxStyle.Render(content)
}

func (m MatchModel) renderPlaying() string {
	// Score display
	score := ScoreStyle.Render(fmt.Sprintf("  %d  -  %d  ", m.homeScore, m.awayScore))

	// Time display
	timeDisplay := fmt.Sprintf("⏱️ %d'", m.minute)

	// Possession bar
	possBar := m.renderPossessionBar()

	// Events log
	eventsLog := m.renderEvents()

	content := fmt.Sprintf(`
%s

%s            %s            %s

%s  %s

━━━━━━━━━━━━━━━━━ MATCH EVENTS ━━━━━━━━━━━━━━━━━

%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[Q] Quit match
`,
		TitleStyle.Render("⚽ LIVE MATCH"),
		ItemStyle.Render(m.homeTeam),
		score,
		ItemStyle.Render(m.awayTeam),
		timeDisplay,
		possBar,
		eventsLog,
	)

	return content
}

func (m MatchModel) renderFinished() string {
	result := "DRAW"
	if m.homeScore > m.awayScore {
		result = "YOU WIN! 🎉"
	} else if m.homeScore < m.awayScore {
		result = "YOU LOSE 😢"
	}

	content := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════╗
║                     FULL TIME                             ║
╠══════════════════════════════════════════════════════════╣
║                                                          ║
║     %s  %d                                               
║                                                          ║
║     %s  %d                                               
║                                                          ║
╠══════════════════════════════════════════════════════════╣
║                                                          ║
║                      %s                                  
║                                                          ║
║              Press [Q] to exit                           ║
╚══════════════════════════════════════════════════════════╝
`,
		m.homeTeam, m.homeScore,
		m.awayTeam, m.awayScore,
		SuccessStyle.Render(result),
	)

	return BoxStyle.Render(content)
}

func (m MatchModel) renderPossessionBar() string {
	homeBlocks := m.possession / 5
	awayBlocks := 20 - homeBlocks

	bar := "Possession: "
	for i := 0; i < homeBlocks; i++ {
		bar += "█"
	}
	bar += "|"
	for i := 0; i < awayBlocks; i++ {
		bar += "░"
	}
	bar += fmt.Sprintf(" %d%% - %d%%", m.possession, 100-m.possession)

	return bar
}

func (m MatchModel) renderEvents() string {
	if len(m.events) == 0 {
		return "  Waiting for action..."
	}

	result := ""
	start := 0
	if len(m.events) > 5 {
		start = len(m.events) - 5
	}
	for _, e := range m.events[start:] {
		icon := "⚽"
		switch e.Type {
		case "goal":
			icon = "⚽"
		case "save":
			icon = "🧤"
		case "card":
			icon = "🟨"
		case "foul":
			icon = "❌"
		}
		result += fmt.Sprintf("  %d' %s %s - %s\n", e.Minute, icon, e.PlayerName, e.Description)
	}

	return result
}
