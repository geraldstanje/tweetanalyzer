package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
  "sort"
  "github.com/jbrukh/bayesian"
)

type SliceSet map[int][]string

type SentimentAnalysis struct {
  classifier *bayesian.Classifier
}

func stringToInt(str string) int {
  intVal, _ := strconv.ParseInt(str, 0, 64)
  return int(intVal)
}

func intToString(input_num int) string {
  return strconv.FormatInt(int64(input_num), 10)
}

func getIndexMax(score []float64) [2]int {
  /*max := score[0]
  max_i := 0
  max_i_2 := 0

  for i, val := range score {
    if val > max {
      max = val
      max_i = i
    } else if val == max {
      max_i_2 = i
    }
  }

  var av = []int{max_i, max_i_2}
  return av*/

  m := make(map[float64]int)
  var keys []float64
  for i, k := range score {
      m[k] = i
      keys = append(keys, k)
  }
  sort.Sort(sort.Reverse(sort.Float64Slice(keys)))
  var av [2]int
  for i, k := range keys {
      fmt.Println(i)
      if i < 2 {
        av[i] = m[k]
      }

      fmt.Println("Key:", k, "Value:", m[k])
  }

  //var av = []int{m[10], m[9]}

  fmt.Println(av)

  return  av
}

func (s SliceSet) add(key int, value string) {
  s[key] = append(s[key], value)
}

func (s SliceSet) peek(key int) ([]string, bool) {
  ret, ok := s[key]
  return ret, ok
}

func (s *SentimentAnalysis) tokenize(text string) []string {
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

	s.createBigrams(text, &result)

	return result
}

func (s *SentimentAnalysis) createBigrams(text string, result *[]string) {
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
		j += k + 2
	}
}

func (s *SentimentAnalysis) loadTrainingSet(filename string) (SliceSet, error) {
  dict := make(SliceSet)

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
      word := strings.Split(s, "\t")
      dict.add(stringToInt(word[1]), word[0])
    }
  }

  return dict, err
}

func (s *SentimentAnalysis) trainClassifier(filename string) {
  s.classifier = bayesian.NewClassifier("-5", "-4", "-3", "-2", "-1", "0", "1", "2", "3", "4", "5")

  dict, _ := s.loadTrainingSet(filename)

  //fmt.Println(dict[-3])
  //fmt.Println(dict[3])

  for i, class := range dict {
    //fmt.Println(len(class))

    s.classifier.Learn(class, bayesian.Class(intToString(i)))
  }
}

func (s *SentimentAnalysis) getClass(sentence string) int {
  // split the sentence into tokens
  words := s.tokenize(strings.ToLower(sentence))
  // get the score for each class
  score, _, _ := s.classifier.LogScores(words)

  fmt.Println(score)

  // get class with max value
  classNum := getIndexMax(score)

  classVal := classNum[0]
  if classNum[1] > 0 {
    classNum[1] = classNum[1] - 5
    classVal = classVal + classNum[1]
  }

  return classVal - 5
}

func main() {
  s := new(SentimentAnalysis)

  s.trainClassifier("AFINN-111.txt")

  classVal := s.getClass("I hate the brilliant good pizza") //("I love the fucking pizza") //brilliant
  
  fmt.Println(classVal)

  if classVal >= -5 && classVal <= -1 {
    fmt.Println("negative")
  } else if classVal > -1 && classVal <= 1 {
    fmt.Println("neutral")
  } else if classVal > 1 && classVal <= 5 {
    fmt.Println("positive")
  }
}
