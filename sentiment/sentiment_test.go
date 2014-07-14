package sentiment

import (
	"testing"
  "strings"
  "fmt"
  "os"
  //"strconv"
  "bufio"
  "io"
)

type SliceData struct {
    i int 
    s string
}

//func (s SliceData) add(key int, value string) {
//  s = append(s, SliceData{key, value})
//}

func loadTestSet(filename string) ([]SliceData, error) {
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
      s = strings.Replace(s, ".", "", -1)
      s = strings.ToLower(s)
      word := strings.Split(s, "\t")

      //fmt.Println(word)
      //fmt.Println(word[0])
      //fmt.Println(word[1])

      //dict.add(stringToInt(word[1]), word[0])
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

/*func stringToInt(str string) int {
  intVal, _ := strconv.ParseInt(str, 0, 64)
  return int(intVal)
}

func intToString(input_num int) string {
  return strconv.FormatInt(int64(input_num), 10)
}*/

func TestSimple(t *testing.T) {
	s := new(SentimentAnalysis)

	s.trainClassifier("trainingset.txt") //("AFINN-111.txt") //("trainingset.txt")

  x, _ := loadTestSet("testdata.txt")
  failCounter := 0
  succCounter := 0
  counter := 0

  for _, value := range x {
    //fmt.Println(value.s)

    classVal := s.getClass(strings.ToLower(value.s))
    if classVal != value.i {
      //t.Error("classification failed")
      //fmt.Println("classification failed")
      failCounter = failCounter + 1
    } else {
      //fmt.Println("classification successful")
      succCounter = succCounter + 1
    }

    counter = counter + 1
    
    //fmt.Println(value.i)

    if counter == 35 {
      break
    }
  }

  fmt.Println(float64(failCounter) / float64(counter))

  //fmt.Println(failCounter)
  //fmt.Println(succCounter)
}
