package dict

import (
	// "bytes"
	"encoding/json"
	"fmt"
	// "io"
	"io/ioutil"
	"net/http"
	"net/url"
	// "os"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/lai323/idict/utils"
	"github.com/muesli/termenv"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var term = termenv.ColorProfile()

type Translate struct {
	Part string
	Mean string
}

type Pronounce struct {
	Phonetic string
	Voice    string
}

type Word struct {
	PronounceUS Pronounce
	PronounceUK Pronounce
	Text        string
	Translates  []Translate
}

type GuessWord struct {
	Label      string
	Value      string
	Hlword     string
	Hlwordshow string
}

type Phrase struct {
	Word  string
	Text  string
	Trans string
}

type Sentence struct {
	Word  string
	Text  string
	Trans string
}

type DictClient interface {
	Fetch(string) (error, Word, []Phrase, []Sentence)
	Guess(string) (error, []GuessWord)
	StorageWords([]Word) error
	StoragePhrases([]Phrase) error
	StorageSentences([]Sentence) error
}

type EuDictClient struct {
}

func (d EuDictClient) StorageWords(words []Word) error {
	return nil
}
func (d EuDictClient) StoragePhrases(phrases []Phrase) error {
	return nil
}
func (d EuDictClient) StorageSentences(sentences []Sentence) error {
	return nil
}
func (d EuDictClient) Fetch(text string) (error, Word, []Phrase, []Sentence) {
	var (
		err       error
		word      Word
		phrases   []Phrase
		sentences []Sentence
	)
	if text == "" {
		return err, word, phrases, sentences
	}

	word.Text = text
	resp, err := euquery(text)
	if err != nil {
		return err, word, phrases, sentences
	}
	defer resp.Body.Close()

	// respfile, err := os.Create("tmp.html")
	// if err != nil {
	// 	return err, word, phrases, sentences
	// }
	// defer respfile.Close()
	// var buf bytes.Buffer
	// tee := io.TeeReader(resp.Body, &buf)
	// if err != nil {
	// 	return err, word, phrases, sentences
	// }
	// _, err = io.Copy(respfile, tee)
	// if err != nil {
	// 	return err, word, phrases, sentences
	// }
	// r, err := charset.NewReader(&buf, resp.Header.Get("Content-Type"))

	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return err, word, phrases, sentences
	}
	doc, err := html.Parse(r)
	if err != nil {
		return err, word, phrases, sentences
	}

	pronouncelist := htmlquery.Find(doc, `//span[@class="Phonitic"]`)
	if len(pronouncelist) != 0 {
		phoneticUK := htmlquery.InnerText(htmlquery.FindOne(pronouncelist[0], `./text()`))
		phoneticUS := htmlquery.InnerText(htmlquery.FindOne(pronouncelist[1], `./text()`))
		word.PronounceUK = Pronounce{Phonetic: phoneticUK}
		word.PronounceUS = Pronounce{Phonetic: phoneticUS}
	}

	translatelist := htmlquery.Find(doc, `//div[@id="ExpFCChild"]/ol/li`)
	for _, trans := range translatelist {
		var (
			transText  string
			transiText string
		)
		transi := htmlquery.Find(trans, `./i`)
		if len(transi) != 0 {
			transiText = htmlquery.InnerText(transi[0])
		}
		transText = htmlquery.InnerText(htmlquery.FindOne(trans, `./text()`))
		transiText = strings.TrimSpace(transiText)
		transText = strings.TrimSpace(transText)
		if transText == "" {
			continue
		}
		word.Translates = append(word.Translates, Translate{Part: transiText, Mean: transText})
	}

	translatelist1 := htmlquery.Find(doc, `//div[@id="ExpFCChild"]//div[@class="exp"]|//div[@id="ExpFCChild"]//div[@class="exp"]/ol/li`)
	for _, trans := range translatelist1 {
		node :=  htmlquery.FindOne(trans, `./text()`)
		if node != nil {
			transText := htmlquery.InnerText(node)
			transText = strings.TrimSpace(transText)
			if transText == "" {
				continue
			}
			word.Translates = append(word.Translates, Translate{Mean: transText})
		}
	}

	translatelist2 := htmlquery.Find(doc, `//div[@id="ExpFCChild"]//div[@id="trans"]`)
	for _, trans := range translatelist2 {
		spans := htmlquery.Find(trans, `./span`)
		var transText string
		for _, span := range spans {
			transText += strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(span, `./text()`)))
		}
		if transText == "" {
			continue
		}
		word.Translates = append(word.Translates, Translate{Mean: transText})
	}

	// {
	// 	// 翻译句子时 参考译文的部分
	// 	sentenceTrans := htmlquery.FindOne(doc, `//div[@id="tbTransResult"]/span`)
	// 	logfile(sentenceTrans)
	// 	transText := htmlquery.InnerText(sentenceTrans)
	// 	transText = strings.TrimSpace(transText)
	// 	if transText != ""{
	// 		word.Translates = append(word.Translates, Translate{Mean: transText})
	// 	}
	// }
	if len(word.Translates) == 0 {
		resp, err := euquerySentence(text)
		if err != nil {
			return err, word, phrases, sentences
		}
		defer resp.Body.Close()

		transb, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err, word, phrases, sentences
		}
		transText := string(transb)
		if transText != "" {
			word.Translates = append(word.Translates, Translate{Mean: transText})
		}
	}

	phraseslist := htmlquery.Find(doc, `//div[@id="phrase"]`)
	for _, pdiv := range phraseslist {
		var (
			phrasetext  string
			phrasetrans string
		)
		phrasetext = htmlquery.InnerText(htmlquery.FindOne(pdiv, `./i/text()`))
		phrasetrans = htmlquery.InnerText(htmlquery.FindOne(pdiv, `./span/text()`))
		// fmt.Println(phrasetext, phrasetrans)
		phrases = append(phrases, Phrase{Word: text, Text: phrasetext, Trans: phrasetrans})
	}
	sentenceslist := htmlquery.Find(doc, `// div[@class="lj_item"]/div[@class="content"]`)
	for _, sdiv := range sentenceslist {
		var (
			senttext  string
			senttrans string
		)
		senttext = htmlquery.InnerText(htmlquery.FindOne(sdiv, `./p[@class="line"]`))
		senttrans = htmlquery.InnerText(htmlquery.FindOne(sdiv, `./p[@class="exp"]`))
		// fmt.Println(senttext, senttrans)
		sentences = append(sentences, Sentence{Word: text, Text: senttext, Trans: senttrans})
	}

	return err, word, phrases, sentences
}

