// FutBoss AI - Team View
// Author: Bruno Lucena (bruno@lucena.cloud)

package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Player struct {
	ID          string
	Name        string
	Position    string
	Overall     int
	Speed       int
	Finishing   int
	Passing     int
	Defense     int
	Price       int
	Temperament string
}

func (p Player) Title() string       { return p.Name }
func (p Player) Description() string { return fmt.Sprintf("%s | OVR: %d | 💰 %d", p.Position, p.Overall, p.Price) }
func (p Player) FilterValue() string { return p.Name }

type TeamModel struct {
	list     list.Model
	players  []Player
	selected *Player
	width    int
	height   int
}

func NewTeamModel() TeamModel {
	// Mock players
	players := []Player{
		{ID: "1", Name: "Roberto Carlos", Position: "GK", Overall: 78, Speed: 45, Finishing: 20, Passing: 65, Defense: 85, Price: 500, Temperament: "calm"},
		{ID: "2", Name: "Marcelo Silva", Position: "CB", Overall: 75, Speed: 60, Finishing: 30, Passing: 55, Defense: 82, Price: 450, Temperament: "calculated"},
		{ID: "3", Name: "João Pedro", Position: "CB", Overall: 73, Speed: 55, Finishing: 25, Passing: 50, Defense: 80, Price: 400, Temperament: "calm"},
		{ID: "4", Name: "Lucas Mendes", Position: "LB", Overall: 72, Speed: 78, Finishing: 40, Passing: 65, Defense: 70, Price: 380, Temperament: "explosive"},
		{ID: "5", Name: "Rafael Costa", Position: "RB", Overall: 71, Speed: 80, Finishing: 35, Passing: 60, Defense: 68, Price: 350, Temperament: "explosive"},
		{ID: "6", Name: "Bruno Fernandes", Position: "CM", Overall: 82, Speed: 70, Finishing: 75, Passing: 88, Defense: 55, Price: 800, Temperament: "calculated"},
		{ID: "7", Name: "Gabriel Santos", Position: "CM", Overall: 76, Speed: 68, Finishing: 60, Passing: 78, Defense: 60, Price: 500, Temperament: "calm"},
		{ID: "8", Name: "Diego Alves", Position: "CAM", Overall: 79, Speed: 75, Finishing: 78, Passing: 82, Defense: 40, Price: 650, Temperament: "explosive"},
		{ID: "9", Name: "Neymar Jr", Position: "LW", Overall: 88, Speed: 90, Finishing: 85, Passing: 80, Defense: 30, Price: 1500, Temperament: "explosive"},
		{ID: "10", Name: "Vinicius Costa", Position: "RW", Overall: 80, Speed: 92, Finishing: 75, Passing: 70, Defense: 35, Price: 900, Temperament: "explosive"},
		{ID: "11", Name: "Pedro Striker", Position: "ST", Overall: 84, Speed: 82, Finishing: 90, Passing: 65, Defense: 25, Price: 1200, Temperament: "calculated"},
	}

	items := make([]list.Item, len(players))
	for i, p := range players {
		items[i] = p
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "⚽ Team Roster"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return TeamModel{
		list:    l,
		players: players,
	}
}

func (m TeamModel) Init() tea.Cmd {
	return nil
}

func (m TeamModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if i, ok := m.list.SelectedItem().(Player); ok {
				m.selected = &i
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m TeamModel) View() string {
	if m.selected != nil {
		return m.renderPlayerDetail(*m.selected)
	}
	return m.list.View()
}

func (m TeamModel) renderPlayerDetail(p Player) string {
	posStyle := GetPositionStyle(p.Position)

	detail := fmt.Sprintf(`
%s
%s

Position: %s
Overall: %d

━━━ ATTRIBUTES ━━━
Speed:     %s %d
Finishing: %s %d
Passing:   %s %d
Defense:   %s %d

Temperament: %s
Market Value: 💰 %d FTC

[ESC] Back  [S] Sell  [T] Tactics
`,
		TitleStyle.Render(p.Name),
		posStyle.Render(p.Position),
		p.Position,
		p.Overall,
		RenderAttrBar(p.Speed), p.Speed,
		RenderAttrBar(p.Finishing), p.Finishing,
		RenderAttrBar(p.Passing), p.Passing,
		RenderAttrBar(p.Defense), p.Defense,
		p.Temperament,
		p.Price,
	)

	return BoxStyle.Render(detail)
}

