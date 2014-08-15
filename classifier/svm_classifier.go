package classifier

import (
  "io"
  "fmt"
  "os"
  "bufio"
  "strings"
  "time"
  "io/ioutil"
  "bytes"
  "github.com/sridif/gosvm"
)

const debug_output_bag_of_words = true

type SvmClassifier struct {
  model *gosvm.Model
  bagOfWords Dict
  spellCorrect *SpellCorrect
}

type SentimentData struct {
  sentimentLabel int // negative == -1, neutral == 0, positive == 1
  text string
}

func (c *SvmClassifier) createFeatureVector(text string) ([]float64) {
  tokenizer := NewTokenizer()
  featureVec := make([]float64, len(c.bagOfWords))

  tokens := tokenizer.Tokenize(text)

  for _, str := range tokens {
    if val, ok := c.bagOfWords[str]; ok {
      featureVec[val] = featureVec[val] + 1
    }
  }

  return featureVec
}

func (c *SvmClassifier) addSentimentData(word []string, index1 int, index2 int) (SentimentData, error) {  
  var t SentimentData
  
  sentiment := word[index1]

  if (strcmp(sentiment, "positive") == 0) || 
     (strcmp(sentiment, "extremely-positive") == 0) ||
     (strcmp(sentiment, "neutral") == 0) || 
     (strcmp(sentiment, "negative") == 0) ||
     (strcmp(sentiment, "extremely-negative") == 0) {
    var i int
    var str string

    if strcmp(sentiment, "positive") == 0 || strcmp(sentiment, "extremely-positive") == 0 {
      i = 1
    } else if strcmp(sentiment, "negative") == 0 || strcmp(sentiment, "extremely-negative") == 0 {
      i = -1
    } else if strcmp(sentiment, "neutral") == 0 {
      i = 0
    }

    str = word[index2]
    t = SentimentData{i, str} // problem with emotiocons: use: t = SentimentData{i, str[0:len(str)-1]}
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

func (c *SvmClassifier) calcWordFreq(s1 []SentimentData, s2 []SentimentData) (Dict, error) {
  tokenizer := NewTokenizer()
  dict := make(Dict)

  for _, sentence := range s1 {
    tokens := tokenizer.Tokenize(sentence.text)

    for _, token := range tokens {
      if len(token) > 1 {
        dict[token] = dict[token] + 1
      }
    }
  }

  for _, sentence := range s2 {
    tokens := tokenizer.Tokenize(sentence.text)

    for _, token := range tokens {
      if len(token) > 1 {
        dict[token] = dict[token] + 1
      }
    }
  }

  SortMapByValue(&dict)
  return dict, nil
}

func (c *SvmClassifier) createBagOfWords(freqMin int, freqMax int, trainingDataSet1 []SentimentData, trainingDataSet2 []SentimentData) error {
  var buffer bytes.Buffer
  c.bagOfWords = make(Dict)

  wordFreq, err := c.calcWordFreq(trainingDataSet1, trainingDataSet2)
  if err != nil {
    return err
  }

  counter := 0
  for key, value := range wordFreq {
    if value >= freqMin && value <= freqMax {
      c.bagOfWords.Add(key, counter)
      counter++
    }
  }

  if debug_output_bag_of_words {
    for word, _ := range c.bagOfWords {
      buffer.WriteString(word)
      buffer.WriteString("\n")
    }
  }

  if debug_output_bag_of_words {
    err = ioutil.WriteFile("bagOfWords.txt", buffer.Bytes(), 0644)
  }
  return err
}

func (c *SvmClassifier) TrainClassifier(trainDataSetFile1 string, trainDataSetFile2 string, stopWordsFile string, emonticonsFile string) (error) {
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

  err = c.createBagOfWords(5, 1000, trainingData1, trainingData2)
  if err != nil {
    return err
  }

  problem := gosvm.NewProblem()

  // We use all features from the bagofWords
  for _, val := range trainingData1 {
    problem.Add(gosvm.TrainingInstance{float64(val.sentimentLabel), gosvm.FromDenseVector(c.createFeatureVector(val.text))})
  }

  for _, val := range trainingData2 {
    problem.Add(gosvm.TrainingInstance{float64(val.sentimentLabel), gosvm.FromDenseVector(c.createFeatureVector(val.text))})
  }

  param := gosvm.DefaultParameters()
  param.Kernel = gosvm.NewLinearKernel()
  param.SVMType = gosvm.NewCSVC(0.05)
  c.model, err = gosvm.TrainModel(param, problem)

  elapsed := time.Now().Sub(start)
  fmt.Println(elapsed)
  fmt.Println("Training finished!")

  return err
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
    label := c.model.Predict(gosvm.FromDenseVector(c.createFeatureVector(val.text)))

    if int(label) != val.sentimentLabel {
      flailCounter++
    }

    counter++
  }

  // print error rate
  fmt.Println("Error:", float64(flailCounter) / float64(counter))

  elapsed := time.Now().Sub(start)
  fmt.Println(elapsed)
  fmt.Println("Test finished!")

  return nil
}

func NewSvmClassifier() *SvmClassifier {
  spellCorrect := NewSpellCorrect() // handle error in case big.txt does not exist

  return &SvmClassifier{spellCorrect: spellCorrect}
}