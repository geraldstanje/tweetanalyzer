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
  AfinnLexicon Dict
  BingLiuLexicon Dict
  MpqaLexicon Dict
  NrcEmotionLexicon Dict
}

type SentimentData struct {
	sentimentLabel int // negative == -1, neutral == 0, positive == 1
	text           string
}

// ~2,500 words
func loadAfinnLexicon(filename string) (Dict, error) {
  dict := make(Dict)

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

      if (len(word) == 2) {
        dict.Add(word[0], stringToInt(word[1]))
      }
    }
  }

  return dict, err
}

// ~6,800 words
func loadBingLiuLexicon(filename1 string, filename2 string) (Dict, error) {
  dict := make(Dict)

  f1, err := os.Open(filename1)
  if err != nil {
    fmt.Println("error opening file ", err)
    return dict, err
  }
  f2, err := os.Open(filename2)
  if err != nil {
    fmt.Println("error opening file ", err)
    return dict, err
  }
  defer f1.Close()
  defer f2.Close()
  r1 := bufio.NewReader(f1)
  for {
    s, err := r1.ReadString('\n')
    if err == io.EOF {
      // do something here
      break
    } else if err != nil {
      return dict, err // if you return error
    } else {
      s = s[0 : len(s)-1] // remove '\n'
      s = strings.ToLower(s)

      dict.Add(s, 1)
    }
  }
  r2 := bufio.NewReader(f2)
  for {
    s, err := r2.ReadString('\n')
    if err == io.EOF {
      // do something here
      break
    } else if err != nil {
      return dict, err // if you return error
    } else {
      s = s[0 : len(s)-1] // remove '\n'
      s = strings.ToLower(s)

      dict.Add(s, -1)
    }
  }

  return dict, err
}

// ~8,000 words
func loadMpqaLexicon(filename string) (Dict, error) {
  dict := make(Dict)

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

      word := strings.Split(s, " ")

      if (len(word) == 6) {
        word1 := strings.Split(word[2], "=")
        polarity := strings.Split(word[5], "=")

        if polarity[1] == "positive" {
          dict.Add(word1[1], 1)
        } else if polarity[1] == "negative" {
          dict.Add(word1[1], -1)
        }
      }
    }
  }

  return dict, err
}

// ~14,000 words
func loadNrcEmotionLexicon(filename string) (Dict, error) {
  dict := make(Dict)

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

      // eight emotions (anger, fear, anticipation, trust, surprise, sadness, joy, or disgust) or 
      // one of two polarities (negative or positive).

      if (len(word) == 3) {
        if word[2] == "1" && word[1] == "negative" {
          dict.Add(word[0], -1)
        } else if word[2] == "1" && word[1] == "positive" {
          dict.Add(word[0], 1)
        }
      }
    }
  }
  
  return dict, err
}

const (
  numHappyEmoticon int = iota
  numSadEmoticon
  numUserToken
  numHashtags
  numUrlTokens
  totalScoreAfinn
  maxPosScoreAfinn
  maxNegScoreAfinn
  numPosScoreAfinn
  numNegScoreAfinn
  totalScoreBingLiu
  totalScoreMpqa
  totalScoreNrcEmotion
)

func (c *SvmClassifier) createFeatureVector(text string) []float64 {
  tokens := c.tokenizer.Tokenize(text)

  featureVec := make([]float64, 13)

  for _, str := range tokens {
    if c.tokenizer.IsNormalizedToken(str) {
      if str == happyToken {
        featureVec[numHappyEmoticon]++
      } else if str == sadToken {
        featureVec[numSadEmoticon]++
      } else if str == userToken {
        featureVec[numUserToken]++
      } else if str == hashtagToken {
        featureVec[numHashtags]++
      } else if str == urlToken {
        featureVec[numUrlTokens]++
      }
    } 
    
    if score, ok := c.AfinnLexicon[str]; ok {
      featureVec[totalScoreAfinn] += float64(score)

      if score > int(featureVec[maxPosScoreAfinn]) {
        featureVec[maxPosScoreAfinn] = float64(score)
      } else if score < int(featureVec[maxNegScoreAfinn]){
        featureVec[maxNegScoreAfinn] = float64(score)
      }
    }

    if score2, ok := c.BingLiuLexicon[str]; ok {
      featureVec[totalScoreBingLiu] += float64(score2)
    }

    if score3, ok := c.MpqaLexicon[str]; ok {
      featureVec[totalScoreMpqa] += float64(score3)
    }

    if score4, ok := c.NrcEmotionLexicon[str]; ok {
      featureVec[totalScoreNrcEmotion] += float64(score4)

      if score4 > 0 {
        featureVec[numPosScoreAfinn]++
      } else {
        featureVec[numNegScoreAfinn]++
      }
    }
  }

  fmt.Println("feature vec:", featureVec)

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
      //dict[token] = dict[token] + 1

			if c.tokenizer.IsNormalizedToken(token) {
				dict[token] = dict[token] + 1
        //dict[token] = 100
			} /*else if len(token) > 1 {
				dict[token] = dict[token] + 1
			}*/
		}
	}

	for _, sentence := range s2 {
		tokens := c.tokenizer.Tokenize(sentence.text)

		for _, token := range tokens {
      //dict[token] = dict[token] + 1

			if c.tokenizer.IsNormalizedToken(token) {
				dict[token] = dict[token] + 1
        //dict[token] = 100
			} /*else if len(token) > 1 {
				dict[token] = dict[token] + 1
			}*/
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
	for key, _ := range wordFreq {
		//if value >= freqMin && value <= freqMax {
			c.bagOfWords.Add(key, counter)
			counter++
		//}
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

func (c *SvmClassifier) LoadClassifier(trainDataSetFile1 string, trainDataSetFile2 string, svmModelFile string) error {
	var err error

  // the following code create the bag of words
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

	c.model, err = gosvm.LoadModel(svmModelFile)
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
  
  AfinnLexicon, err := loadAfinnLexicon("data/AFINN-111.txt")
  if err != nil {
    return nil, err
  }

  BingLiuLexicon, err := loadBingLiuLexicon("data/positive-words.txt", "data/negative-words.txt")
  if err != nil {
    return nil, err
  }

  MpqaLexicon, err := loadMpqaLexicon("data/subjclueslen1-HLTEMNLP05.tff")
  if err != nil {
    return nil, err
  }

  NrcEmotionLexicon, err := loadNrcEmotionLexicon("data/NRC-emotion-lexicon-wordlevel-alphabetized-v0.92.txt")
  if err != nil {
    return nil, err
  }

	return &SvmClassifier{tokenizer: tokenizer, AfinnLexicon: AfinnLexicon, BingLiuLexicon: BingLiuLexicon, MpqaLexicon: MpqaLexicon, NrcEmotionLexicon: NrcEmotionLexicon}, err
}
