package classifier

import (
  "io"
  "fmt"
  "os"
  "bufio"
  "strings"
  "github.com/reiver/go-porterstemmer"
  "github.com/sridif/gosvm"
)

type SvmClassifier struct {
  model *svm.Model
}

type SentimentData struct {
  i int    // label: negative == -1, neutral == 0, positive == 1
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

func NewSvmClassifier() *SvmClassifier {
  return &SvmClassifier{}
}

func (c *SvmClassifier) CreateFeatureVector(text string) ([]float64) {
  featureVec := make([]float64, len(c.bagOfWords))

  tokenizer := NewTokenizer()
  tokens := tokenizer.Tokenize(text)

  for _, str := range tokens {
    stemmed := porterstemmer.StemString(str)

    if val, ok := c.bagOfWords[stemmed]; ok {
      featureVec[val] = featureVec[val] + 1
    }
  }

  return featureVec
}

func (c *SvmClassifier) addSentimentData(word []string, index1 int, index2 int) (SentimentData, error) {  
  var t SentimentData
  
  sentiment := word[index1]

  if (strcmp(sentiment, "positive") == 0) || 
     (strcmp(sentiment, "neutral") == 0) || 
     (strcmp(sentiment, "negative") == 0) {
    var i int
    var str string

    if strcmp(sentiment, "positive") == 0 {
      i = 1
    } else if strcmp(sentiment, "negative") == 0 {
      i = -1
    } else if strcmp(sentiment, "neutral") == 0 {
      i = 0
    }

    str = word[index2]
    t = SentimentData{i, str[0:len(str)-1]}
    return t, nil
  }

  return t, fmt.Errorf("Error: incorrect SentimentData format")
}

func (c *SvmClassifier) loadTrainDataSet(filename string, index1 int, index2 int) ([]SentimentData, error) {
  dict := make([]SentimentData, 0)

  f, err := os.Open(filename)
  if err != nil {
    fmt.Println("error opening file ", err)
    return nil, err
  }
  defer f.Close()
  r := bufio.NewReader(f)
  for {
    s, err := r.ReadString('\n')
    if err == io.EOF {
      // do something here
      break
    } else if err != nil {
      return nil, err // if you return error
    } else {
      s = s[0 : len(s)-1] // remove '\n'
      s = strings.ToLower(s)

      word := strings.Split(s, "\t")

      if len(word) > 1 {
        t, err := c.addSentimentData(word, index1, index2)
        if err != nil {
          return nil, err
        }
        dict = append(dict, t)
      }
    }
  }

  return dict, err
}

func (c *SvmClassifier) loadTestDataSet(filename string, index1 int, index2 int) ([]SentimentData, error) {
  dict := make([]SentimentData, 0)

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

      if len(word) > 1 && strings.HasPrefix(word[0], "twitter") {
        t, err := c.addSentimentData(word, index1, index2)
        if err != nil {
          return nil, err
        }
        dict = append(dict, t)
      }
    }
  }

  return dict, err
}

func (c *SvmClassifier) createDict(filename string) (SliceSet, error) {
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

func (c *SvmClassifier) calcWordFreq(s1 []SliceData, s2 []SliceData) (PairList, error) {
  dict := make(map[string]int)
  var sorted PairList

  for _, sentence := range s1 {
    sentence.s = strings.ToLower(sentence.s)
    words := myExp.FindAllString(sentence.s, -1)

    for _, w := range words {
      stemmed := porterstemmer.StemString(w)
      if len(stemmed) > 1 {
        dict[stemmed] = dict[stemmed] + 1
      }
    }
  }

for _, sentence := range s2 {
    sentence.s = strings.ToLower(sentence.s)
    words := myExp.FindAllString(sentence.s, -1)

    for _, w := range words {
      stemmed := porterstemmer.StemString(w)
      if len(stemmed) > 1 {
        dict[stemmed] = dict[stemmed] + 1
      }
    }
  }

  sorted = sortMapByValue(dict)

  return sorted, nil
}

func (c *SvmClassifier) createBagOfWords(filename string, freqMin int, freqMax int, stopWordsfilename string, trainingDataSet1 []SliceData, trainingDataSet2 []SliceData) error {
  var buffer bytes.Buffer

  calcWordFreq, err := c.calcWordFreq(trainingDataSet1, trainingDataSet2)
  if err != nil {
    return err
  }

  for _, word := range wordFreq {
    if _, ok := stopWords[word.Key]; !ok {
      if word.Value >= freqMin && word.Value <= freqMax {
        buffer.WriteString(word.Key)
        buffer.WriteString("\n")
      }
    }
  }

  err := ioutil.WriteFile(filename, buffer.Bytes(), 0644)
  return err
}

func (c *SvmClassifier) loadBagOfWords(filename string) (SliceSet, error) {
  
}

func (c *SvmClassifier) TrainClassifier(trainDataSetFile1 string, trainDataSetFile2 string) (error) {
  // Perform training
  fmt.Println("Start Training")
  start := time.Now()

  // train the model
  trainingData1, err := c.loadTrainDataSet(trainDataSetFile1, 2, 3)
  if err != nil {
    return err
  }

  trainingData2, err := c.loadTrainDataSet(trainDataSetFile2, 2, 3)
  if err != nil {
    return err
  }

  problem := gosvm.NewProblem()

  // We will use the words from the bagofWords as our features
  for _, val := range trainingData1 {
    problem.Add(gosvm.TrainingInstance{float64(val.i), gosvm.FromDenseVector(tokenize(bagOfWords, val.s))})
  }

  for _, val := range trainingData2 {
    problem.Add(gosvm.TrainingInstance{float64(val.i), gosvm.FromDenseVector(tokenize(bagOfWords, val.s))})
  }

  param := gosvm.DefaultParameters()
  param.Kernel = gosvm.NewLinearKernel()
  param.SVMType = gosvm.NewCSVC(0.05)
  c.model, err = gosvm.TrainModel(param, problem)
  if err != nil {
    log.Fatal(err)
  }

  elapsed := time.Now().Sub(start)
  fmt.Println(elapsed)
  fmt.Println("Training finished!")
}

func (c *SvmClassifier) TestClassifier(testDataSetFile string) (error) {
  fmt.Println("Start Testing")
  start := time.Now()

  // test the model
  testData, err := c.loadTestDataSet(testDataSetFile, 2, 3)
  if err != nil {
    return err
  }

  flailCounter := 0
  counter := 0

  for _, val := range testData {
    label := c.model.Predict(gosvm.FromDenseVector(tokenize(c.bagOfWords, val.s)))

    if int(label) != val.i {
      flailCounter++
    }

    counter++
  }

  // print error rate
  fmt.Println("Error:", float64(flailCounter) / float64(counter))

  elapsed := time.Now().Sub(start)
  fmt.Println(elapsed)
  fmt.Println("Test finished!")
}