package main

import (
	rec ".."
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Missing filename on command line\n")
		flag.Usage()
		os.Exit(1)
	}

	filename := flag.Args()[0]

	rec := rec.NewRec()
	if err := LoadMovieLens(filename, rec); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	var userId int
	for tmpUserId, _ := range rec.Matrix.Rows {
		userId = tmpUserId
	}
	// fmt.Printf("Some input row: %s\n", rec.Matrix.Rows[userId])
	rec.NormalizeUsers()
	// fmt.Printf("After normalizing: %s\n", rec.Matrix.Rows[userId])

	items, scores := rec.Recommend(userId, 3)
	if items == nil {
		fmt.Printf("Prediction failed")
	}
	fmt.Printf("Recommendations for %d: %d, predicted scores %s\n", userId, items, scores)
}

func LoadMovieLens(ratingsDatFile string, rec *rec.Rec) error {
	fh, err := os.Open(ratingsDatFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed opening input file: %s\n", err.Error())
		os.Exit(1)
	}
	defer fh.Close()

	bufRd := bufio.NewReader(fh)
	for {
		line, err := bufRd.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("Failed reading input file: %s", err.Error())
		}
		tokens := strings.Split(line, "::")
		// fmt.Printf("Line is %s, tokens are %s\n", line, tokens)
		if line == "" {
			break
		}
		user, err := strconv.Atoi(tokens[0])
		if err != nil {
			return fmt.Errorf("Invalid user ID: %s", tokens[0])
		}
		movieId, err := strconv.Atoi(tokens[1])
		if err != nil {
			return fmt.Errorf("Invalid movie ID")
		}
		rating, err := strconv.Atoi(tokens[2])
		if err != nil {
			return fmt.Errorf("Invalid rating\n")
		}

		rec.AddRating(user, movieId, float32(rating))

		if err == io.EOF {
			break
		}
	}
	return nil
}
