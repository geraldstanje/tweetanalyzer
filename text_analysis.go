package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Tokenizer struct {
}

func newEnglishTokenizer() *Tokenizer {
	return &Tokenizer{}
}

func (t *Tokenizer) tokenize(text string) []string {
	var result []string
	i := 0

	for {
		j := strings.Index(text[i:], " ")

		if j < 0 {
			break
		}

		result = append(result, text[i:i+j])
		i += j + 1
	}

	i = 0
	j := strings.Index(text, " ")
	if j < 0 {
		return nil
	}

	t.createBigrams(text, &result)

	return result
}

func (t *Tokenizer) createBigrams(text string, result *[]string) {
	i := 0
	j := strings.Index(text, " ")

	if j < 0 {
		return
	}
	j += 1
	for {
		k := strings.Index(text[j:], " ")
		if k < 0 {
			*result = append(*result, text[i:])
			break
		}
		*result = append(*result, text[i:j+k])
		i = j
		j += k + 1
	}
}

func stringToInt(str string) int {
	intVal, _ := strconv.ParseInt(str, 0, 64)
	return int(intVal)
}

func loadTrainingSet(filename string) (map[string]int, error) {
	dic := make(map[string]int)

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file ", err)
		return dic, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		s, err := r.ReadString('\n')
		if err == io.EOF {
			// do something here
			break
		} else if err != nil {
			return dic, err // if you return error
		} else {
			s = s[0 : len(s)-1] // remove '\n'
			word := strings.Split(s, "\t")
			dic[word[0]] = stringToInt(word[1])
		}
	}

	return dic, err
}

func analyseSentiment(sentence string, dic map[string]int) int {
	var val int

	tokenizer := newEnglishTokenizer()
	words := tokenizer.tokenize(sentence)

	for _, w := range words {
		val = val + dic[strings.ToLower(w)]
	}

	return val
}

func main() {
	dic, _ := loadTrainingSet("AFINN-111.txt")
	val := analyseSentiment("The weather is really good today", dic)

	if val > 0 {
		fmt.Println("positive")
	} else {
		fmt.Println("negative")
	}
}
