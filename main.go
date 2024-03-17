package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	url2 "net/url"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := NewModel()

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

type Model struct {
	title     string
	textinput textinput.Model
	terms Terms
	err   error
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter search term: "
	ti.Focus()
	return Model{
		title:     "Testes",
		textinput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			v := m.textinput.Value()
			return m, handleQuerySearch(v)
		}
	case TermsResponseMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}

		m.terms = msg.Terms
		return m, nil
	}
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	s := m.textinput.View() + "\n\n"

	if len(m.terms.List) > 0 {
		s += m.terms.List[0].Definition + "\n\n"
		s += m.terms.List[0].Example + "\n\n"
		s += fmt.Sprintf("upvotes: %d\ndownvotes: %d\n\n", m.terms.List[0].ThumbsUp, m.terms.List[0].ThumbsDown)
	}

	return s
}

func handleQuerySearch(q string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("https://api.urbandictionary.com/v0/define?term=%s", url2.QueryEscape(q))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return TermsResponseMsg{
				Err: err,
			}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return TermsResponseMsg{
				Err: err,
			}
		}

		var terms Terms
		err = json.NewDecoder(res.Body).Decode(&terms)
		if err != nil {
			return TermsResponseMsg{
				Err: err,
			}
		}

		return TermsResponseMsg{
			Terms: terms,
		}
	}
}

type Terms struct {
	List []struct {
		Definition  string    `json:"definition"`
		Permalink   string    `json:"permalink"`
		ThumbsUp    int       `json:"thumbs_up"`
		Author      string    `json:"author"`
		Word        string    `json:"word"`
		Defid       int       `json:"defid"`
		CurrentVote string    `json:"current_vote"`
		WrittenOn   time.Time `json:"written_on"`
		Example     string    `json:"example"`
		ThumbsDown  int       `json:"thumbs_down"`
	} `json:"list"`
}

type TermsResponseMsg struct {
	Terms Terms
	Err   error
}
