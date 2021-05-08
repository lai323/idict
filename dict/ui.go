package dict

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	idictconfig "github.com/lai323/idict/config"
	"github.com/lai323/idict/ui"
	"github.com/lai323/idict/wordset"
	"github.com/muesli/reflow/wordwrap"
	te "github.com/muesli/termenv"
)

var (
	focusedPrompt = te.String(": ").Foreground(te.ColorProfile().Color("205")).String()
)

type GuessMsg struct {
	words []wordset.GuessWord
}

type guessModel struct {
	words  GuessMsg
	active bool
	cursor int
}

func (m guessModel) View() string {
	var wordtext []string
	selected := m.cursor - 1
	for i, w := range m.words.words {
		style := ui.StyleGuessText
		if i == selected {
			style = ui.StyleGuessTextSelect
		}
		wordtext = append(wordtext,
			style(ui.Line(
				70,
				ui.Cell{
					// Width: ,
					Align: ui.LeftAlign,
					Text:  w.Value,
				},
				ui.Cell{
					Align: ui.LeftAlign,
					Text:  w.Label,
				},
			)))
	}
	return strings.Join(wordtext, "\n")
}

type VoiceMsg struct {
}

func initialDictModel(text string, config *idictconfig.Config) (DictModel, error) {
	m := DictModel{config: config}
	cli, err := NewEuDictClient(*config)
	if err != nil {
		return m, err
	}
	m.config = config
	m.cli = cli
	m.text = text
	m.textInput = textinput.NewModel()
	m.textInput.Placeholder = "Type to input"
	m.textInput.Focus()
	m.textInput.Prompt = focusedPrompt
	m.textInput.TextColor = ui.InputTextColor
	m.textInput.CharLimit = 600
	m.textInput.Width = 60
	m.textInput.SetValue(text)
	m.textInput.SetCursor(len(text))
	m.guessctx = context.Background()
	m.guessdelay = 500

	m.helpmode = ui.HelpModel{
		Keyhelp: [][]string{
			{"?", "back"},
			{"i", "active input"},
			{"v", "voice (if ffplay has been set)"},
			{"j", "up"},
			{"k", "down"},
			{"u", "page up"},
			{"d", "page down"},
		},
	}
	return m, nil
}

type DictModel struct {
	cli             DictClient
	config          *idictconfig.Config
	text            string
	ready           bool
	textInput       textinput.Model
	lastinputat     int64
	viewport        viewport.Model
	viewportContent string
	transmodel      ui.TransModel
	guessmodel      guessModel
	helpmode        ui.HelpModel
	guessctx        context.Context
	guessctxcancel  context.CancelFunc
	guessdelay      int64
	width           int
	lastmodel       string
}

func (m DictModel) Init() tea.Cmd {
	if m.text == "" {
		return textinput.Blink
	}
	m.textInput.SetValue(m.text)
	return tea.Batch(textinput.Blink, m.fetchCmd(m.text))
}

func (m *DictModel) fetchCmd(text string) func() tea.Msg {
	return func() tea.Msg {
		err, word := m.cli.FetchCache(text)
		if err != nil {
			panic(fmt.Errorf("fetch translate word error: %s", err.Error()))
		}
		return ui.WordMsg{word}
	}
}

func (m *DictModel) helpCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.HelpMsg{}
	}
}

func (m *DictModel) voiceCmd() tea.Cmd {
	voicecmd := m.config.FfplayPath
	voicearg := m.config.FfplayArgs
	word := m.transmodel.Word.Text
	if !(voicecmd != "" && word != "") {
		return nil
	}

	return func() tea.Msg {
		text := "QYN" + base64.StdEncoding.EncodeToString([]byte(word))
		url := `https://api.frdic.com/api/v2/speech/speakweb?langid=en&voicename=en_us_female&txt=` + text
		voicearg = append(voicearg, url)
		cmd := exec.Command(voicecmd, voicearg...)
		cmd.Run()
		return VoiceMsg{}
	}
}

