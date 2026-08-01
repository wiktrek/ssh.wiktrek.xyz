package modes

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	SIZE             = 20
	snakeTick        = 180 * time.Millisecond
	initialDirection = directionRight
)

// 0 = nothing 1 = apple 2 = snake
type P struct {
	v int
}

type point struct {
	row int
	col int
}

type direction int

const (
	directionUp direction = iota
	directionDown
	directionLeft
	directionRight
)

type Snake struct {
	score     int
	body      []point
	direction direction
}
type Grid struct {
	values [SIZE][SIZE]P
}
type SnakeModel struct {
	snake Snake
	grid  Grid
	Err   string
}

type snakeTickMsg struct{}

var (
	snakeStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	snakeHeadStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	appleStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	backgroundStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0"))
)

// SnakeTick starts the snake's movement loop.
func SnakeTick() tea.Cmd {
	return tea.Tick(snakeTick, func(time.Time) tea.Msg {
		return snakeTickMsg{}
	})
}

func (m SnakeModel) Update(msg tea.Msg) (SnakeModel, bool, tea.Cmd) {
	switch msg := msg.(type) {
	case snakeTickMsg:
		if m.Err != "" {
			return m, false, nil
		}
		if m.move() {
			return m, false, SnakeTick()
		}
		return m, false, nil
	case tea.KeyMsg:
		if msg.String() == "r" {
			m.GenerateGrid()
			return m, false, SnakeTick()
		}
		if m.Err != "" && msg.String() != "q" {
			return m, false, nil
		}
		switch msg.String() {
		case "w":
			m.changeDirection(directionUp)
		case "s":
			m.changeDirection(directionDown)
		case "a":
			m.changeDirection(directionLeft)
		case "d":
			m.changeDirection(directionRight)
		case "q":
			m.Err = ""
			return m, true, nil
		}
	}
	return m, false, nil
}
func (m *SnakeModel) GenerateGrid() {
	m.Err = ""
	m.snake = Snake{
		body:      []point{{row: 0, col: 0}},
		direction: initialDirection,
	}
	for i := range SIZE {
		for j := range SIZE {
			m.grid.values[i][j] = P{0}
		}
	}
	m.spawnApple()
	m.renderSnake()
}

func (m *SnakeModel) changeDirection(next direction) {
	if len(m.snake.body) > 1 {
		if (m.snake.direction == directionUp && next == directionDown) ||
			(m.snake.direction == directionDown && next == directionUp) ||
			(m.snake.direction == directionLeft && next == directionRight) ||
			(m.snake.direction == directionRight && next == directionLeft) {
			return
		}
	}
	m.snake.direction = next
}

func (m *SnakeModel) move() bool {
	head := m.snake.body[0]
	next := head
	switch m.snake.direction {
	case directionUp:
		next.row--
	case directionDown:
		next.row++
	case directionLeft:
		next.col--
	case directionRight:
		next.col++
	}

	if next.row < 0 || next.row >= SIZE || next.col < 0 || next.col >= SIZE || m.occupied(next) {
		m.Err = "Game over! Press r to restart or q to return to the main menu."
		return false
	}

	grew := m.grid.values[next.row][next.col].v == 1
	m.snake.body = append([]point{next}, m.snake.body...)
	if grew {
		m.snake.score++
		m.spawnApple()
	} else {
		m.snake.body = m.snake.body[:len(m.snake.body)-1]
	}
	m.renderSnake()
	return true
}

func (m SnakeModel) occupied(p point) bool {
	for _, segment := range m.snake.body {
		if segment == p {
			return true
		}
	}
	return false
}

func (m *SnakeModel) spawnApple() {
	for {
		apple := point{row: rand.Intn(SIZE), col: rand.Intn(SIZE)}
		if !m.occupied(apple) {
			m.grid.values[apple.row][apple.col] = P{1}
			return
		}
	}
}

func (m *SnakeModel) renderSnake() {
	for i := range SIZE {
		for j := range SIZE {
			if m.grid.values[i][j].v == 2 {
				m.grid.values[i][j] = P{0}
			}
		}
	}
	for _, segment := range m.snake.body {
		m.grid.values[segment.row][segment.col] = P{2}
	}
}
func (m SnakeModel) render_grid() string {
	s := "\n"
	for i := range SIZE {
		row := ""
		for v := range SIZE {
			switch m.grid.values[i][v].v {
			case 0:
				row += backgroundStyle.Render("██")
			case 1:
				row += appleStyle.Render("██")
			case 2:
				if len(m.snake.body) > 0 && m.snake.body[0] == (point{row: i, col: v}) {
					row += snakeHeadStyle.Render("██")
				} else {
					row += snakeStyle.Render("██")
				}
			}
		}
		// Two characters wide compensates for the terminal's tall character cells.
		s += row + "\n"
	}
	return s
}

func (m SnakeModel) View() string {
	view := "Snake: \nScore: " + fmt.Sprint(m.snake.score) + "\n"
	if m.Err != "" {
		view += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(errorColor)).Render(m.Err)
	}
	view += m.render_grid()
	view += "\n\nUse w/a/s/d to move."
	view += "\nPress r to restart."
	view += "\n\n\nPress q to return to main menu!"
	return view
}
