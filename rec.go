package rec

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// TODO hashmaps are probably not the most memory-efficient way to do this. Consider binary search

type Matrix struct {
	Rows map[int]Row
}

func NewMatrix() *Matrix {
	return &Matrix{make(map[int]Row)}
}

type Row map[int]float32

func (r Row) String() string {
	keys := make([]int, 0, len(r))
	for item, _ := range r {
		keys = append(keys, item)
	}
	sort.Ints(keys) // Output in sorted item ID order
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for _, item := range keys {
		rating := r[item]
		buf.WriteString(fmt.Sprintf("%d: %.1f,\n", item, rating))
	}
	buf.WriteString("}\n")
	return buf.String()
}

// type IdxVal struct {
// 	Pos int
// 	Val float32
// }

// type Rec struct {

// }

type Rec struct {
	Matrix *Matrix
}

func NewRec() *Rec {
	return &Rec{
		NewMatrix(),
	}
}

func (m *Rec) AddRating(user int, item int, rating float32) {
	row, ok := m.Matrix.Rows[user]
	if !ok {
		row = make(Row)
		m.Matrix.Rows[user] = row
	}
	row[item] = rating
}

func (r *Rec) NormalizeUsers() {
	for _, row := range r.Matrix.Rows {
		min := float32(math.Inf(1))
		max := float32(math.Inf(-1))
		for _, rating := range row {
			if rating < min {
				min = rating
			}
			if rating > max {
				max = rating
			}
		}

		for item, rating := range row {
			row[item] = scale(min, max, rating)
		}
	}
}

func (r *Rec) GetRating(user, item int) (rating float32, ok bool) {
	row, ok := r.Matrix.Rows[user]
	if !ok {
		return 0, false
	}
	rating, ok = row[item]
	if !ok {
		return 0, false
	}
	return rating, ok
}

const (
	// An item must be liked by this many neighbors to be recommended.
	support = 40

	// A neighbor must rate an item at least this much for it to count as a supporting vote.
	// In the range [-1,1]
	likeThreshold = float32(0.1)
)

// Predict a rating by averaging the ratings of the nearest neighbors who have
// rated it.
func (r *Rec) PredictRating(user int, item int) (_ float32, ok bool) {
	neighbors := r.nearestNeighbors(user)

	var sum float32
	var count int
	for _, neighbor := range neighbors {
		neighborRatings := r.Matrix.Rows[neighbor.user]
		rating, ok := neighborRatings[item]
		if !ok {
			continue
		}
		count++
		sum += rating
		if count == support {
			return sum / float32(count), true
		}
	}
	return 0, false // Not enough data to make a prediction
}

// User similarity-based collaborative filtering: given a user, find similar
// users and recommend the things that they like.
func (r *Rec) UserCoFilter(user int, count int) (items []int, predicted []float32) {
	neighbors := r.nearestNeighbors(user)

	// TODO time and space efficiency

	userRow := r.Matrix.Rows[user]
	candidateCounts := make(map[int]int)
	candidateSums := make(map[int]float32)
	for _, neighbor := range neighbors {
		for item, rating := range r.Matrix.Rows[neighbor.user] {
			if _, ok := userRow[item]; ok {
				continue // user has already rated this item, can't recommend
			}
			if rating < likeThreshold {
				continue
			}
			candidateCounts[item] += 1
			candidateSums[item] += rating
			if candidateCounts[item] == support {
				items = append(items, item)
				if len(items) == count {
					predicted = make([]float32, count)
					for i, item := range items {
						predicted[i] = candidateSums[item] / float32(candidateCounts[item])
					}
					return
				}
			}
		}
	}

	return nil, nil
	// fmt.Printf("!!!!!!!!!!!!!neighbors of: %s\n", r.Matrix.Rows[user])
	// for i := 0; i < 3; i++ {
	// 	simPair := neighbors[i]
	// 	fmt.Printf("=========== neighbor %d similarity=%.1f:\n%s\n", i,
	// 		simPair.similarity, r.Matrix.Rows[simPair.user])
	// }
	// panic("intentional")
}

