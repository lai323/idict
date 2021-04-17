package dict

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"testing"
)

func TestEuDictGuess(t *testing.T) {
	cli := EuDictClient{}
	words, err := cli.Guess("ap")
	fmt.Println(words, err)
}

func TestEuDictFetch(t *testing.T) {
	cli := EuDictClient{}
	err, word := cli.Fetch("guess")
	fmt.Println(err, word)
}

func TestEuquerySentence(t *testing.T) {
	resp, _ := euquerySentence("function over from car look pretty red")
	// resp, _ := euquerySentence("Our department has engaged a foreign teacher as phonetic adviser")

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", bodyText)
}

func testurl() {
	fmt.Println(url.QueryEscape("123 123 text/plain"))
	params := url.Values{}
	params.Add("name", "@Rajeev")
	params.Add("phone", "+919999999999")
	fmt.Println(params.Encode())
}
