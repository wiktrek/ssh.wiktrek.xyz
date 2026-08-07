package modes

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	dinoWidth      = 58
	dinoGround     = 8
	dinoStartX     = 7
	dinoTick       = 95 * time.Millisecond
	gravity        = 0.72
	jumpVelocity   = -4.8
	minObstacleGap = 17
	maxObstacleGap = 28
)

type dinoTickMsg struct{}

type dinoObstacle struct {
	x      int
	height int
}

type DinoModel struct {
	playerY   float64
	velocity  float64
	score     int
	best      int
	frame     int
	started   bool
	gameOver  bool
	nextGap   int
	obstacles []dinoObstacle
}

var (
	dinoTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f5c542"))
	dinoScoreStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#a9b7c6"))
	dinoPlayerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f5c542"))
	dinoCactusStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#78c091"))
	dinoGroundStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#607080"))
	dinoMutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#728096"))
	dinoDangerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff6b6b"))
)

func NewDinoModel() DinoModel {
	return DinoModel{best: 0, nextGap: minObstacleGap + rand.Intn(maxObstacleGap-minObstacleGap+1)}
}

func DinoTick() tea.Cmd {
	return tea.Tick(dinoTick, func(time.Time) tea.Msg { return dinoTickMsg{} })
}

func (m DinoModel) onGround() bool { return m.playerY >= 0 }

func (m *DinoModel) reset() {
	best := m.best
	*m = NewDinoModel()
	m.best = best
}

func (m *DinoModel) jump() {
	if !m.started {
		m.started = true
	}
	if m.onGround() && !m.gameOver {
		m.velocity = jumpVelocity
	}
}

func (m *DinoModel) step() {
	if !m.started || m.gameOver {
		return
	}
	m.frame++
	m.playerY += m.velocity
	m.velocity += gravity
	if m.playerY > 0 {
		m.playerY = 0
		m.velocity = 0
	}

	for i := range m.obstacles {
		m.obstacles[i].x--
	}
	if len(m.obstacles) == 0 || m.obstacles[len(m.obstacles)-1].x < dinoWidth-m.nextGap {
		m.obstacles = append(m.obstacles, dinoObstacle{x: dinoWidth + 2, height: 1 + rand.Intn(2)})
		m.nextGap = minObstacleGap + rand.Intn(maxObstacleGap-minObstacleGap+1)
	}
	for len(m.obstacles) > 0 && m.obstacles[0].x < -2 {
		m.obstacles = m.obstacles[1:]
	}
	m.score++
	if m.score > m.best {
		m.best = m.score
	}

	for _, obstacle := range m.obstacles {
		if obstacle.x <= dinoStartX+2 && obstacle.x+1 >= dinoStartX && m.playerY >= float64(-obstacle.height+1) {
			m.gameOver = true
			return
		}
	}
}

func (m DinoModel) Update(msg tea.Msg) (DinoModel, bool, tea.Cmd) {
	switch msg := msg.(type) {
	case dinoTickMsg:
		m.step()
		return m, false, DinoTick()
	case tea.KeyMsg:
		switch msg.String() {
		case " ", "up", "w":
			if m.gameOver {
				m.reset()
			} else {
				m.jump()
			}
		case "r":
			m.reset()
		case "q", "esc":
			return m, true, nil
		}
	}
	return m, false, nil
}

func (m DinoModel) cloudLine() string {
	var line strings.Builder
	for i := 0; i < dinoWidth; i++ {
		if (i+m.frame/8)%19 == 0 || (i+m.frame/8)%29 == 0 {
			line.WriteString("·")
		} else {
			line.WriteByte(' ')
		}
	}
	return line.String()
}

func (m DinoModel) playfield() string {
	rows := make([][]string, dinoGround+1)
	for row := range rows {
		rows[row] = make([]string, dinoWidth)
		for col := range rows[row] {
			rows[row][col] = " "
		}
	}
	playerRow := dinoGround - int(-m.playerY)
	if playerRow < 0 {
		playerRow = 0
	}
	if playerRow > dinoGround-1 {
		playerRow = dinoGround - 1
	}
	if playerRow >= 0 && playerRow < len(rows) {
		rows[playerRow][dinoStartX] = dinoPlayerStyle.Render("▰")
		rows[playerRow][dinoStartX+1] = dinoPlayerStyle.Render("▰")
		rows[playerRow][dinoStartX+2] = dinoPlayerStyle.Render("▰")
	}
	for _, obstacle := range m.obstacles {
		for offset := 0; offset < obstacle.height; offset++ {
			row := dinoGround - 1 - offset
			if obstacle.x >= 0 && obstacle.x < dinoWidth && row >= 0 {
				rows[row][obstacle.x] = dinoCactusStyle.Render("▐")
			}
			if obstacle.x+1 >= 0 && obstacle.x+1 < dinoWidth && row >= 0 {
				rows[row][obstacle.x+1] = dinoCactusStyle.Render("▌")
			}
		}
	}
	for col := range rows[dinoGround] {
		rows[dinoGround][col] = dinoGroundStyle.Render("─")
	}
	var out strings.Builder
	for _, row := range rows {
		out.WriteString(strings.Join(row, ""))
		out.WriteByte('\n')
	}
	return out.String()
}

func (m DinoModel) View() string {
	var out strings.Builder
	out.WriteString(dinoTitleStyle.Render("DINO RUN"))
	out.WriteString("   ")
	out.WriteString(dinoScoreStyle.Render(fmt.Sprintf("SCORE %04d   HI %04d", m.score/3, m.best/3)))
	out.WriteByte('\n')
	out.WriteString(dinoMutedStyle.Render(m.cloudLine()))
	out.WriteByte('\n')
	out.WriteString(m.playfield())
	if m.gameOver {
		out.WriteString(dinoDangerStyle.Render("GAME OVER"))
		out.WriteString("  ")
		out.WriteString(dinoMutedStyle.Render("press SPACE or R to run again"))
	} else if !m.started {
		out.WriteString(dinoMutedStyle.Render("SPACE / ↑  jump     Q  menu"))
	} else {
		out.WriteString(dinoMutedStyle.Render("SPACE / ↑  jump     R  restart     Q  menu"))
	}
	return out.String()
}
