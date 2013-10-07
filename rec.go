package rec

import (
	"bytes"
	"fmt"
	"math"
	"sort"
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

const (
	// An item must be liked by this many neighbors to be recommended.
	support = 5

	// A neighbor must rate an item at least this much for it to count as a supporting vote.
	// In the range [-1,1]
	likeThreshold = float32(0.1)
)

func (r *Rec) Recommend(user int, count int) (items []int, predicted []float32) {
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