func (d EuDictClient) Guess(text string) (error, []GuessWord) {
	var (
		guesses []GuessWord
		err     error
	)
	if text == "" {
		return err, guesses
	}
	resp, err := http.Get("https://dict.eudic.net/dicts/prefix/" + text)
	if err != nil {
		return utils.FmtErrorf("guess word error", err), guesses
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return utils.FmtErrorf("guess word error", err), guesses
	}
	err = json.Unmarshal(body, &guesses)
	if err != nil {
		return utils.FmtErrorf("guess word error", err), guesses
	}
	for _, w := range guesses {
		bold := getBold(w.Hlword)
		oldBold := fmt.Sprintf("<b>%s</b>", bold)
		newBold := termenv.String("\n  ↑/↓: Navigate • q: Quit\n").Foreground(term.Color("241")).String()
		w.Hlwordshow = strings.Replace(w.Hlword, oldBold, newBold, -1)
	}
	return nil, guesses
}

var bold = regexp.MustCompile(`<b>(.*)</b>`)

func getBold(msg string) string {
	ret := bold.FindStringSubmatch(msg)
	if len(ret) == 2 {
		return ret[1]
	}
	return ""
}

func euquery(text string) (*http.Response, error) {
	var (
		err  error
		resp *http.Response
	)
	url := fmt.Sprintf("https://dict.eudic.net/dicts/en/%s", text)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return resp, err
	}
	req.Header.Set("authority", "dict.eudic.net")
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("sec-ch-ua", `Chromium";v="88", "Google Chrome";v="88", ";Not A Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("sec-fetch-dest", "document")
	// req.Header.Set("referer", "https://dict.eudic.net/dicts/en/illuminate")
	req.Header.Set("referer", "https://dict.eudic.net/dicts/")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	return client.Do(req)
}

func euquerySentence(text string) (*http.Response, error) {
	var (
		err  error
		resp *http.Response
	)

	params := url.Values{}
	params.Add("to", "zh-CN")
	params.Add("from", "en")
	params.Add("text", text)
	params.Add("contentType", "text/plain")
	bodystr := params.Encode()
	var body = strings.NewReader(bodystr)

	url := "https://dict.eudic.net/Home/TranslationAjax"
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return resp, err
	}
	req.Header.Set("authority", "dict.eudic.net")
	req.Header.Set("sec-ch-ua", `Chromium";v="88", "Google Chrome";v="88", ";Not A Brand";v="99"`)
	req.Header.Set("accept", "*/*")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.190 Safari/537.36")
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("origin", "https://dict.eudic.net")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("referer", "https://dict.eudic.net/dicts/")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	return client.Do(req)
}
