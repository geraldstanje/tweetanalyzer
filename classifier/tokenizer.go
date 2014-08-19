package classifier

import (
	"bufio"
	"fmt"
	"github.com/reiver/go-porterstemmer"
	"github.com/sajari/fuzzy"
	"io"
	"os"
	"regexp"
	"strings"
)

// normalized tokens
const happyToken = "__h__"
const sadToken = "__s__"
const userToken = "__user__"
const hashtagToken = "__hashtag__"
const urlToken = "__url__"

type myRegexp struct {
	url           *regexp.Regexp
	user          *regexp.Regexp
	hashtag       *regexp.Regexp
	removeSpChars *regexp.Regexp
	removeWhitesp *regexp.Regexp
}

type Tokenizer struct {
	spellCorrect   *fuzzy.Model
	stopWords      Dict
	happyEmoticons map[string]bool
	sadEmoticons   map[string]bool
	negationsWords map[string]string
	myRegexp       myRegexp
}

func NewTokenizer() (*Tokenizer, error) {
	happy := `:-) :) ;) :o) :] :3 :c) :> =] 8) =) :} :^) :-d :d 8-d 8d x-d xd =-d =d =-3 =3 :-)) :'-) :') :* :^* >:p :-p x-p xp :p =p :-b :b >:) >;) >:-) <3 ;-) ;) ;-] ;] ;d ;^) >;) |;-)`
	sad := `>:[ :-( :( :-c :c :-< :< :-[ :[ :{ ;( :-|| :@ >:( :'-( :'( >:\\ >:/ :-/ :-. :\\ =/ =\\ :L =L :S >.< d; :-(( :(( ;-(( ;((`
	// http://sentiment.christopherpotts.net/lingstruc.html#negation
	// improvement
	negations := map[string]string{
		"don't":     "do not",
		"doesn't":   "do not",
		"didn't":    "do not",
		"won't":     "will not",
		"shouldn't": "will not",
		"wouldn't":  "will not",
		"isn't":     "is not",
		"aren't":    "is not",
		"wasn't":    "is not",
		"weren't":   "is not",
		"ain't":     "is not",
		"couldn't":  "can not",
		"can't":     "can not",
		"haven't":   "has not",
		"hasn't":    "has not",
		"hadn't":    "has not",
		"shan't":    "shall not",
		"mightn't":  "shall not",
		"mustn't":   "shall not",
		"mayn't":    "may not",
	}

	happyEmoticons := createMap(happy)
	sadEmoticons := createMap(sad)

	url := regexp.MustCompile(`((www\.[^\s]+)|(https?://[^\s]+))`)
	user := regexp.MustCompile(`@[^\s]+`)
	hashtag := regexp.MustCompile(`#([^\s]+)`)
	removeSpChars := regexp.MustCompile(`[^a-zA-Z_'-]+`)
	removeWhitesp := regexp.MustCompile(`[\s]+`)
	regexp := myRegexp{url: url, user: user, hashtag: hashtag, removeSpChars: removeSpChars, removeWhitesp: removeWhitesp}

	stopWords, err := createDict("data/stop_words.txt")
	if err != nil {
		return nil, err
	}

	spellCorrect := fuzzy.NewModel()
	spellCorrect.Train(fuzzy.SampleEnglish())
	spellCorrect.Threshold = 3

	return &Tokenizer{spellCorrect: spellCorrect, stopWords: stopWords, happyEmoticons: happyEmoticons, sadEmoticons: sadEmoticons, negationsWords: negations, myRegexp: regexp}, err
}

func createDict(filename string) (Dict, error) {
	dict := make(Dict)
	counter := 0

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file ", err)
		return dict, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		s, err := r.ReadString('\n')
		if err == io.EOF {
			// do something here
			break
		} else if err != nil {
			return dict, err // if you return error
		} else {
			s = s[0 : len(s)-1] // remove '\n'

			dict.Add(s, counter)
			counter++
		}
	}

	return dict, err
}

func createMap(emotionsString string) map[string]bool {
	m := make(map[string]bool)

	emotions := strings.Split(emotionsString, " ")
	for _, str := range emotions {
		m[str] = true
	}
	return m
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func (t *Tokenizer) IsNormalizedToken(token string) bool {
	if token == happyToken ||
		token == sadToken ||
		token == userToken ||
		token == hashtagToken ||
		token == urlToken {
		return true
	}
	return false
}

func (t *Tokenizer) normalizeTokens(text *string) {
	// Normalize apostrophe
	*text = strings.Replace(*text, "`", "'", -1)
	*text = strings.Replace(*text, "’", "'", -1)
	// Normalize https?://*
	*text = t.myRegexp.url.ReplaceAllString(*text, urlToken)
	// Normalize @username
	*text = t.myRegexp.user.ReplaceAllString(*text, userToken)
	// Normalize #hashtag
	*text = t.myRegexp.hashtag.ReplaceAllString(*text, hashtagToken)
	// Normalize happy emoticons
	for candidate := range t.happyEmoticons {
		*text = strings.Replace(*text, candidate, " "+happyToken+" ", -1)
	}
	// Normalize sad emoticons
	for candidate := range t.sadEmoticons {
		*text = strings.Replace(*text, candidate, " "+sadToken+" ", -1)
	}
	// Normalize negations
	for negation, candidate := range t.negationsWords {
		*text = strings.Replace(*text, candidate, negation, -1)
	}
}

func (t *Tokenizer) Tokenize(text string) []string {
	var acceptedTokens []string

	// Convert to lower case
	text = strings.ToLower(text)

	// replace unicode
	text = strings.Replace(text, "\\u2019", "'", -1)
	text = strings.Replace(text, "\\u002c", ",", -1)

	// Normalize url, username, hashtag, emoticon
	t.normalizeTokens(&text)

	// Remove numbers, remove special characters
	text = t.myRegexp.removeSpChars.ReplaceAllString(text, " ")

	// Remove additional white spaces
	text = t.myRegexp.removeWhitesp.ReplaceAllString(text, " ")
	// remove last space character
	text = trimSuffix(text, " ")

	// split tweet into tokens
	tokens := strings.Split(text, " ")

	for _, token := range tokens {
		if t.IsNormalizedToken(token) {
			acceptedTokens = append(acceptedTokens, token)
		} else {
			token = t.spellCorrect.SpellCheck(token)

			if _, ok := t.stopWords[token]; !ok {
				stemmed := porterstemmer.StemString(token)
				acceptedTokens = append(acceptedTokens, stemmed)
			}
		}
	}

	return acceptedTokens
}
