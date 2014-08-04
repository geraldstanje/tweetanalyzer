package classifier

import (
  "fmt"
  "testing"
)

func TestTokenize(t *testing.T) {
  tokenizer := NewTokenizer()
  text := "Gas by my house hit $3.39!!!! I'm going to Chapel Hill on Sat. :)"
  tokens := tokenizer.Tokenize(text)
  
  x := fmt.Sprintf("%v", tokens)
  if x != `[gas by my house hit im going to chapel hill on sat :)]` {
    t.Fatal(x)
  }
}