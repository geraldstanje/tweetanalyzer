package classifier

import (
  "regexp"
  "strings"
  "github.com/reiver/go-porterstemmer"
)

type Tokenizer struct {
}

func NewTokenizer() *Tokenizer {
  return &Tokenizer{}
}

func createMap(emotionsString string) map[string]bool {
  m := make(map[string]bool)

  emotions := strings.Split(emotionsString, " ")
  for _,str := range emotions {
    m[str] = true
  }

  return m
}

func (t *Tokenizer) Tokenize(text string) ([]string) {  
  happy := `:-) :) ;) :o) :] :3 :c) :> =] 8) =) :} :^) :-D :D 8-D 8D x-D xD X-D XD =-D =D =-3 =3 :-)) :'-) :')
            :* :^* >:P :-P :P X-P x-p xp XP :-p :p =p :-b :b >:) >;) >:-) <3`
  sad := ">:[ :-( :( :-c :c :-<  :< :-[ :[ :{ ;( :-|| :@ >:( :'-( :'( >:\\ >:/ :-/ :-. :\\ =/ =\\ :L =L :S >.<"

  happyEmoticons := createMap(happy)
  sadEmoticons := createMap(sad)

  // Convert to lower case
  text = strings.ToLower(text)
  // Convert https?://* to URL
  re := regexp.MustCompile(`((www\.[^\s]+)|(https?://[^\s]+))`)
  text = re.ReplaceAllString(text, "URL")
  // Convert @username to AT_USER
  re = regexp.MustCompile(`@[^\s]+`)
  text = re.ReplaceAllString(text, "AT_USER")
  // Replace #word with word
  re = regexp.MustCompile(`#([^\s]+)`)
  text = re.ReplaceAllString(text, "$1")
  // Remove numbers 
  re = regexp.MustCompile(`[0-9]*`)
  text = re.ReplaceAllString(text, "")
  // Remove additional white spaces
  re = regexp.MustCompile(`[\s]+`)
  text = re.ReplaceAllString(text, " ")
  // split tweet into tokens
  tokens := strings.Split(text, " ")

  for i, t := range tokens {
    if _,ok := happyEmoticons[t]; ok {
      tokens[i] = "__h__"
    }
    if _,ok := sadEmoticons[t]; ok {
      tokens[i] = "__s__"
    }

    t = porterstemmer.StemString(t)
  }

  return tokens
}