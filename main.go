package main

import (
	"fmt"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
)

func main() {
	s, _ := wish.NewServer(
		wish.WithAddress(serverAddress),
		wish.WithHostKeyPath(hostKeyPath),
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
	money   int
	workers int
	err     string
}
type incomeTickMsg struct{}

type model struct {
	choice  int
	clicker clicker
}

func workerIncomeTick() tea.Cmd {
	return tea.Tick(workerIncomeInterval, func(time.Time) tea.Msg {
		return incomeTickMsg{}
	})
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case incomeTickMsg:
		if m.choice == 1 {
			m.clicker.money += m.clicker.workers
			return m, workerIncomeTick()
		}
		return m, nil
	case tea.KeyMsg:
		if m.choice == 0 && msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "c" && m.choice == 0 {
			m.choice = 1
			m.clicker = clicker{money: 0}
			return m, workerIncomeTick()
		}
		if m.choice == 1 {
			if msg.String() == " " {
				m.clicker.money++
				return m, nil
			}
			if msg.String() == "w" {
				workerPrice := int(math.Round(float64(workerCost) * math.Pow(workerPriceMultiplier, float64(m.clicker.workers))))
				if m.clicker.money >= workerPrice {
					m.clicker.money -= workerPrice
					m.clicker.workers++
					m.clicker.err = ""
					return m, nil
				} else {
					m.clicker.err = "Not enough money to hire a worker!"
				}
			}
			if msg.String() == "q" {
				m.choice = 0
				m.clicker.err = ""
				return m, nil
			}
		}
	}
	return m, nil
}
func (m model) clickerView() string {
	clicker := "Clicker"
	clicker += "\nMoney: " + fmt.Sprint(m.clicker.money)
	clicker += "\nWorkers: " + fmt.Sprint(m.clicker.workers)
	clicker += "\nPress space to earn money!"
	currentPrice := int(math.Round(float64(workerCost) * math.Pow(workerPriceMultiplier, float64(m.clicker.workers))))
	clicker += "\nPress w to hire a worker (cost: " + fmt.Sprint(currentPrice) + ")"
	if m.clicker.err != "" {
		clicker += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(errorColor)).Render(m.clicker.err)
	}
	clicker += "\n\n\nPress q to return to main menu!"
	return clicker
}
func (m model) View() string {
	if m.choice == 1 {
		return m.clickerView()
	}
	return "Wow you know how ot use ssh!\nPress c to play Clicker\n\nPress q to quit."
}
