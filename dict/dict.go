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
	"github.com/lai323/idict/config"
	"github.com/lai323/idict/wordset"
	"github.com/muesli/termenv"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var term = termenv.ColorProfile()

type DictClient interface {
	Fetch(string) (error, wordset.Word)
	Guess(string) (error, []wordset.GuessWord)
	FetchCache(string) (error, wordset.Word)
}

type EuDictClient struct {
	config         config.Config
	wordcache      wordset.WordCache
	defaultWordset wordset.WordSet
}

func NewEuDictClient(config config.Config) (EuDictClient, error) {
	cli := EuDictClient{}
	defaultWordset, err := wordset.NewWordSet("default", wordset.WordSetManage{StoragePath: config.StoragePath}.WordSetDir())
	if err != nil {
		return cli, err
	}

	err = defaultWordset.Load()
	if err != nil {
		return cli, err
	}

	wordcache, err := wordset.NewWordCache(config.StoragePath)
	if err != nil {
		return cli, err
	}
	cli.config = config
	cli.wordcache = wordcache
	cli.defaultWordset = defaultWordset
	return cli, err
}

func (d EuDictClient) FetchCache(text string) (error, wordset.Word) {
	word, exist, err := d.wordcache.Get(text)
	if err != nil {
		return err, word
	}
	if err != nil {
		return err, word
	}
	if exist {
		if word.PronounceUS.Phonetic != "" {
			err = d.defaultWordset.Append(word.Text)
			if err != nil {
				return err, word
			}
		}
		return nil, word
	}

	err, word = d.Fetch(text)
	if err != nil {
		return err, word
	}
	if word.PronounceUS.Phonetic != "" {
		err = d.wordcache.Set(word)
		if err != nil {
			return err, word
		}
		err = d.defaultWordset.Append(word.Text)
		if err != nil {
			return err, word
		}
	}
	return err, word
}

func (d EuDictClient) Fetch(text string) (error, wordset.Word) {
	var (
		err       error
		word      wordset.Word
		phrases   []wordset.Phrase
		sentences []wordset.Sentence
	)
	if text == "" {
		return err, word
	}

	word.Text = strings.TrimSpace(strings.ToLower(text))
	resp, err := euquery(text)
	if err != nil {
		return err, word
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
		return err, word
	}
	doc, err := html.Parse(r)
	if err != nil {
		return err, word
	}

	pronouncelist := htmlquery.Find(doc, `//span[@class="Phonitic"]`)
	if len(pronouncelist) != 0 {
		phoneticUK := htmlquery.InnerText(htmlquery.FindOne(pronouncelist[0], `./text()`))
		if len(pronouncelist) == 2 {
			phoneticUS := htmlquery.InnerText(htmlquery.FindOne(pronouncelist[1], `./text()`))
			word.PronounceUS = wordset.Pronounce{Phonetic: phoneticUS}
		}
		word.PronounceUK = wordset.Pronounce{Phonetic: phoneticUK}
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
		word.Translates = append(word.Translates, wordset.Translate{Part: transiText, Mean: transText})
	}

	translatelist1 := htmlquery.Find(doc, `//div[@id="ExpFCChild"]//div[@class="exp"]|//div[@id="ExpFCChild"]//div[@class="exp"]/ol/li`)
	for _, trans := range translatelist1 {
		node := htmlquery.FindOne(trans, `./text()`)
		if node != nil {
			transText := htmlquery.InnerText(node)
			transText = strings.TrimSpace(transText)
			if transText == "" {
				continue
			}
			word.Translates = append(word.Translates, wordset.Translate{Mean: transText})
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
		word.Translates = append(word.Translates, wordset.Translate{Mean: transText})
	}

	translatelist3 := htmlquery.FindOne(doc, `//div[@id="ExpFCChild"]`)
	if translatelist3 != nil {

		var (
			transText  string
			transiText string
		)
		transnode := htmlquery.FindOne(translatelist3, `./text()`)
		transinode := htmlquery.FindOne(translatelist3, `./i/text()`)
		if transnode != nil {
			transText = strings.TrimSpace(htmlquery.InnerText(transnode))
		}
		if transinode != nil {
			transiText = strings.TrimSpace(htmlquery.InnerText(transinode))

		}
		word.Translates = append(word.Translates, wordset.Translate{Part: transiText, Mean: transText})
	}

	// {
	// 	// 翻译句子时 参考译文的部分
	// 	sentenceTrans := htmlquery.FindOne(doc, `//div[@id="tbTransResult"]/span`)
	// 	transText := htmlquery.InnerText(sentenceTrans)
	// 	transText = strings.TrimSpace(transText)
	// 	if transText != ""{
	// 		word.Translates = append(word.Translates, Translate{Mean: transText})
	// 	}
	// }
	if len(word.Translates) == 0 {
		resp, err := euquerySentence(text)
		if err != nil {
			return err, word
		}
		defer resp.Body.Close()

		transb, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err, word
		}
		transText := string(transb)
		if transText != "" {
			word.Translates = append(word.Translates, wordset.Translate{Mean: transText})
		}
	}

	phraseslist := htmlquery.Find(doc, `//div[@id="phrase"]`)
	for _, pdiv := range phraseslist {
		var (
			phrasetext  string
			phrasetrans string
		)
		itext := htmlquery.FindOne(pdiv, `./i/text()`)
		if itext == nil {
			continue
		}
		phrasetext = htmlquery.InnerText(itext)
		phrasetrans = htmlquery.InnerText(htmlquery.FindOne(pdiv, `./span/text()`))
		// fmt.Println(phrasetext, phrasetrans)
		phrases = append(phrases, wordset.Phrase{Word: text, Text: phrasetext, Trans: phrasetrans})
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
		sentences = append(sentences, wordset.Sentence{Word: text, Text: senttext, Trans: senttrans})
	}

	word.Phrases = phrases
	word.Sentences = sentences
	return err, word
}

func (d EuDictClient) Guess(text string) (error, []wordset.GuessWord) {
	var (
		guesses []wordset.GuessWord
		err     error
	)
	if text == "" {
		return err, guesses
	}
	text = strings.TrimSpace(text)
	resp, err := http.Get("https://dict.eudic.net/dicts/prefix/" + text)
	if err != nil {
		// 忽略这个异常
		// return utils.FmtErrorf("guess word error http get", err), guesses
		return nil, guesses
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 忽略这个异常
		// return utils.FmtErrorf("guess word error ioutil ReadAll", err), guesses
		return nil, guesses
	}
	err = json.Unmarshal(body, &guesses)
	if err != nil {
		// 忽略这个异常
		// return utils.FmtErrorf("guess word error json Unmarshal", err), guesses
		return nil, guesses
	}
	for i := range guesses {
		l := guesses[i].Label
		v := guesses[i].Value
		l = strings.Replace(l, "（", "(", -1)
		l = strings.Replace(l, "）", ")", -1)
		l = strings.Replace(l, "，", " ,", -1)
		// l = strings.Replace(l, "...", "", -1)
		guesses[i].Value = strings.TrimSpace(v)
		guesses[i].Label = strings.TrimSpace(l)
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
