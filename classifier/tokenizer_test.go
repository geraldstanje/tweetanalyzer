package classifier

import (
  "fmt"
  "testing"
)

func TestTokenize(t *testing.T) {
  tokenizer := NewTokenizer()
  text := "The quick brown fox was jumping over the 2 lazy dogs... #crazyfox @thedog http://fox.com :) \\u2019, iam great."
  tokens := tokenizer.Tokenize(text)

  x := fmt.Sprintf("%v", tokens)
  if x != `[quick brown fox jump over lazi dog crazyfox __h__ iam great]` {
    t.Fatal(x)
  }
}