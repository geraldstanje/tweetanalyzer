package main

import (
	"fmt"
	"log"
  "strings"
  "regexp"
	"github.com/sridif/gosvm"
)

var myExp = regexp.MustCompile(`\s`)

/*
 1   2   3   4   5
          +---+---+---+---+---+
Positive: | 1 | 1 | 1 | 0 | 0 |
          +---+---+---+---+---+

          +---+---+---+---+---+
Negative: | 1 | 0 | 1 | 1 | 1 |
          +---+---+---+---+---+
*/

func tokenize(sentence string) []float64 {
  dict := map[string]int {
    "a": 0,
    "beautiful": 1,
    "album": 2,
    "crappy": 3,
    "ugly": 4,
  }

  var vec []float64
  vec = make([]float64, 5)
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

  // We will use the words as our features (a: 1, beautiful: 2, album: 3, crappy: 4, ugly: 5)
	problem.Add(gosvm.TrainingInstance{0, gosvm.FromDenseVector(tokenize("A beautiful album"))})
	problem.Add(gosvm.TrainingInstance{1, gosvm.FromDenseVector(tokenize("A crappy ugly album"))})

	param := gosvm.DefaultParameters()
  fmt.Println(param)
	model, err := gosvm.TrainModel(param, problem)
	if err != nil {
		log.Fatal(err)
	}

	label1 := model.Predict(gosvm.FromDenseVector(tokenize("This is a beautiful book")))
	fmt.Printf("Predicted label: %f\n", label1)

  label2 := model.Predict(gosvm.FromDenseVector(tokenize("Thoday is crappy weather")))
  fmt.Printf("Predicted label: %f\n", label2)
}