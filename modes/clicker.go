package modes

import (
	"fmt"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	errorColor            = "#ff5f5f"
	workerCost            = 50
	workerPriceMultiplier = 1.2
	workerIncomeInterval  = 3 * time.Second
)

type incomeTickMsg struct{}

type ClickerModel struct {
	Money   int
	Workers int
	Err     string
}

func IncomeTick() tea.Cmd {
	return tea.Tick(workerIncomeInterval, func(time.Time) tea.Msg {
		return incomeTickMsg{}
	})
}

func (m ClickerModel) Update(msg tea.Msg) (ClickerModel, bool, tea.Cmd) {
	switch msg := msg.(type) {
	case incomeTickMsg:
		m.Money += m.Workers
		return m, false, IncomeTick()
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			m.Money++
		case "w":
			workerPrice := m.workerPrice()
			if m.Money >= workerPrice {
				m.Money -= workerPrice
				m.Workers++
				m.Err = ""
			} else {
				m.Err = "Not enough money to hire a worker!"
			}
		case "q":
			m.Err = ""
			return m, true, nil
		}
	}
	return m, false, nil
}

func (m ClickerModel) workerPrice() int {
	return int(math.Round(float64(workerCost) * math.Pow(workerPriceMultiplier, float64(m.Workers))))
}

// View renders the clicker mode.
func (m ClickerModel) View() string {
	view := "Clicker"
	view += "\nMoney: " + fmt.Sprint(m.Money)
	view += "\nWorkers: " + fmt.Sprint(m.Workers)
	view += "\nPress space to earn money!"
	view += "\nPress w to hire a worker (cost: " + fmt.Sprint(m.workerPrice()) + ")"
	if m.Err != "" {
		view += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(errorColor)).Render(m.Err)
	}
	view += "\n\n\nPress q to return to main menu!"
	return view
}
