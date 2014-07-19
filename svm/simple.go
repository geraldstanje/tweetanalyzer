package main

import (
	"bufio"
	"fmt"
	"github.com/sridif/gosvm"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var myExp = regexp.MustCompile(`\s`)

type SliceData struct {
	i int // label: negative == 0, positive == 1
	s string // text
}

type SliceSet map[string]int

func (s SliceSet) Add(key string, value int) {
  s[key] = value
}

func (s SliceSet) Peek(key string) (int, bool) {
  ret, ok := s[key]
  return ret, ok
}

func stringToInt(str string) int {
	intVal, _ := strconv.ParseInt(str, 0, 64)
	return int(intVal)
}

func intToString(input_num int) string {
	return strconv.FormatInt(int64(input_num), 10)
}

func loadDataSet(filename string) ([]SliceData, error) {
	dict := make([]SliceData, 1)

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
			s = strings.ToLower(s)
			word := strings.Split(s, "\t")

			x := SliceData{}
			if len(word) > 1 {
				x.i = stringToInt(word[0])
				x.s = word[1]
			} else {
				x.s = word[0]
			}

			dict = append(dict, x)
		}
	}

	return dict, err
}

func createDict(filename string) (SliceSet, error) {
  dict := make(SliceSet)
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

func tokenize(dict SliceSet, sentence string) []float64 {
	var vec []float64
	vec = make([]float64, len(dict))
	sentence = strings.ToLower(sentence)
	words := myExp.Split(sentence, -1)

	for _, w := range words {
		if val, ok := dict[w]; ok {
			vec[val] = 1
		}
	}

	return vec
}

func main() {
	problem := gosvm.NewProblem()

  // create bag of words dictionary, which is used for the densevector
  bagOfWords, err := createDict("bag_of_words.txt")
  if err != nil {
    fmt.Println(err)
  }

  // train the model
  trainingData, err := loadDataSet("trainingset.txt")
  if err != nil {
    log.Fatal(err)
  }

	// We will use the words as our features
	for _, val := range trainingData {
		problem.Add(gosvm.TrainingInstance{float64(val.i), gosvm.FromDenseVector(tokenize(bagOfWords, val.s))})
	}

	param := gosvm.DefaultParameters()
	param.Kernel = gosvm.NewRBFKernel(0.2) //NewPolynomialKernel(1.0, 0.1, 1)
	model, err := gosvm.TrainModel(param, problem)
	if err != nil {
		log.Fatal(err)
	}

  // test the model
  testData, err := loadDataSet("testdata.txt")
  if err != nil {
    log.Fatal(err)
  }
  flailCounter := 0
  counter := 0
  for _, val := range testData {
    // we dont analyze neutral messages
    if val.i != 2 {
      label := model.Predict(gosvm.FromDenseVector(tokenize(bagOfWords, val.s)))

      if val.i == 4 {
        val.i = 1
      }

      if int(label) != val.i {
        flailCounter++
      }

      counter++
    }
  }

  // print error rate
  fmt.Println(float64(flailCounter)/float64(counter))
}
