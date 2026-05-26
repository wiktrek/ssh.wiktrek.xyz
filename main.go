package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
)

func main() {
	s, _ := wish.NewServer(
		wish.WithAddress(":2222"),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
		),
	)
	s.ListenAndServe()
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return model{}, []tea.ProgramOption{tea.WithAltScreen()}
}

type clicker struct {
	money int
}
type model struct {
	choice  int
	clicker clicker
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "c" {
			m.choice = 1
			m.clicker = clicker{money: 0}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.choice == 1 {
		return "I will add Clicker here!\n\nPress q to quit."
	}
	return "Wow you know how to use ssh!\nPress c to play Clicker\n\nPress q to quit."
}
