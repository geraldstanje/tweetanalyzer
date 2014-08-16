package classifier

import (
	"fmt"
	"testing"
)

func TestTokenize(t *testing.T) {
	tokenizer, err := NewTokenizer()
	if err != nil {
		t.Fatal(err)
	}

	text := "The quick brown fox was jumping over the 2 lazy dogs... #crazyfox @thedog http://fox.com :), The weather isn\\u2019t bad, I'm great."
	tokens := tokenizer.Tokenize(text)

	fmt.Println(tokens)

	x := fmt.Sprintf("%v", tokens)
	if x != `[quick brown fox jump over lazi dog __hashtag__ __user__ __url__ __h__ weather ist bad great]` {
		t.Fatal(x)
	}
}
