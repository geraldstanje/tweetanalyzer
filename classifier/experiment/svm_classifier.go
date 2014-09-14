package classifier

import (
  "bufio"
  "fmt"
  "github.com/sridif/gosvm"
  "io"
  "os"
  "strings"
  "time"
)

type SvmClassifier struct {
  model      *gosvm.Model
}

type SentimentData struct {
  sentimentLabel int // negative == -1, neutral == 0, positive == 1
  text           string
}

func (c *SvmClassifier) createFeatureVector(text string) []float64 {
  featureVec := make([]float64, 0)

  features := strings.Split(text, " ")

  for _, str := range features {
    featureVec = append(featureVec, stringToFloat(str))
  }

  return featureVec
}

func (c *SvmClassifier) loadTrainDataSet(filename string) ([]SentimentData, error) {
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

      word := strings.Split(s, ",")

      if len(word) > 1 {
        dict = append(dict, SentimentData{stringToInt(word[0]), word[1]})
      }
    }
  }

  return dict, err
}

func (c *SvmClassifier) loadTestDataSet(filename string) ([]SentimentData, error) {
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

      word := strings.Split(s, ",")

      if len(word) > 1 {
        dict = append(dict, SentimentData{stringToInt(word[0]), word[1]})
      }
    }
  }

  return dict, err
}

func (c *SvmClassifier) TrainClassifier(trainDataSetFile string) error {
  // Perform training
  fmt.Println("Start Training")
  start := time.Now()

  // train the model
  trainingData, err := c.loadTrainDataSet(trainDataSetFile)
  if err != nil {
    return err
  }

  problem := gosvm.NewProblem()

  // We use all features
  for _, val := range trainingData {
    problem.Add(gosvm.TrainingInstance{float64(val.sentimentLabel), gosvm.FromDenseVector(c.createFeatureVector(val.text))})
  }

  param := gosvm.DefaultParameters()
  param.Kernel = gosvm.NewLinearKernel()
  param.SVMType = gosvm.NewCSVC(0.1)
  c.model, err = gosvm.TrainModel(param, problem)
  if err != nil {
    return err
  }

  elapsed := time.Now().Sub(start)
  fmt.Println(elapsed)
  fmt.Println("Training finished!")

  err = c.model.Save("svm_model")
  fmt.Println("Svm model saved!")
  return err
}

func (c *SvmClassifier) LoadClassifier(filename string) error {
  var err error

  c.model, err = gosvm.LoadModel(filename)
  return err
}

func (c *SvmClassifier) ClassifyTweet(tweet string) float64 {
  label := c.model.Predict(gosvm.FromDenseVector(c.createFeatureVector(tweet)))
  return label
}

func (c *SvmClassifier) TestClassifier(testDataSetFile string) error {
  fmt.Println("Start Testing")
  start := time.Now()

  // test the model
  testData, err := c.loadTestDataSet(testDataSetFile)
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
  fmt.Println("Error:", float64(flailCounter)/float64(counter))

  elapsed := time.Now().Sub(start)
  fmt.Println(elapsed)
  fmt.Println("Test finished!")

  return nil
}

func NewSvmClassifier() (*SvmClassifier, error) {
  return &SvmClassifier{}, nil
}