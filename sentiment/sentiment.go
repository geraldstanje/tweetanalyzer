package sentiment

import (
	"bufio"
	"fmt"
	"github.com/jbrukh/bayesian"
	"io"
	"os"
	"strconv"
	"strings"
)

type SliceSet map[int][]string

const (
	Negative bayesian.Class = "negative"
	Positive bayesian.Class = "positive"
)

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

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func getIndexMax(score []float64) []int {
	max := score[0]
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
	return av
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
      s = strings.Replace(s, ".", "", -1)
      s = strings.ToLower(s)
			word := strings.Split(s, "\t")

      //fmt.Println(word)
      //fmt.Println(word[0])
      //fmt.Println(word[1])

      //dict.add(stringToInt(word[1]), word[0])
			dict.add(stringToInt(word[0]), word[1])
		}
	}

	return dict, err
}

func (s *SentimentAnalysis) trainClassifier(filename string) {
	s.classifier = bayesian.NewClassifier(Negative, Positive)

	dict, _ := s.loadTrainingSet(filename)

  //fmt.Println(dict)

	for i, class := range dict {
    //fmt.Println(i)
    //fmt.Println(i)
    //fmt.Println(class[0])

		// negative sentiment score
		if i == 0 {
				s.classifier.Learn(class, Negative)
			// positive sentiment score
		} else if i == 1 {
				s.classifier.Learn(class, Positive)
		}

    /*if i >= -5 && i <= -1 {
      s.classifier.Learn(class, Negative)
    } else if i > 1 && i <= 5 {
      s.classifier.Learn(class, Positive)
    }

    if i == 0 {
      s.classifier.Learn(class, Negative)
      s.classifier.Learn(class, Positive)
    }*/

	}
}

func (s *SentimentAnalysis) getClass(sentence string) int {
	// split the sentence into tokens
	words := s.tokenize(strings.ToLower(sentence))
	// get the score for each class
	score, _, _ := s.classifier.LogScores(words)

  fmt.Println(score)

  if score[0] > score[1] {
    return 0 
  }

  return 1
	// get class with max value
	//classNum := getIndexMax(score)

	//if classNum[1] > 0 {
	//	return 2
	//}

	//if classNum[0] == 0 {
	//	return 0
	//}

	//return 1
}
