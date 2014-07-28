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
  "sort"
  "io/ioutil"
  "bytes"
)

var myExp = regexp.MustCompile(`\s`)

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

// A data structure to hold a key/value pair.
type Pair struct {
  Key string
  Value int
}

type SliceData struct {
	i int    // label: negative == 0, positive == 1
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

func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

// A function to turn a map into a PairList, then sort and return it. 
func sortMapByValue(m map[string]int) PairList {
   p := make(PairList, len(m))
   i := 0
   for k, v := range m {
      p[i] = Pair{k, v}
      i++
   }
   sort.Sort(p)
   return p
}

func stringToInt(str string) int {
	intVal, _ := strconv.ParseInt(str, 0, 64)
	return int(intVal)
}

func intToString(input_num int) string {
	return strconv.FormatInt(int64(input_num), 10)
}

func strcmp(a, b string) int {
  min := len(b)
  if len(a) < len(b) {
    min = len(a)
  }
  diff := 0
  for i := 0; i < min && diff == 0; i++ {
    diff = int(a[i]) - int(b[i])
  }
  if diff == 0 {
    diff = len(a) - len(b)
  }
  return diff
}

func loadDataSet(filename string, index1 int, index2 int) ([]SliceData, error) {
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

      if len(word) > 1 {
        if (strcmp(word[index1], "positive") == 0) || (strcmp(word[index1], "neutral") == 0) || (strcmp(word[index1], "negative") == 0) {
          x := SliceData{}
			    
          if strcmp(word[index1], "positive") == 0 {
            x.i = 1
          } else if strcmp(word[index1], "negative") == 0 {
            x.i = -1
          } else if strcmp(word[index1], "neutral") == 0 {
            x.i = 0
          }

				  x.s = word[index2]
          dict = append(dict, x)
        }
      }
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
    x := w
    x = strings.Replace(x, "\t", "", -1)
    x = strings.Replace(x, "?", "", -1)
    x = strings.Replace(x, "!", "", -1)
    x = strings.Replace(x, ".", "", -1)
    x = strings.Replace(x, ",", "", -1)
    x = strings.Replace(x, ":", "", -1)
    x = strings.Replace(x, "(", "", -1)
    x = strings.Replace(x, ")", "", -1)
    x = strings.Replace(x, "-", "", -1)
    x = strings.Replace(x, "0", "", -1)
    x = strings.Replace(x, "1", "", -1)
    x = strings.Replace(x, "2", "", -1)
    x = strings.Replace(x, "\"", "", -1)

		if val, ok := dict[x]; ok {
			vec[val] = 1
		}
	}

	return vec
}

func calcWordFreq(s []SliceData) (PairList, error) {
  dict := make(map[string]int)
  var sorted PairList

  for _, sentence  := range(s) {
    x := sentence.s
    x = strings.ToLower(x)
    x = strings.Replace(x, "\t", "", -1)
    x = strings.Replace(x, "?", "", -1)
    x = strings.Replace(x, "!", "", -1)
    x = strings.Replace(x, ".", "", -1)
    x = strings.Replace(x, ",", "", -1)
    x = strings.Replace(x, ":", "", -1)
    x = strings.Replace(x, "(", "", -1)
    x = strings.Replace(x, ")", "", -1)
    x = strings.Replace(x, "-", "", -1)
    x = strings.Replace(x, "0", "", -1)
    x = strings.Replace(x, "1", "", -1)
    x = strings.Replace(x, "2", "", -1)
    x = strings.Replace(x, "\"", "", -1)

    words := strings.Split(x, " ")

    for _, w := range(words) {
      if len(w) > 1 {
        dict[w] = dict[w] + 1
      }
    }
  }

  sorted = sortMapByValue(dict)
  return sorted, nil
}

func createBagOfWords(filename string, wordFreq PairList, duplicateThreshold int) error {
  var buffer bytes.Buffer

  for _, word := range(wordFreq) {
    if word.Value > duplicateThreshold {
      buffer.WriteString(word.Key + "\n")
    }
  }

  err := ioutil.WriteFile(filename, []byte(buffer.String()), 0644)
  return err
}

func main() {
	problem := gosvm.NewProblem()

	// train the model
	trainingData, err := loadDataSet("2014_a_train.txt", 4, 5)
	if err != nil {
		log.Fatal(err)
	}

  wordFreq, err := calcWordFreq(trainingData)
  if err != nil {
    log.Fatal(err)
  }

  err = createBagOfWords("bag_of_words.txt", wordFreq, 30)
  if err != nil {
    log.Fatal(err)
  }

  // create bag of words dictionary, which is used for the densevector
  bagOfWords, err := createDict("bag_of_words.txt")
  if err != nil {
    fmt.Println(err)
  }

	// We will use the words as our features
	for _, val := range trainingData {
		problem.Add(gosvm.TrainingInstance{float64(val.i), gosvm.FromDenseVector(tokenize(bagOfWords, val.s))})
	}

	param := gosvm.DefaultParameters()
  param.Kernel = gosvm.NewLinearKernel()
	param.SVMType = gosvm.NewCSVC(0.005)//0.005)
  //param.Kernel = gosvm.NewRBFKernel(0.2) //NewPolynomialKernel(1.0, 0.1, 1)
	model, err := gosvm.TrainModel(param, problem)
	if err != nil {
		log.Fatal(err)
	}

	// test the model
	testData, err := loadDataSet("2014_a_dev.txt", 4, 5)
	if err != nil {
		log.Fatal(err)
	}

	flailCounter := 0
	counter := 0
	for _, val := range testData {
		label := model.Predict(gosvm.FromDenseVector(tokenize(bagOfWords, val.s)))
    
    if int(label) != val.i {
			flailCounter++
		}

		counter++
	}

	// print error rate
	fmt.Println(float64(flailCounter) / float64(counter))
}
