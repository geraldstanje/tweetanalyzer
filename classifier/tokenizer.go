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

//const acronymToken = "__acronym__"

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
	//acronyms       map[string]bool
	myRegexp myRegexp
}

func NewTokenizer() (*Tokenizer, error) {
	happy := `:-) :) ;) :o) :] :3 :c) :> =] 8) =) :} :^) :-d :d 8-d 8d x-d xd =-d =d =-3 =3 :-)) :'-) :') :* :^* >:p :-p x-p xp :p =p :-b :b >:) >;) >:-) <3 ;-) ;) ;-] ;] ;d ;^) >;) |;-)`
	sad := `>:[ :-( :( :-c :c :-< :< :-[ :[ :{ ;( :-|| :@ >:( :'-( :'( >:\\ >:/ :-/ :-. :\\ =/ =\\ :L =L :S >.< d; :-(( :(( ;-(( ;((`
	//acronym := `gr8t rotf 2moro 2nite brb btw b4n bcnu bff cya dbeyr dilligas fud fwiw gr8 ily imho irl iso j/k l8r lmao lol lylas mhoty nimby np nub oic omg ot pov rbtl rotflmao rt thx tx thks sh sitd sol stby swak tfh rtm rtfm tlc tmi ttyl ttyl tyvm vbg weg wtf wywh xoxo aw`
	// http://www.netlingo.com/top50/popular-text-terms.php

	happyEmoticons := createMap(happy)
	sadEmoticons := createMap(sad)
	//acronyms := createMap(acronym)

	url := regexp.MustCompile(`((www\.[^\s]+)|(https?://[^\s]+))`)
	user := regexp.MustCompile(`@[^\s]+`)
	hashtag := regexp.MustCompile(`#([^\s]+)`)
	removeSpChars := regexp.MustCompile(`[^a-zA-Z_'-]+`)
	removeWhitesp := regexp.MustCompile(`[\s]+`)
	regexp := myRegexp{url: url, user: user, hashtag: hashtag, removeSpChars: removeSpChars, removeWhitesp: removeWhitesp}

	stopWords, err := createDict("data/stop_words.txt") // http://www.webconfs.com/stop-words.php
	if err != nil {
		return nil, err
	}

	spellCorrect := fuzzy.NewModel()
	spellCorrect.Train(fuzzy.SampleEnglish())
	spellCorrect.Threshold = 3

	return &Tokenizer{spellCorrect: spellCorrect, stopWords: stopWords, happyEmoticons: happyEmoticons, sadEmoticons: sadEmoticons, myRegexp: regexp}, err
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
	// Normalize acronyms
	//for candidate := range t.acronyms {
	//	*text = strings.Replace(*text, candidate, " "+acronymToken+" ", -1)
	//}
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
