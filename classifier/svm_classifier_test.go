package classifier

import (
	"testing"
)

func TestSvmClassifier(t *testing.T) {
	classifier, err := NewSvmClassifier()
	if err != nil {
		t.Fatal(err)
	}

	err = classifier.TrainClassifier("2014_b_train.txt", "2014_b_dev.txt")
	if err != nil {
		t.Fatal(err)
	}

	/*err = classifier.LoadClassifier("svm_model")
	  if err != nil {
	    t.Fatal(err)
	  }*/

	err = classifier.TestClassifier("2014_b_test_gold.txt")
	if err != nil {
		t.Fatal(err)
	}
}
