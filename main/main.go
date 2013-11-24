package main

import (
	recpkg ".."
	"flag"
	"fmt"
	"math"
	"os"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Missing filename on command line\n")
		flag.Usage()
		os.Exit(1)
	}

	filename := flag.Args()[0]

	fmt.Printf("Loading training set\n")
	trainData := recpkg.NewRec()
	if err := recpkg.LoadMovieLens(filename, trainData, true); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Loading test set\n")
	testData := recpkg.NewRec()
	if err := recpkg.LoadMovieLens(filename, testData, false); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	// fmt.Printf("Some input row: %s\n", rec.Matrix.Rows[userId])
	trainData.NormalizeUsers()
	testData.NormalizeUsers()
	// fmt.Printf("After normalizing: %s\n", rec.Matrix.Rows[userId])

	var errSumSq float64
	count := 0
	for userId, row := range testData.Matrix.Rows {
		for item, actualRating := range row {
			predictedRating, ok := trainData.PredictRating(userId, item)
			if !ok {
				fmt.Printf("Prediction failed\n")
			} else {
				err := float64(actualRating) - float64(predictedRating)
				errSumSq += math.Pow(err, 2)
				count++
				rmse := math.Sqrt(errSumSq / float64(count))
				fmt.Printf("User=%d item=%d predicted=%f actual=%f rmse=%f\n",
					userId, item, predictedRating, actualRating, rmse)
			}
		}
	}

	// var userId int
	// for tmpUserId, _ := range trainData.Matrix.Rows {
	// 	userId = tmpUserId
	// 	break
	// }

	// items, scores := trainData.UserCoFilter(userId, 3)
	// if items == nil {
	// 	fmt.Printf("Prediction failed\n")
	// 	os.Exit(1)
	// }

	// for i := 0; i < len(items); i++ {
	// 	item := items[i]
	// 	predicted := scores[i]
	// 	actual, ok := testData.GetRating(userId, item)
	// 	if !ok {
	// 		fmt.Printf("No rating available for item %d\n", item)
	// 	} else {
	// 		fmt.Printf("Prediction/actual for %d: %f/%f\n", item, predicted, actual)
	// 	}
	// }

	// fmt.Printf("Recommendations for %d: %d, predicted scores %s\n", userId, items, scores)
}
