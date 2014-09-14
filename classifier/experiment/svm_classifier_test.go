package classifier

import (
	"testing"
)

func TestSvmClassifier(t *testing.T) {
	classifier, err := NewSvmClassifier()
	if err != nil {
		t.Fatal(err)
	}

	err = classifier.TrainClassifier("training.txt")
	if err != nil {
		t.Fatal(err)
	}

	err = classifier.TestClassifier("test.txt")
	if err != nil {
		t.Fatal(err)
	}
}
