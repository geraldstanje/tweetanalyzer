package main

import (
  "github.com/white-pony/go-fann"
  "fmt"
)

func main() {
  const numLayers = 3
  const desiredError = 0.00001
  const maxEpochs = 500000
  const epochsBetweenReports = 1000

  ann := fann.CreateStandard(numLayers, []uint32{2, 3, 1})

  ann.SetTrainingAlgorithm(fann.TRAIN_INCREMENTAL)
  ann.SetActivationFunctionHidden(fann.GAUSSIAN_SYMMETRIC)
  ann.SetActivationFunctionOutput(fann.GAUSSIAN_SYMMETRIC)
  ann.SetLearningMomentum(0.4)
  
  trainInput := [][]fann.FannType{{-1.0, -1.0}, {-1.0, 1.0}, {1.0, -1.0}, {1.0, 1.0}}
  trainOutput := [][]fann.FannType{{-1.0}, {1.0}, {1.0}, {-1.0}}

  for i := 1; i <= maxEpochs; i++ {
    ann.ResetMSE()

    for i, _ := range trainInput {
      ann.Train(trainInput[i], trainOutput[i])
    }

    if ann.GetMSE() < desiredError {
      break
    }
  }

  fmt.Println("Testing network")

  test_data := fann.ReadTrainFromFile("xor.test")

  ann.ResetMSE()

  var i uint32
  for i = 0; i < test_data.Length(); i++ {
    ann.Test(test_data.GetInput(i), test_data.GetOutput(i))
  }

  fmt.Printf("MSE error on test data: %f\n", ann.GetMSE())


  fmt.Println("Saving network.");
  ann.Save("robot_float.net")
  fmt.Println("Cleaning up.")

  input := []fann.FannType{-1.0, 1.0}
  output := ann.Run(input)
  fmt.Printf("xor test (%f,%f) -> %f\n", input[0], input[1], output[0])

  test_data.Destroy()
  ann.Destroy()
}
