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
	minesweeperMaxRows = 16
	minesweeperMaxCols = 30
)

type mineDifficulty int

const (
	difficultyEasy mineDifficulty = iota
	difficultyMedium
	difficultyHard
)

type mineBoardSize int

const (
	boardSmall mineBoardSize = iota
	boardMedium
	boardLarge
)

type mineCell struct {
	mine, revealed, flagged bool
	nearby                  int
}

type MinesweeperModel struct {
	board       [minesweeperMaxRows][minesweeperMaxCols]mineCell
	cursor      point
	rows        int
	cols        int
	mines       int
	flags       int
	difficulty  mineDifficulty
	boardSize   mineBoardSize
	configuring bool
	started     bool
	finished    bool
	won         bool
	message     string
}

var (
	mineTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f4d35e"))
	mineMutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#8b949e"))
	mineBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#52606d"))
	mineCursorStyle = lipgloss.NewStyle().Background(lipgloss.Color("#334155")).Bold(true)
	mineHiddenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cbd5e1"))
	mineFlagStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f97068"))
	mineBombStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f5f"))
	mineNumbers     = []lipgloss.Style{
		{},
		lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#fb7185")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#c084fc")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#22d3ee")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#e5e7eb")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8")),
	}
)

func NewMinesweeperModel() MinesweeperModel {
	m := MinesweeperModel{}
	m.reset()
	return m
}

func (m *MinesweeperModel) reset() {
	m.board = [minesweeperMaxRows][minesweeperMaxCols]mineCell{}
	m.cursor = point{}
	m.configuring = true
	m.applyBoardSize()
	m.flags = 0
	m.started, m.finished, m.won = false, false, false
	m.message = "Choose a difficulty and board size, then press Enter."
}

func (m MinesweeperModel) Update(msg tea.Msg) (MinesweeperModel, bool, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if m.configuring {
			return m.updateConfig(key.String())
		}
		switch key.String() {
		case "q", "esc":
			return m, true, nil
		case "r":
			m.reset()
		case "up", "w", "k":
			m.moveCursor(-1, 0)
		case "down", "s", "j":
			m.moveCursor(1, 0)
		case "left", "a", "h":
			m.moveCursor(0, -1)
		case "right", "d", "l":
			m.moveCursor(0, 1)
		case "f":
			m.toggleFlag()
		case " ", "space", "enter":
			m.reveal(m.cursor.row, m.cursor.col)
		}
	}
	return m, false, nil
}

func (m MinesweeperModel) updateConfig(key string) (MinesweeperModel, bool, tea.Cmd) {
	switch key {
	case "q", "esc":
		return m, true, nil
	case "left", "a", "h":
		if m.difficulty > difficultyEasy {
			m.difficulty--
		}
	case "right", "d", "l":
		if m.difficulty < difficultyHard {
			m.difficulty++
		}
	case "up", "w", "k":
		if m.boardSize > boardSmall {
			m.boardSize--
		}
	case "down", "s", "j":
		if m.boardSize < boardLarge {
			m.boardSize++
		}
	case " ", "enter", "space":
		m.configuring = false
		m.resetBoard()
	case "r":
		m.reset()
	}
	m.applyBoardSize()
	return m, false, nil
}

func (m *MinesweeperModel) applyBoardSize() {
	switch m.boardSize {
	case boardMedium:
		m.rows, m.cols = 16, 16
	case boardLarge:
		m.rows, m.cols = 16, 30
	default:
		m.rows, m.cols = 9, 9
	}
	m.mines = m.mineCount()
}

func (m MinesweeperModel) mineCount() int {
	boardCells := m.rows * m.cols
	percent := 0.10
	if m.difficulty == difficultyMedium {
		percent = 0.15
	} else if m.difficulty == difficultyHard {
		percent = 0.20
	}
	count := int(float64(boardCells) * percent)
	if count < 1 {
		return 1
	}
	return count
}

func (m *MinesweeperModel) resetBoard() {
	m.board = [minesweeperMaxRows][minesweeperMaxCols]mineCell{}
	m.cursor = point{}
	m.flags = 0
	m.started, m.finished, m.won = false, false, false
	m.message = "Reveal squares and mark suspected mines."
}

func (m *MinesweeperModel) moveCursor(row, col int) {
	m.cursor.row = (m.cursor.row + row + m.rows) % m.rows
	m.cursor.col = (m.cursor.col + col + m.cols) % m.cols
}

func (m *MinesweeperModel) toggleFlag() {
	if m.finished {
		return
	}
	cell := &m.board[m.cursor.row][m.cursor.col]
	if cell.revealed {
		return
	}
	if !cell.flagged && m.flags >= m.mines {
		m.message = "No flags left — unflag one first."
		return
	}
	cell.flagged = !cell.flagged
	if cell.flagged {
		m.flags++
	} else {
		m.flags--
	}
	m.message = "Mark the squares you think contain mines."
}

func (m *MinesweeperModel) reveal(row, col int) {
	if m.finished || m.board[row][col].flagged || m.board[row][col].revealed {
		return
	}
	if !m.started {
		m.placeMines(row, col)
		m.started = true
	}
	if m.board[row][col].mine {
		m.finished, m.won = true, false
		m.message = "Boom. Press r to try again."
		for r := 0; r < m.rows; r++ {
			for c := 0; c < m.cols; c++ {
				if m.board[r][c].mine {
					m.board[r][c].revealed = true
				}
			}
		}
		return
	}
	m.revealArea(row, col)
	if m.revealedSafeCells() == m.rows*m.cols-m.mines {
		m.finished, m.won = true, true
		m.message = "Board cleared! Press r for another round."
	}
}

