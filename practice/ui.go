package practice

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	idictconfig "github.com/lai323/idict/config"
	"github.com/lai323/idict/dict"
	"github.com/lai323/idict/ui"
	"github.com/lai323/idict/utils"
	"github.com/lai323/idict/wordset"
	"github.com/muesli/reflow/wordwrap"
)

type NextMsg struct {
	word wordset.Word
}

func init() {
	rand.Seed(time.Now().Unix())
}

func getWords(worsetName string, config *idictconfig.Config) (map[string]int, error) {
	ws, err := wordset.NewWordSet(worsetName, wordset.WordSetManage{StoragePath: config.StoragePath}.WordSetDir())
	if err != nil {
		return ws.Words, err
	}
	exist, err := ws.Exist()
	if err != nil {
		return ws.Words, err
	}
	if !exist {
		return ws.Words, fmt.Errorf("WrodSet %s not exist", worsetName)
	}
	err = ws.Load()
	return ws.Words, nil
}

func initialModel(worsetName string, config *idictconfig.Config) (*PracModel, error) {
	m := &PracModel{config: config}

	pefile := path.Join(config.StoragePath, "practice_extent.json")
	pe, err := NewPracExtent(pefile, config.RestudyInterval)
	if err != nil {
		return m, err
	}

	words, err := getWords(worsetName, config)
	if err != nil {
		return m, err
	}

	wordNumTotal := len(words)
	for _, w := range pe.RememberWords() {
		delete(words, w)
	}
	wordslice := []string{}
	for w := range words {
		wordslice = append(wordslice, w)
	}

	cli, err := dict.NewEuDictClient(*config)
	if err != nil {
		return m, err
	}

	m.config = config
	m.cli = cli
	m.Words = wordslice
	m.wordNumTotal = wordNumTotal
	m.pracExtent = pe
	m.textInput = textinput.NewModel()
	m.textInput.Placeholder = "Type to input"
	m.textInput.Focus()
	m.textInput.Prompt = ""
	m.textInput.TextColor = ui.InputTextColor
	m.textInput.CharLimit = 600
	m.textInput.Width = 60
	m.textInput.Placeholder = "__________"

	m.helpmode = ui.HelpModel{
		Keyhelp: [][]string{
			{"?", "back"},
			{"i", "active input"},
			{"a", "show answer"},
		},
	}
	return m, nil
}

type PracModel struct {
	config              *idictconfig.Config
	textInput           textinput.Model
	viewport            viewport.Model
	cli                 dict.DictClient
	width               int
	ready               bool
	helpmode            ui.HelpModel
	viewportContent     string
	viewportContentLast string
	lastmodel           string
	answertext          string
	successed           bool
	failed              bool
	Words               []string
	wordNumTotal        int
	pracExtent          pracExtent
	sencursor           int
	batchWord           []string
	batchWordTotal      int
	batchWordCursor     int
	currentWord         wordset.Word
	showAnswer          bool
}

func (m *PracModel) Init() tea.Cmd {
	cmds := m.next()
	cmds = append(cmds, textinput.Blink)
	return tea.Batch(cmds...)
}

func (m *PracModel) genBatchWord() {
	words := m.pracExtent.ReviewWords()
	if len(words) >= m.config.GroupNum {
		words = words[:m.config.GroupNum]
	} else {
		cursor := m.config.GroupNum - len(words)
		if cursor > len(m.Words) {
			cursor = len(m.Words)
		}

		words = append(words, m.Words[:cursor]...)
		m.Words = m.Words[cursor:]
	}
	m.batchWord = words
	m.batchWordTotal = len(words)
	m.batchWordCursor = 0
}

func (m *PracModel) next() []tea.Cmd {
	cmds := []tea.Cmd{}
	if len(m.batchWord) == 0 {
		m.genBatchWord()
	}
	if len(m.batchWord) == 0 {
		cmds = append(cmds, tea.Quit)
		return cmds
	}

	voicecmd := m.config.FfplayPath
	voicearg := m.config.FfplayArgs
	wordtxet := m.batchWord[0]

	if voicecmd != "" && wordtxet != "" {
		cmds = append(cmds, func() tea.Msg {
			text := "QYN" + base64.StdEncoding.EncodeToString([]byte(wordtxet))
			url := `https://api.frdic.com/api/v2/speech/speakweb?langid=en&voicename=en_us_female&txt=` + text
			voicearg = append(voicearg, url)
			cmd := exec.Command(voicecmd, voicearg...)
			cmd.Run()
			return dict.VoiceMsg{}
		})
	}

	cmds = append(cmds, func() tea.Msg {
		err, word := m.cli.FetchCache(wordtxet)
		if err != nil {
			panic(err)
		}
		return NextMsg{word: word}
	})
	return cmds
}

