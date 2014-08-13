package classifier

import (
  "fmt"
  "testing"
)

func TestTokenize(t *testing.T) {
  tokenizer := NewTokenizer()
  text := "The quick brown fox was jumping over the 2 lazy dogs #crazyfox @thedog http://fox.com :)"
  tokens := tokenizer.Tokenize(text)

  x := fmt.Sprintf("%v", tokens)
  if x != `[the quick brown fox was jumping over the lazy dogs crazyfox AT_USER URL __h__]` {
    t.Fatal(x)
  }
}