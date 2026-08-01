package modes

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSnakeResetClearsGameOverState(t *testing.T) {
	m := SnakeModel{}
	m.GenerateGrid()
	m.Err = "Game over! Press r to restart or q to return to the main menu."
	m.snake.score = 3

	var cmd tea.Cmd
	m, _, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if m.Err != "" {
		t.Fatalf("expected reset to clear error, got %q", m.Err)
	}
	if m.snake.score != 0 {
		t.Fatalf("expected reset to clear score, got %d", m.snake.score)
	}
	if cmd == nil {
		t.Fatal("expected reset to restart snake tick")
	}
}
