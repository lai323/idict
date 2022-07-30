package practice

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lai323/idict/config"
	"github.com/lai323/idict/db"
	"github.com/lai323/idict/dict"
	"github.com/lai323/idict/ui"
	"github.com/lai323/idict/utils"
	"github.com/muesli/reflow/wordwrap"
)

type NextMsg struct {
	word db.Word
}

func init() {
	rand.Seed(time.Now().Unix())
}

func initialModel(collection string, cfg *config.Config, dictdb db.DictDB, shuffle bool) (*PracModel, error) {
	rememberCount := cfg.RememberCount()
	if rememberCount == 0 {
		return nil, fmt.Errorf("Practice rememberCount can not be 0")
	}
	m := &PracModel{}

	collectionWords, err := dictdb.CollectionWords(collection)
	if err != nil {
		return m, err
	}
	total := len(collectionWords)

	words := map[string]*db.WordPractice{}
	wordsOrder := []string{}
	rememberedWords := []string{}
	for _, w := range collectionWords {
		info, err := dictdb.PracticeGet(w)
		if err != nil {
			return nil, err
		}
		if info.Degree >= rememberCount {
			rememberedWords = append(rememberedWords, info.Text)
			continue
		}
		words[info.Text] = &info
		wordsOrder = append(wordsOrder, info.Text)
	}
	if shuffle {
		rand.Shuffle(len(wordsOrder), func(i, j int) {
			wordsOrder[i], wordsOrder[j] = wordsOrder[j], wordsOrder[i]
		})
	}

	restudyIntervals := []restudyInterval{}
	for degree, hour := range cfg.RestudyInterval {
		if hour == -1 {
			continue
		}
		restudyIntervals = append(restudyIntervals, restudyInterval{Degree: degree, Hour: hour})
	}
	sort.Slice(restudyIntervals, func(i, j int) bool {
		return restudyIntervals[i].Degree > restudyIntervals[j].Degree
	})

	cli, err := dict.NewEuDictClient(dictdb)
	if err != nil {
		return m, err
	}

	m.cli = cli
	m.dictdb = dictdb
	m.words = words
	m.wordsOrder = wordsOrder
	m.rememberedWords = rememberedWords
	m.total = total
	m.textInput = textinput.NewModel()
	m.textInput.Placeholder = "Type to input"
	m.textInput.Focus()
	m.textInput.Prompt = ""
	m.textInput.TextColor = ui.InputTextColor
	m.textInput.CharLimit = 600
	m.textInput.Width = 60
	m.textInput.Placeholder = "__________"
	m.ffplayPath = cfg.FfplayPath
	m.ffplayArgs = cfg.FfplayArgs
	m.groupNum = cfg.GroupNum
	m.restudyInterval = restudyIntervals
	m.rememberCount = rememberCount

	m.helpmode = ui.HelpModel{
		Keyhelp: [][]string{
			{"?", "back"},
			{"i", "active input"},
			{"a", "show answer"},
		},
	}
	return m, nil
}

type restudyInterval struct {
	Degree int
	Hour   int
}

type PracModel struct {
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

	dictdb          db.DictDB
	words           map[string]*db.WordPractice
	wordsOrder      []string
	rememberedWords []string
	total           int
	sencursor       int
	batchWord       []string
	batchWordTotal  int
	batchWordCursor int
	currentWord     db.Word
	showAnswer      bool
	ffplayPath      string
	ffplayArgs      []string
	restudyInterval []restudyInterval
	groupNum        int
	rememberCount   int
}

func (m *PracModel) Init() tea.Cmd {
	cmds := m.next()
	cmds = append(cmds, textinput.Blink)
	return tea.Batch(cmds...)
}

func (m *PracModel) remember(w string) error {
	p := m.words[w]
	p.Degree = p.Degree + 1
	p.LastTime = time.Now().Unix()
	if p.Degree >= m.rememberCount {
		m.rememberedWords = append(m.rememberedWords, w)
	}
	return m.dictdb.PracticePut(*p)
}

func (m *PracModel) forget(w string) error {
	p := m.words[w]
	p.Degree = 0
	p.LastTime = time.Now().Unix()
	return m.dictdb.PracticePut(*p)
}

func (m *PracModel) reviewWords() []string {
	words := []string{}
	now := time.Now().Unix()
	for _, info := range m.words {
		if info.Degree >= m.rememberCount || info.Degree == 0 {
			continue
		}
		match := false
		for _, r := range m.restudyInterval {
			if info.Degree >= r.Degree {
				match = true
				if info.LastTime+60*60*int64(r.Hour) < now {
					words = append(words, info.Text)
				}
				break
			}
		}
		if !match {
			words = append(words, info.Text)
		}
	}
	return words
}

func (m *PracModel) genBatchWord() {
	words := m.reviewWords()
	if len(words) >= m.groupNum {
		words = words[:m.groupNum]
	} else {
		cursor := m.groupNum - len(words)
		if cursor > len(m.wordsOrder) {
			cursor = len(m.wordsOrder)
		}

		words = append(words, m.wordsOrder[:cursor]...)
		m.wordsOrder = m.wordsOrder[cursor:]
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
	wordtxet := m.batchWord[0]

	if m.ffplayPath != "" && wordtxet != "" {
		cmds = append(cmds, func() tea.Msg {
			text := "QYN" + base64.StdEncoding.EncodeToString([]byte(wordtxet))
			url := `https://api.frdic.com/api/v2/speech/speakweb?langid=en&voicename=en_us_female&txt=` + text
			arg := append(m.ffplayArgs, url)
			cmd := exec.Command(m.ffplayPath, arg...)
			cmd.Run()
			return dict.VoiceMsg{}
		})
	}

	cmds = append(cmds, func() tea.Msg {
		err, word := m.cli.Fetch(wordtxet)
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
	m.answertext = strings.TrimSpace(m.textInput.Value())
	if strings.ToLower(m.answertext) == strings.ToLower(m.currentWord.Text) {
		m.successed = true
		m.textInput.Blur()
		m.batchWord = m.batchWord[1:]
		err := m.remember(m.currentWord.Text)
		if err != nil {
			panic(err)
		}
	} else {
		m.failed = true
		err := m.forget(m.currentWord.Text)
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
		m.total,
		len(m.rememberedWords),
		len(m.reviewWords()),
		m.batchWordTotal,
		m.batchWordCursor,
		m.words[m.currentWord.Text].Degree,
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

func Start(cfg *config.Config, worset string, dictdb db.DictDB, shuffle bool) {
	m, err := initialModel(worset, cfg, dictdb, shuffle)
	if err != nil {
		fmt.Printf("could not start program: %s\n", err)
		dictdb.Close()
		os.Exit(1)
	}
	if err := tea.NewProgram(m).Start(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		dictdb.Close()
		os.Exit(1)
	}
	dictdb.Close()
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
