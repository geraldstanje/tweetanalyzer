package main

import (
  "github.com/white-pony/go-fann"
  "fmt"
  "os"
  "bufio"
  "io"
  "strings"
  "strconv"
  "sort"
  "regexp"
)

type SliceData struct {
    i int 
    s string
}

/*func (s SliceData) add(key int, value string) {
  s[key] = append(s[key], value)
}

func (s SliceData) peek(key int) ([]string, bool) {
  ret, ok := s[key]
  return ret, ok
}*/

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

var splitTokenRx = regexp.MustCompile(`\s`)

func tokenize(sl []SliceData) map[string]int {
  dict := make(map[string]int)

  for _, val := range sl {
    //fmt.Println(val.s)
    for _, w := range splitTokenRx.Split(val.s, -1) {
      dict[w]++
    }
  }

  return dict
}

// A data structure to hold a key/value pair.
type Pair struct {
  Key string
  Value int
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair
func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

// A function to turn a map into a PairList, then sort and return it. 
func sortMapByValue(m map[string]int) PairList {
   p := make(PairList, len(m))
   i := 0
   for k, v := range m {
      p[i] = Pair{k, v}
   }
   sort.Sort(p)
   return p
}

func main() {
  const numLayers = 3
  const desiredError = 0.00001
  const maxEpochs = 500000
  const epochsBetweenReports = 1000

  trainingSet, _ := loadDataSet("trainingset.txt")
  dict := tokenize(trainingSet)

  for key, val := range dict {
    fmt.Println(key, val)
  }

  //fmt.Println(dict)

  vs := sortMapByValue(dict)
  for _, _ = range vs {
    //fmt.Println(i.Value)
  }

  return 

  //fmt.Println(trainingSet)

  ann := fann.CreateStandard(numLayers, []uint32{2, 3, 1})

  ann.SetTrainingAlgorithm(fann.TRAIN_INCREMENTAL)
  //ann.SetActivationFunctionHidden(fann.SIGMOID)
  ann.SetActivationFunctionOutput(fann.SIGMOID_SYMMETRIC)
  ann.SetLearningMomentum(0.4)
  
  trainInput := [][]fann.FannType{{-1.0, -1.0}, {-1.0, 1.0}, {1.0, -1.0}, {1.0, 1.0}}
  trainOutput := [][]fann.FannType{{-1.0}, {1.0}, {1.0}, {-1.0}}

  for i := 1; i <= maxEpochs; i++ {
    ann.ResetMSE()

    for i, _ := range trainInput {
      ann.Train(trainInput[i], trainOutput[i])
    }

    if ann.GetMSE() < desiredError {
      break
    }
  }

  fmt.Println("Testing network")

  test_data := fann.ReadTrainFromFile("xor.test")

  ann.ResetMSE()

  var i uint32
  for i = 0; i < test_data.Length(); i++ {
    ann.Test(test_data.GetInput(i), test_data.GetOutput(i))
  }

  fmt.Printf("MSE error on test data: %f\n", ann.GetMSE())


  fmt.Println("Saving network.");
  ann.Save("robot_float.net")
  fmt.Println("Cleaning up.")

  input := []fann.FannType{-1.0, 1.0}
  output := ann.Run(input)
  fmt.Printf("xor test (%f,%f) -> %f\n", input[0], input[1], output[0])

  test_data.Destroy()
  ann.Destroy()
}