func (m *PracModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "i", "backspace":
			if !m.textInput.Focused() && m.successed != true {
				// 需要重置 viewport 因为联想内容和当前内容高度不同
				m.viewport.GotoTop()
				m.textInput.Focus()
				return m, tea.Batch(cmds...)
			}
			m.failed = false
		case "enter":
			if m.successed != true {
				m.answer()
			} else {
				cmds = append(cmds, m.next()...)
			}
		case "a":
			if !m.textInput.Focused() && m.helpmode.Active == false {
				m.showAnswer = true
			}
		case "esc":
			m.textInput.Blur()
		case "ctrl+c":
			return m, tea.Quit
		case "?":
			if !m.textInput.Focused() {
				cmds = append(cmds, m.helpCmd())
			}
		default:
			m.failed = false
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		viewportHeight := msg.Height - 2 // footer 占一行 infobar 占一行
		if !m.ready {
			m.viewport = viewport.Model{Width: msg.Width, Height: viewportHeight}
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}
		m.viewport.SetContent(wordwrap.String(m.viewportContent, m.viewport.Width))
	case ui.HelpMsg:
		m.updatehelp()
	case NextMsg:
		m.successed = false
		m.failed = false
		m.textInput.Focus()
		m.answertext = ""
		m.currentWord = msg.word
		m.sencursor = 0
		m.showAnswer = false
		m.batchWordCursor += 1
		m.textInput.SetValue("")
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

func (m *PracModel) answer() {
	m.answertext = strings.TrimSpace(strings.ToLower(m.textInput.Value()))
	if m.answertext == strings.ToLower(m.currentWord.Text) {
		m.successed = true
		m.textInput.Blur()
		m.batchWord = m.batchWord[1:]
		err := m.pracExtent.Remember(m.currentWord.Text)
		if err != nil {
			panic(err)
		}
	} else {
		m.failed = true
		err := m.pracExtent.Forget(m.currentWord.Text)
		if err != nil {
			panic(err)
		}
	}
}

func (m *PracModel) helpCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.HelpMsg{}
	}
}

func (m *PracModel) updatehelp() {
	if m.helpmode.Active {
		m.helpmode.Active = false
		m.viewportContent = m.viewportContentLast
		m.viewport.SetContent(wordwrap.String(m.viewportContent, m.viewport.Width))
		return
	}
	m.helpmode.Active = true
	m.viewportContentLast = m.viewportContent
	m.viewportContent = m.helpmode.View()
	m.viewport.SetContent(wordwrap.String(m.viewportContent, m.viewport.Width))
}

func (m *PracModel) View() string {
	if m.helpmode.Active {
		return m.helpmode.View()
	}
	return m.PracView()
}

func (m *PracModel) PracView() string {
	if !m.ready || m.currentWord.Text == "" {
		return "\n  Initalizing..."
	}
	if m.width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.width)
	}

	wordtrans := ""
	for _, t := range m.currentWord.Translates {
		// 有时候翻译中会有这个单词的其他时态，检查一下避免显示答案
		transtext := strings.ToLower(utils.SpaceMap(t.Mean + t.Part))
		if strings.Contains(transtext, m.currentWord.Text) {
			continue
		}
		if strings.Contains(transtext, "时态") {
			continue
		}
		wordtrans += fmt.Sprintf("%s %s\n", ui.StyleMean(t.Mean), ui.StylePart(t.Part))
	}

	endstr := ""
	startstr := ""
	transtr := ""
	if len(m.currentWord.Sentences) != 0 {
		if m.sencursor == 0 {
			m.sencursor = rand.Intn(len(m.currentWord.Sentences))
		}

		sen := m.currentWord.Sentences[m.sencursor]
		senslice := strings.Split(strings.ToLower(sen.Text), strings.ToLower(sen.Word))
		if len(senslice) > 1 {
			endstr = strings.Join(senslice[1:], sen.Word)
			startstr = senslice[0] + " "
			transtr = sen.Trans
		}
	}

	transtr = transtr + "\n\n\n\n" + wordtrans

	answerstr := ""
	if m.showAnswer || m.successed {
		answermodel := ui.TransModel{Word: m.currentWord}
		answerstr = strings.Join(
			[]string{
				ui.StyleMean(m.currentWord.Text), "\n",
				answermodel.View(),
			},
			"",
		)
		// m.textInput.Focus()
		// m.textInput.SetValue(m.currentWord.Text)
	}

	if m.successed {
		endstr = endstr + ui.StyleSuccess("  √")
	}
	if m.failed {
		endstr = endstr + ui.Stylefail("  X")
	}

	m.viewportContent = strings.Join(
		[]string{
			"\n",
			"\n",
			startstr, m.textInput.View(), endstr,
			"\n",
			"\n",
			transtr,
			"\n",
			"\n",
			answerstr,
		},
		"",
	)
	m.viewport.SetContent(wordwrap.String(m.viewportContent, m.viewport.Width))

	infobarstr := infobar(
		m.wordNumTotal,
		len(m.pracExtent.RememberWords()),
		len(m.pracExtent.ReviewWords()),
		m.batchWordTotal,
		m.batchWordCursor,
		m.pracExtent.CorrectNum(m.currentWord.Text),
		m.viewport.Width,
	)
	return strings.Join(
		[]string{
			m.viewport.View(), "\n",
			infobarstr, "\n",
			ui.Footer(m.viewport.Width),
		},
		"",
	)
}

func infobar(total, remember, toreview, group, groupcursor, correct, width int) string {
	totaltext := "total " + strconv.Itoa(total)
	remembertext := "remember " + strconv.Itoa(remember)
	toreviewtext := "to review " + strconv.Itoa(toreview)
	grouptext := "group " + strconv.Itoa(group) + "/" + strconv.Itoa(groupcursor)
	correcttext := "word correct " + strconv.Itoa(correct)
	return ui.Line(
		width,
		ui.Cell{
			Width: len(totaltext) + 2,
			Text:  ui.StyleWordCount(totaltext),
		},
		ui.Cell{
			Width: len(remembertext) + 2,
			Text:  ui.StyleWordCount(remembertext),
		},
		ui.Cell{
			Width: len(toreviewtext) + 2,
			Text:  ui.StyleWordCount(toreviewtext),
		},
		ui.Cell{
			Width: len(correcttext) + 2,
			Text:  ui.StyleWordCount(correcttext),
		},
		ui.Cell{
			Text:  ui.StyleWordCount(grouptext),
			Align: ui.RightAlign,
		},
	)
}

func Start(config *idictconfig.Config) func(string) error {
	return func(worset string) error {
		m, err := initialModel(worset, config)
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