func (m *DictModel) guessCmd() tea.Cmd {
	nowunix := UnixMillNow()
	if m.lastinputat == 0 {
		m.lastinputat = nowunix
		return nil
	}
	if m.guessctxcancel != nil && nowunix-m.lastinputat < m.guessdelay {
		m.guessctxcancel()
	}

	ctx, cancel := context.WithCancel(m.guessctx)
	m.guessctxcancel = cancel

	return func() tea.Msg {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Millisecond * time.Duration(m.guessdelay)):
				err, words := m.cli.Guess(m.textInput.Value())
				if err != nil {
					panic(fmt.Errorf("guess word error: %s", err.Error()))
				}
				return GuessMsg{words: words}
			}
		}
	}
}

func (m DictModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "i", "backspace":
			if !m.textInput.Focused() {
				// 需要重置 viewport 因为联想内容和当前内容高度不同
				m.viewport.GotoTop()
				m.textInput.Focus()
				return m, tea.Batch(cmds...)
			}
		case "enter":
			m.textInput.Blur()
			if m.guessmodel.cursor != 0 {
				text := m.guessmodel.words.words[m.guessmodel.cursor-1].Value
				m.textInput.SetValue(text)
				m.textInput.SetCursor(len(text))
				cmds = append(cmds, m.fetchCmd(text))
			} else {
				text := m.textInput.Value()
				if text != "" {
					cmds = append(cmds, m.fetchCmd(text))
				}
			}
		case "esc":
			m.textInput.Blur()
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.guessmodel.active {
				m.guessmodel.cursor -= 1
				if m.guessmodel.cursor <= 0 {
					m.guessmodel.cursor = len(m.guessmodel.words.words)
				}
				m.updateguess()
			}
		case "down":
			if m.guessmodel.active {
				m.guessmodel.cursor += 1
				if m.guessmodel.cursor > len(m.guessmodel.words.words) {
					m.guessmodel.cursor = 1
				}
				m.updateguess()
			}
		case "?":
			if !m.textInput.Focused() {
				cmds = append(cmds, m.helpCmd())
			}
		case "v":
			if !m.textInput.Focused() {
				cmd := m.voiceCmd()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
		if m.textInput.Focused() {
			cmds = append(cmds, m.guessCmd())
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		viewportHeight := msg.Height - 2 // input 占一行 footer 占一行
		if !m.ready {
			m.viewport = viewport.Model{Width: msg.Width, Height: viewportHeight}
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}
		m.viewport.SetContent(wordwrap.String(m.viewportContent, m.viewport.Width))
	case ui.WordMsg:
		m.guessmodel.active = false
		m.guessmodel.cursor = 0
		m.transmodel.Word = msg.Word
		m.updatetrans()
	case GuessMsg:
		if m.textInput.Focused() {
			m.guessmodel.words = msg
			m.guessmodel.active = true
			m.updateguess()
		}
	case ui.HelpMsg:
		m.updatehelp()
	case VoiceMsg:
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

func (m *DictModel) updatetrans() {
	m.lastmodel = "trans"
	m.viewportContent = m.transmodel.View()
	m.viewport.SetContent(wordwrap.String(m.viewportContent, m.viewport.Width))
}

func (m *DictModel) updateguess() {
	m.lastmodel = "guess"
	m.viewportContent = m.guessmodel.View()
	m.viewport.SetContent(wordwrap.String(m.viewportContent, m.viewport.Width))
}

func (m *DictModel) updatehelp() {
	if m.helpmode.Active {
		m.helpmode.Active = false
		if m.lastmodel == "guess" {
			m.updateguess()
			return
		}
		if m.lastmodel == "trans" {
			m.updatetrans()
			return
		}
		m.viewportContent = ""
		m.viewport.SetContent(wordwrap.String(m.viewportContent, m.viewport.Width))
		return
	}
	m.helpmode.Active = true
	m.viewportContent = m.helpmode.View()
	m.viewport.SetContent(wordwrap.String(m.viewportContent, m.viewport.Width))
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
			m.textInput.View(), "\n",
			m.viewport.View(), "\n",
			ui.Footer(m.viewport.Width),
		},
		"",
	)
}

func Start(config *idictconfig.Config) func(string) error {
	return func(text string) error {
		m, err := initialDictModel(text, config)
		if err != nil {
			fmt.Printf("could not start program: %s\n", err)
			os.Exit(1)
		}
		if err := tea.NewProgram(m).Start(); err != nil {
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

func UnixMillNow() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
