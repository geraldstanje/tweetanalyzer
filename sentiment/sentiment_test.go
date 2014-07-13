package sentiment

import (
	"testing"
)

func TestSimple(t *testing.T) {
	s := new(SentimentAnalysis)

	s.trainClassifier("AFINN-111.txt")

	classVal := s.getClass("I blame the clever pizza cook")
	if classVal != "neutral" {
		t.Error("neutral classification failed")
	}

	classVal = s.getClass("I love the brilliant pizza")
	if classVal != "positive" {
		t.Error("positive classification failed")
	}

	classVal = s.getClass("I hate the fucking pizza")
	if classVal != "negative" {
		t.Error("negative classification failed")
	}

	classVal = s.getClass("I like the pizza")
	if classVal != "positive" {
		t.Error("positive classification failed")
	}
}
