package classifier

import (
  "regexp"
  "strings"
  "os"
  "fmt"
  "bufio"
  "io"
  "github.com/reiver/go-porterstemmer"
)

type Tokenizer struct {
  stopWords Dict
  happyEmoticons map[string]bool
  sadEmoticons map[string]bool
}

func NewTokenizer() *Tokenizer {
  t := Tokenizer{}

  happy := `:-) :) ;) :o) :] :3 :c) :> =] 8) =) :} :^) :-d :d 8-d 8d x-d xd =-d =d =-3 =3 :-)) :'-) :') :* :^* >:p :-p x-p xp :p =p :-b :b >:) >;) >:-) <3 ;-) ;) ;-] ;] ;d ;^) >;) |;-)`
  sad := `>:[ :-( :( :-c :c :-< :< :-[ :[ :{ ;( :-|| :@ >:( :'-( :'( >:\\ >:/ :-/ :-. :\\ =/ =\\ :L =L :S >.< d; ;(`
  t.happyEmoticons = createMap(happy)
  t.sadEmoticons = createMap(sad)

  var err error
  t.stopWords, err = createDict("stop_words.txt")
  if err != nil {
    fmt.Println(err)
  }

  return &t
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
  for _,str := range emotions {
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

func (t *Tokenizer) Tokenize(text string) ([]string) {
  var acceptedTokens []string

  // Convert to lower case
  text = strings.ToLower(text)
  // replace unicode
  text = strings.Replace(text, "\\u2019", "", -1)
  text = strings.Replace(text, "\\u002c", "", -1)

  // Convert https?://* to URL
  re := regexp.MustCompile(`((www\.[^\s]+)|(https?://[^\s]+))`)
  text = re.ReplaceAllString(text, "")//URL")
  // Convert @username to AT_USER
  re = regexp.MustCompile(`@[^\s]+`)
  text = re.ReplaceAllString(text, "")//AT_USER")
  // Replace #word with word
  re = regexp.MustCompile(`#([^\s]+)`)
  text = re.ReplaceAllString(text, "$1")
  // Remove numbers
  re = regexp.MustCompile(`[0-9]*`)
  text = re.ReplaceAllString(text, "")
  // Remove special chars
  text = strings.Replace(text, ",", " ", -1)
  text = strings.Replace(text, "!", " ", -1)
  text = strings.Replace(text, "?", " ", -1)
  text = strings.Replace(text, ".", " ", -1)
  text = strings.Replace(text, "$", " ", -1)
  text = strings.Replace(text, "'", " ", -1)
  // Remove additional white spaces
  re = regexp.MustCompile(`[\s]+`)
  text = re.ReplaceAllString(text, " ")
  // remove last space character
  text = trimSuffix(text, " ")
  // split tweet into tokens
  tokens := strings.Split(text, " ")

  for i, token := range tokens {
    emoticonFound := false

    if _,ok := t.happyEmoticons[token]; ok {
      emoticonFound = true
      tokens[i] = "__h__"
    }
    if _,ok := t.sadEmoticons[token]; ok {
      emoticonFound = true
      tokens[i] = "__s__"
    }

    if emoticonFound {
      acceptedTokens = append(acceptedTokens, tokens[i])
    } else {
      if _,ok := t.stopWords[token]; !ok {
        stemmed := porterstemmer.StemString(token)
        acceptedTokens = append(acceptedTokens, stemmed)
      }
    }
  }

  return acceptedTokens
}