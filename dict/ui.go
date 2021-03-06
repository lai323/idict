package dict

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	idictconfig "github.com/lai323/idict/config"
	"github.com/lai323/idict/ui"
	te "github.com/muesli/termenv"
	"log"
	"os"
	"strings"
	"time"
	// "github.com/charmbracelet/bubbles/viewport"
)

const (
	footerHeight = 1
	headerHeight = 0

	focusedTextColor = "205"
)

var (
	color               = te.ColorProfile().Color
	focusedPrompt       = te.String(": ").Foreground(color("205")).String()
	blurredPrompt       = ": "
	focusedSubmitButton = "[ " + te.String("Submit").Foreground(color("205")).String() + " ]"
	blurredSubmitButton = "[ " + te.String("Submit").Foreground(color("240")).String() + " ]"
)

// type guessesModel struct {
// 	guessesFocus int
// 	guesses      []string
// }

type transModel struct {
	word WordMsg
}

func (m transModel) View() string {
	var (
		transtext    string
		phrasetext   string
		sentencetext string
	)

	for _, t := range m.word.Word.Translates {
		transtext += fmt.Sprintf("%s %s\n", t.Mean, t.Part)
	}
	for _, p := range m.word.Phrases {
		phrasetext += fmt.Sprintf("%s: %s\n", p.Text, p.Trans)
	}
	for _, s := range m.word.Sentences {
		sentencetext += fmt.Sprintf("%s: %s\n", s.Text, s.Trans)
	}

	return strings.Join([]string{
		"\n" + transtext,
		phrasetext,
		sentencetext,
	}, "\n")
}

type DictModel struct {
	cli                 DictClient
	config              *idictconfig.Config
	text                string
	ready               bool
	textInput           textinput.Model
	viewport            viewport.Model
	transmodel          transModel
	width               int
	textInputLeftMargin int
	// guesses   guessesModel
}

type WordMsg struct {
	Word      Word
	Phrases   []Phrase
	Sentences []Sentence
}

func Fetch(cli DictClient, text string) func() tea.Msg {
	return func() tea.Msg {
		_, word, phrases, sentences := cli.Fetch(text)
		return WordMsg{Word: word, Phrases: phrases, Sentences: sentences}
	}
}

func (m DictModel) Init() tea.Cmd {
	if m.text == "" {
		return textinput.Blink
	}
	m.textInput.SetValue(m.text)
	return tea.Batch(textinput.Blink, Fetch(m.cli, m.text))
}

func initialDictModel(text string) DictModel {
	m := DictModel{}
	m.cli = EuDictClient{}
	m.text = text
	m.textInput = textinput.NewModel()
	m.textInput.Placeholder = "Type to input"
	m.textInput.Focus()
	m.textInput.Prompt = focusedPrompt
	m.textInput.TextColor = focusedTextColor
	m.textInput.CharLimit = 300
	m.textInput.Width = 60
	m.textInput.SetValue(text)
	m.textInput.SetCursor(len(text))
	m.textInputLeftMargin = 0
	return m
}

func (m DictModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c":
			return m, tea.Quit

		// Cycle between inputs
		// case "tab", "shift+tab", "enter", "up", "down":
		case "enter":
			fallthrough
		case "esc":
			m.textInput.Blur()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		viewportHeight := msg.Height - headerHeight - footerHeight - 1
		if !m.ready {
			m.viewport = viewport.Model{Width: msg.Width, Height: viewportHeight}
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}
		m.textInput.Blur()
	case WordMsg:
		log.Println(msg)
		m.transmodel.word = msg
		m.viewport.SetContent(strings.Join(
			[]string{
				m.transmodel.View(),
			}, "",
		))
	}

	// Handle character input and blinks
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	if !m.textInput.Focused() {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m DictModel) View() string {
	if !m.ready {
		return "\n  Initalizing..."
	}
	if m.width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.width)
	}

	return strings.Join(
		[]string{
			strings.Repeat("\n", headerHeight),
			strings.Repeat(" ", m.textInputLeftMargin) + m.textInput.View(),
			m.viewport.View(),
			footer(m.viewport.Width),
		},
		"\n",
	)
}

func footer(width int) string {

	if width < 80 {
		return ui.StyleLogo(" idict ")
	}

	t := time.Now()
	tstr := fmt.Sprintf("%s %02d:%02d:%02d", t.Weekday().String(), t.Hour(), t.Minute(), t.Second())

	return ui.Line(
		width,
		ui.Cell{
			Width: 10,
			Text:  ui.StyleLogo(" idict "),
		},
		ui.Cell{
			Width: 50,
			Text:  ui.StyleHelp("ctrl+c: exit esc: quit input enter: query"),
		},
		ui.Cell{
			// Text:  ui.StyleHelp("⟳  " + tstr),
			Text:  ui.StyleHelp(tstr),
			Align: ui.RightAlign,
		},
	)

}

func Start(config *idictconfig.Config) func(string) error {
	return func(text string) error {
		if err := tea.NewProgram(initialDictModel(text)).Start(); err != nil {
			fmt.Printf("could not start program: %s\n", err)
			os.Exit(1)
		}
		return nil
	}
}

func logfile(v interface{}) {
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(v)
}