func (r *Rec) nearestNeighbors(user int) []simPair {
	// TODO time and space efficiency
	neighbors := make(simPairSlice, 0, len(r.Matrix.Rows)-1)
	for otherUser, _ := range r.Matrix.Rows {
		if otherUser == user {
			continue
		}
		sim := r.cosineSimilarity(user, otherUser)
		neighbors = append(neighbors, simPair{sim, otherUser})
	}

	sort.Sort(sort.Reverse(neighbors))
	return neighbors
}

// A pair containing a user ID and its similarity to some other implicit user.
type simPair struct {
	similarity float32
	user       int
}

func (s *simPair) String() string {
	return fmt.Sprintf("{similarity=%.1f user=%d}", s.similarity, s.user)
}

type simPairSlice []simPair

// For sort.Interface
func (s simPairSlice) Less(i, j int) bool {
	return s[i].similarity < s[j].similarity
}

// For sort.Interface
func (s simPairSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// For sort.Interface
func (s simPairSlice) Len() int {
	return len(s)
}

// The formula for cosine similarity can be found at http://en.wikipedia.org/wiki/Cosine_similarity
// It's the sum of the elementwise products of the two vectors, divided by the product of their
// euclidean norms.
func (r *Rec) cosineSimilarity(user1, user2 int) float32 {
	// itemsUnion := make([]int, 0, intMax(len(user1), len(user2)))
	// intSlice := make(sort.IntSlice)
	// intSlice.

	// itemsProcessed := make(map[int]struct{}) // a set, an entry implies set membership

	u1Row := r.Matrix.Rows[user1]
	u2Row := r.Matrix.Rows[user2]

	var numerator, u1SumSquares, u2SumSquares float32
	for item, u1Rating := range u1Row {
		if u2Rating, ok := u2Row[item]; ok {
			numerator += u1Rating * u2Rating
			// itemsProcessed[item] = struct{}{}
		}
		u1SumSquares += u1Rating * u1Rating
	}

	// TODO consider using user avg rating instead of skipping unrated items?

	for _, u2Rating := range u2Row {
		u2SumSquares += u2Rating * u2Rating
	}

	u1Norm := math.Sqrt(float64(u1SumSquares))
	u2Norm := math.Sqrt(float64(u2SumSquares))
	cosSim := numerator / float32(u1Norm*u2Norm)
	return float32(math.Abs(float64(cosSim)))

	// // itemsUnion := make(map[int]bool, intMax(len(user1), len(user2)))
	// for _, user := range []Row{user1, user2} {
	// 	for item, _ := range user {
	// 		itemsUnion[item] = true
	// 	}
	// }

	// for item, _ := range itemsUnion {
	// 	user1Rating, ok := user1[item]
	// 	if !ok {
	// 		continue
	// 	}
	// 	user2Rating, ok := user2[item]
	// 	if !ok {
	// 		continue
	// 	}
	// 	numerator += user1Rating * user2Rating
	// }

	// user1Norm

}

func LoadMovieLens(ratingsDatFile string, rec *Rec, trainOrTest bool) error {
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
		if line == "" {
			break
		}

		// We return 80% of the data for training, or 20% for testing. These
		// two data sets are mutually exclusive.
		isTestData := sha1HashMod(line, 10) < 2
		// fmt.Printf("isTestData=%v\n", isTestData)
		if isTestData && trainOrTest {
			continue
		} else if !isTestData && !trainOrTest {
			continue
		}

		tokens := strings.Split(line, "::")
		// fmt.Printf("Line is %s, tokens are %s\n", line, tokens)

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

func sha1HashMod(s string, mod int) int {
	hash := sha1.New()
	hash.Write([]byte(s))
	hashBytes := hash.Sum(nil)
	asU64 := binary.BigEndian.Uint64(hashBytes)
	hashMod := int(asU64 % (uint64(mod)))
	// fmt.Printf("Hash of %s: %d\n", s, hashMod)
	return hashMod
}

func intMax(i1, i2 int) int {
	if i1 >= i2 {
		return i1
	}
	return i2
}

// Returns a number in the range [-1.0,1.0]. Returns 0 if min==max.
func scale(min, max, val float32) float32 {
	rangeSize := max - min
	if rangeSize == 0 {
		return 0
	}
	fracOfRange := (val - min) / rangeSize
	return -1 + (fracOfRange)*2
}
