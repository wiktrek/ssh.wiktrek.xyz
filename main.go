package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/joho/godotenv"

	"github.com/wiktrek/ssh.wiktrek.xyz/modes"
)

func main() {
	_ = godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "2332"
	}
	serverAddress := ":" + port

	s, _ := wish.NewServer(
		wish.WithAddress(serverAddress),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
		),
	)
	fmt.Println("SSH server started at 127.0.0.1" + serverAddress)
	s.ListenAndServe()
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return model{}, []tea.ProgramOption{tea.WithAltScreen()}
}

type model struct {
	choice  int
	clicker modes.ClickerModel
	snake   modes.SnakeModel
}

func link(url, text string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Underline(true)
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, style.Render(text))
}
func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.choice == 0 && msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "c" && m.choice == 0 {
			m.choice = 1
			m.clicker = modes.ClickerModel{}
			return m, modes.IncomeTick()
		}
		if msg.String() == "s" && m.choice == 0 {
			m.choice = 2
			m.snake = modes.SnakeModel{}
			m.snake.GenerateGrid()
			return m, modes.SnakeTick()
		}
	}
	if m.choice == 1 {
		var exit bool
		var cmd tea.Cmd
		m.clicker, exit, cmd = m.clicker.Update(msg)
		if exit {
			m.choice = 0
		}
		return m, cmd
	}
	if m.choice == 2 {
		var exit bool
		var cmd tea.Cmd
		m.snake, exit, cmd = m.snake.Update(msg)
		if exit {
			m.choice = 0
		}
		return m, cmd
	}
	return m, nil
}
func (m model) View() string {
	if m.choice == 1 {
		return m.clicker.View()
	}
	if m.choice == 2 {
		return m.snake.View()
	}
	return "Wow you know how to use ssh!\nPress c to play Clicker\nPress s to play Snake!\n\nPress q to quit.\n\n" + link("https://github.com/wiktrek/ssh.wiktrek.xyz", "GitHub Repo") + "\n"
}