func (m *MinesweeperModel) placeMines(safeRow, safeCol int) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	placed := 0
	for placed < m.mines {
		row, col := rng.Intn(m.rows), rng.Intn(m.cols)
		if (row == safeRow && col == safeCol) || m.board[row][col].mine {
			continue
		}
		m.board[row][col].mine = true
		placed++
	}
	for row := 0; row < m.rows; row++ {
		for col := 0; col < m.cols; col++ {
			if !m.board[row][col].mine {
				m.board[row][col].nearby = m.nearbyMines(row, col)
			}
		}
	}
}

func (m MinesweeperModel) nearbyMines(row, col int) int {
	total := 0
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			r, c := row+dr, col+dc
			if r >= 0 && r < m.rows && c >= 0 && c < m.cols && m.board[r][c].mine {
				total++
			}
		}
	}
	return total
}

func (m *MinesweeperModel) revealArea(row, col int) {
	if row < 0 || row >= m.rows || col < 0 || col >= m.cols {
		return
	}
	cell := &m.board[row][col]
	if cell.revealed || cell.flagged || cell.mine {
		return
	}
	cell.revealed = true
	if cell.nearby != 0 {
		return
	}
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr != 0 || dc != 0 {
				m.revealArea(row+dr, col+dc)
			}
		}
	}
}

func (m MinesweeperModel) revealedSafeCells() int {
	total := 0
	for row := 0; row < m.rows; row++ {
		for col := 0; col < m.cols; col++ {
			if m.board[row][col].revealed && !m.board[row][col].mine {
				total++
			}
		}
	}
	return total
}

func (m MinesweeperModel) View() string {
	if m.configuring {
		return m.configView()
	}
	var b strings.Builder
	b.WriteString(mineTitleStyle.Render("MINESWEEPER"))
	b.WriteString(mineMutedStyle.Render("  /  clear the field"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  MINES %02d/%02d    ", m.flags, m.mines))
	if m.finished && m.won {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Render("CLEARED"))
	} else if m.finished {
		b.WriteString(mineBombStyle.Render("DETONATED"))
	} else {
		b.WriteString(mineMutedStyle.Render("IN PLAY"))
	}
	b.WriteString("\n\n")
	for col := 1; col <= m.cols; col++ {
		b.WriteString(mineBorderStyle.Render(fmt.Sprintf("%2d", col)))
	}
	b.WriteString("\n")
	for row := 0; row < m.rows; row++ {
		b.WriteString(fmt.Sprintf("%2d", row+1))
		for col := 0; col < m.cols; col++ {
			b.WriteString(m.renderCell(row, col))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(mineMutedStyle.Render(m.message))
	b.WriteString("\n\n")
	b.WriteString(mineMutedStyle.Render("ARROWS / WASD move   SPACE reveal   F flag   R restart   Q menu"))
	return b.String()
}

func (m MinesweeperModel) configView() string {
	difficulties := []string{"EASY", "MEDIUM", "HARD"}
	sizes := []string{"SMALL  9 x 9", "MEDIUM 16 x 16", "LARGE  16 x 30"}
	var b strings.Builder
	b.WriteString(mineTitleStyle.Render("MINESWEEPER / SETUP"))
	b.WriteString("\n\n")
	b.WriteString(mineMutedStyle.Render("Choose your field before you begin."))
	b.WriteString("\n\n")
	b.WriteString("  DIFFICULTY   ")
	for i, difficulty := range difficulties {
		if mineDifficulty(i) == m.difficulty {
			b.WriteString(mineTitleStyle.Render("[" + difficulty + "]"))
		} else {
			b.WriteString(mineMutedStyle.Render(" " + difficulty + " "))
		}
		b.WriteString("  ")
	}
	b.WriteString("\n")
	b.WriteString("  BOARD SIZE   ")
	for i, size := range sizes {
		if mineBoardSize(i) == m.boardSize {
			b.WriteString(mineTitleStyle.Render("[" + size + "]"))
		} else {
			b.WriteString(mineMutedStyle.Render(" " + size + " "))
		}
		b.WriteString("  ")
	}
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  FIELD  %d x %d     MINES  %d     DENSITY  %d%%", m.rows, m.cols, m.mines, m.mines*100/(m.rows*m.cols)))
	b.WriteString("\n\n")
	b.WriteString(mineMutedStyle.Render("←/→ difficulty   ↑/↓ board size   ENTER start   Q menu"))
	return b.String()
}

func (m MinesweeperModel) renderCell(row, col int) string {
	cell := m.board[row][col]
	value, style := "■", mineHiddenStyle
	if !cell.revealed {
		if cell.flagged {
			value, style = "⚑", mineFlagStyle
		}
	} else if cell.mine {
		value, style = "✹", mineBombStyle
	} else if cell.nearby > 0 {
		value, style = fmt.Sprint(cell.nearby), mineNumbers[cell.nearby]
	} else {
		value = "·"
	}
	if row == m.cursor.row && col == m.cursor.col {
		return mineCursorStyle.Render(" " + style.Render(value) + " ")
	}
	return mineBorderStyle.Render(" ") + style.Render(value) + mineBorderStyle.Render(" ")
}
