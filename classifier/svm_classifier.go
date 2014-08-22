package classifier

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/sridif/gosvm"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const dump_bag_of_words_to_file = true

type SvmClassifier struct {
	tokenizer  *Tokenizer
	model      *gosvm.Model
	bagOfWords Dict
}

type SentimentData struct {
	sentimentLabel int // negative == -1, neutral == 0, positive == 1
	text           string
}

func (c *SvmClassifier) createFeatureVector(text string) []float64 {
	featureVec := make([]float64, len(c.bagOfWords))

	tokens := c.tokenizer.Tokenize(text)

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

	if sentiment == "positive" ||
		sentiment == "extremely-positive" ||
		sentiment == "neutral" ||
		sentiment == "negative" ||
		sentiment == "extremely-negative" {
		var i int
		var str string

		if sentiment == "positive" ||
			sentiment == "extremely-positive" {
			i = 1
		} else if sentiment == "negative" ||
			sentiment == "extremely-negative" {
			i = -1
		} else if sentiment == "neutral" {
			i = 0
		}

		str = word[index2]
		t = SentimentData{i, str}
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
	dict := make(Dict)

	for _, sentence := range s1 {
		tokens := c.tokenizer.Tokenize(sentence.text)

		for _, token := range tokens {
			if c.tokenizer.IsNormalizedToken(token) {
				dict[token] = 100
			} else if len(token) > 1 {
				dict[token] = dict[token] + 1
			}
		}
	}

	for _, sentence := range s2 {
		tokens := c.tokenizer.Tokenize(sentence.text)

		for _, token := range tokens {
			if c.tokenizer.IsNormalizedToken(token) {
				dict[token] = 100
			} else if len(token) > 1 {
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

	if dump_bag_of_words_to_file {
		for word, _ := range c.bagOfWords {
			buffer.WriteString(word)
			buffer.WriteString("\n")
		}
	}

	if dump_bag_of_words_to_file {
		err = ioutil.WriteFile("bagOfWords.txt", buffer.Bytes(), 0644)
	}
	return err
}

func (c *SvmClassifier) TrainClassifier(trainDataSetFile1 string, trainDataSetFile2 string) error {
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

	err = c.createBagOfWords(5, 800, trainingData1, trainingData2)
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
	fmt.Println("Error:", float64(flailCounter)/float64(counter))

	elapsed := time.Now().Sub(start)
	fmt.Println(elapsed)
	fmt.Println("Test finished!")

	return nil
}

func NewSvmClassifier() (*SvmClassifier, error) {
	tokenizer, err := NewTokenizer()
	if err != nil {
		return nil, err
	}
	return &SvmClassifier{tokenizer: tokenizer}, err
}
