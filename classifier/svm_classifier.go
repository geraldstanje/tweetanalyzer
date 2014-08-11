package classifier

import (
  "io"
  "fmt"
  "os"
  "bufio"
  "strings"
  "regexp"
  "time"
  "io/ioutil"
  "bytes"
  "github.com/reiver/go-porterstemmer"
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
  featureVec := make([]float64, len(c.bagOfWords))

  tokenizer := NewTokenizer()
  tokens := tokenizer.Tokenize(text)

  for _, str := range tokens {
    //str = c.spellCorrect.correct(str)
    stemmed := porterstemmer.StemString(str)

    if val, ok := c.bagOfWords[stemmed]; ok {
      featureVec[val] = featureVec[val] + 1
    }
  }

  /*bigrams := createBigrams(text)

  for _, w := range bigrams {
    if len(w) > 1 {
      if val, ok := c.bagOfWords[w]; ok {
        featureVec[val] = featureVec[val] + 1
      }
    }
  }*/

  return featureVec
}

func (c *SvmClassifier) addSentimentData(word []string, index1 int, index2 int, flag bool) (SentimentData, error) {  
  var t SentimentData
  
  sentiment := word[index1]
  sentiment = strings.ToLower(sentiment)

  if (strcmp(sentiment, "positive") == 0) || 
     (strcmp(sentiment, "extremely-positive") == 0) ||
     (strcmp(sentiment, "neutral") == 0) || 
     (strcmp(sentiment, "negative") == 0) ||
     (strcmp(sentiment, "extremely-negative") == 0) {
    var i int
    var str string

    if strcmp(sentiment, "positive") == 0 || strcmp(sentiment, "extremely-positive") == 0 {
      i = 1
    } else if strcmp(sentiment, "negative") == 0 {
      i = -1
    } else if strcmp(sentiment, "neutral") == 0 || strcmp(sentiment, "extremely-negative") == 0 {
      i = 0
    }

    str = word[index2]
    if !flag {
      t = SentimentData{i, str[0:len(str)-1]}
      return t, nil
    } else {
      t = SentimentData{i, str}
      return t, nil
    }
  }

  return t, fmt.Errorf("Error: incorrect SentimentData format")
}

func (c *SvmClassifier) loadTrainDataSet(filename string, index1 int, index2 int, flag bool) ([]SentimentData, error) {
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
        t, err := c.addSentimentData(word, index1, index2, flag)
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
        t, err := c.addSentimentData(word, index1, index2, false)
        if err != nil {
          return nil, err
        }
        dict = append(dict, t)
      }
    }
  }

  return dict, err
}

func (c *SvmClassifier) createDict(filename string) (Dict, error) {
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

func (c *SvmClassifier) calcWordFreq(s1 []SentimentData, s2 []SentimentData) (Dict, error) {
  dict := make(Dict)
  var myExp = regexp.MustCompile(`([A-Za-z]+)`)

  for _, sentence := range s1 {
    sentence.text = strings.ToLower(sentence.text)
    words := myExp.FindAllString(sentence.text, -1)

    for _, w := range words {
      //w = c.spellCorrect.correct(w)

      stemmed := porterstemmer.StemString(w)
      if len(stemmed) > 1 {
        dict[stemmed] = dict[stemmed] + 1
      }
    }

    /*bigrams := createBigrams(sentence.text)

    for _, w := range bigrams {
      if len(w) > 1 {
        dict[w] = dict[w] + 1
      }
    }*/
  }

  for _, sentence := range s2 {
    sentence.text = strings.ToLower(sentence.text)
    words := myExp.FindAllString(sentence.text, -1)

    for _, w := range words {
      //w = c.spellCorrect.correct(w)

      stemmed := porterstemmer.StemString(w)
      if len(stemmed) > 1 {
        dict[stemmed] = dict[stemmed] + 1
      }
    }

    /*bigrams := createBigrams(sentence.text)

    for _, w := range bigrams {
      if len(w) > 1 {
        dict[w] = dict[w] + 1
      }
    }*/
  }

  SortMapByValue(&dict)

  return dict, nil
}

func (c *SvmClassifier) createBagOfWords(stopWordsFile string, freqMin int, freqMax int, trainingDataSet1 []SentimentData, trainingDataSet2 []SentimentData, emoticons []SentimentData) error {
  var buffer bytes.Buffer

  c.bagOfWords = make(Dict, 0)

  stopWords, err := c.createDict(stopWordsFile)
  if err != nil {
    return err
  }

  wordFreq, err := c.calcWordFreq(trainingDataSet1, trainingDataSet2)
  if err != nil {
    return err
  }

  for key, value := range wordFreq {
    if _, ok := stopWords[key]; !ok {
      if value >= freqMin && value <= freqMax {
        c.bagOfWords.Add(key, value)
      }
    }
  }

  //for _, word := range emoticons {
    //c.bagOfWords.Add(word.text, 1)
  //}

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
  trainingData1, err := c.loadTrainDataSet(trainDataSetFile1, 2, 3, false)
  if err != nil {
    return err
  }

  trainingData2, err := c.loadTrainDataSet(trainDataSetFile2, 2, 3, false)
  if err != nil {
    return err
  }

  emoticons, err := c.loadTrainDataSet(emonticonsFile, 1, 0, true)
  if err != nil {
    return err
  }

  err = c.createBagOfWords(stopWordsFile, 5, 1000, trainingData1, trainingData2, emoticons)
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