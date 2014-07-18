package main

import (
	"fmt"
	"github.com/sridif/gosvm"
	"log"
)

func main() {
	problem := gosvm.NewProblem()
	problem.Add(gosvm.TrainingInstance{-1, gosvm.FromDenseVector([]float64{-1, -1})})
	problem.Add(gosvm.TrainingInstance{1, gosvm.FromDenseVector([]float64{-1, 1})})
	problem.Add(gosvm.TrainingInstance{1, gosvm.FromDenseVector([]float64{1, -1})})
	problem.Add(gosvm.TrainingInstance{-1, gosvm.FromDenseVector([]float64{1, 1})})
	param := gosvm.DefaultParameters()
	param.Kernel = gosvm.NewRBFKernel(0.1) //NewPolynomialKernel(1.0, 0.1, 1)
	model, err := gosvm.TrainModel(param, problem)
	if err != nil {
		log.Fatal(err)
	}

	label1 := model.Predict(gosvm.FromDenseVector([]float64{-1, 1}))
	fmt.Printf("Predicted label: %f\n", label1)

	label2 := model.Predict(gosvm.FromDenseVector([]float64{1, 1}))
	fmt.Printf("Predicted label: %f\n", label2)
}
