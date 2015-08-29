package frequency

import (
	"encoding/gob"
	"math"
	"os"
	"sync"
)

var EnglishAnalyzer = &Analyzer{
	// Generated from the combined texts of Moby Dick, Jane Eare, and other Project Gutenberg titles
	frequency: [256]int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 132230, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1118323, 4805, 16267, 10, 17, 11, 30, 13831, 1299, 1299, 352, 0, 97346, 24497, 47923, 198, 1133, 1409, 659, 517, 471, 490, 349, 350, 504, 301, 6172, 16468, 3, 2, 3, 4411, 20, 12139, 5370, 5999, 3507, 5993, 4066, 3747, 7966, 26560, 1878, 695, 3840, 6518, 4805, 5236, 4835, 460, 3992, 8246, 16050, 1289, 960, 6071, 298, 2106, 221, 407, 0, 407, 0, 963, 0, 425022, 80895, 140227, 226517, 688599, 127083, 105382, 321942, 349210, 4730, 35065, 221740, 128556, 373375, 395487, 90412, 5787, 325314, 341986, 476299, 150327, 53345, 109065, 8599, 95926, 4605, 14, 0, 14, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	size:      6921848,
}

type Analyzer struct {
	mu        sync.RWMutex
	frequency [256]int64
	size      int64
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Feed - Feed an analyzer with contents, updating the frequency table.
// The analyzer state is updated - not replaced, so multiple Feed calls are OK.
func (a *Analyzer) Feed(contents []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update the character count in analyzer
	for _, character := range contents {
		a.frequency[character] += 1
		a.size += 1
	}

	return
}

// Score - Score contents according to the analyzer frequency tables.  Return a value in the range of 0 - 1.
func (a *Analyzer) Score(contents []byte) float64 {
	other := NewAnalyzer()
	other.Feed(contents)

	a.mu.RLock()
	defer a.mu.RUnlock()

	return scoreFrequencies(a, other)
}

// ScoreString - Score string.  Return a value in the range 0 - 1.
func (a *Analyzer) ScoreString(text string) float64 {
	return a.Score([]byte(text))
}

// Save - save the analyzer state to a file at path.
func (a *Analyzer) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	a.mu.RLock()
	defer a.mu.RUnlock()
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(a.frequency); err != nil {
		return err
	}

	return nil
}

// Restore - restore the state previously saved at path, overwriting current analyzer state
func (a *Analyzer) Restore(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	a.mu.Lock()
	defer a.mu.Unlock()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&a.frequency); err != nil {
		return err
	}

	for _, v := range a.frequency {
		a.size += v
	}

	return nil
}

// relativeDifference - the relative difference between two numbers, a and b, as a value in the range 0 -1.
func relativeDifference(a, b float64) float64 {
	if a == 0 && b == 0 {
		return 0
	}
	return math.Abs(a-b) / max(a, b)
}

func scoreFrequencies(ref, target *Analyzer) (score float64) {
	var r float64 = 0
	var t float64 = 0

	for i := 0; i < 256; i++ {
		r = float64(ref.frequency[i]) / float64(ref.size)
		t = float64(target.frequency[i]) / float64(target.size)
		score += (r * (1 - relativeDifference(r, t)))
	}

	return score
}

func max(a, b float64) float64 {
	if a > b {
		return a
	} else {
		return b
	}
}
