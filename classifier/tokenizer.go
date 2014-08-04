package classifier

import (
  "regexp"
  "strings"
)

var splitToken = regexp.MustCompile(`([A-Za-z]+|[:o*\-\]\[\)\(\}\{]{2,3})+`)

type Tokenizer struct {
}

func NewTokenizer() *Tokenizer {
  return &Tokenizer{}
}

func (t *Tokenizer) Tokenize(text string) ([]string) {  
  text = strings.ToLower(text)
  text = strings.Replace(text, "\\u2019", "", -1)
  text = strings.Replace(text, "\\u002c", ",", -1)
  text = strings.Replace(text, "'", "", -1)

  words := splitToken.FindAllString(text, -1)

  return words
}