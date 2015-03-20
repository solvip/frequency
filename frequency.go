package frequency

import (
	"encoding/gob"
	"math"
	"os"
	"sync"
)

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
