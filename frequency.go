package frequency

import (
	"encoding/gob"
	"math"
	"os"
	"sync"
	"unicode"
)

type Analyzer struct {
	Frequency map[rune]float64
	CCount    map[rune]float64
	Size      int64
	mutex     sync.RWMutex
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		Size:      0,
		Frequency: make(map[rune]float64),
		CCount:    make(map[rune]float64),
		mutex:     sync.RWMutex{},
	}
}

// lock - Lock analyzer for reading and writing
func (a *Analyzer) lock() {
	a.mutex.Lock()
}

// unlock - Unlock analyzer for reading and writing
func (a *Analyzer) unlock() {
	a.mutex.Unlock()
}

// rlock - Lock analyzer for reading
func (a *Analyzer) rlock() {
	a.mutex.RLock()
}

// runlock - Unlock analyzer for reading
func (a *Analyzer) runlock() {
	a.mutex.RUnlock()
}

// Feed - Feed an analyzer with contents, updating the frequency and count tables.
// The analyzer state is updated - not replaced, so multiple Feed calls are OK.
func (a *Analyzer) Feed(contents []byte) {
	a.lock()
	defer a.unlock()

	// Update the character count in analyzer
	for _, character := range contents {
		r := unicode.ToLower(rune(character))

		if val, ok := a.CCount[r]; ok {
			a.CCount[r] = val + 1
		} else {
			a.CCount[r] = 1
		}
		a.Size += 1
	}

	// Update the frequency table according to the new count.
	for k, v := range a.CCount {
		a.Frequency[k] = v / float64(a.Size)
	}

	return
}

// Score - Score contents according to the analyzer frequency tables.  Return a value in the range of 0 - 1.
func (a *Analyzer) Score(contents []byte) float64 {
	other := NewAnalyzer()
	other.Feed(contents)

	a.rlock()
	defer a.runlock()

	return scoreFrequencies(a, other) * scoreOccurances(a, other)
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

	a.rlock()
	defer a.runlock()
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(*a); err != nil {
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

	a.lock()
	defer a.unlock()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(a); err != nil {
		return err
	}

	return nil
}

// relativeDifference - the relative difference between two numbers, a and b, as a value in the range 0 -1.
func relativeDifference(a, b float64) float64 {
	// If one of a, b is negative, we shift both numbers into the positive range,
	// while keeping the difference the same.
	if a == 0 && b == 0 {
		return 0
	} else if a < 0 && b >= 0 {
		a = math.Abs(a)
		b = a + b
	} else if b < 0 && a >= 0 {
		b = math.Abs(b)
		a = a + b
	}

	return math.Abs((a - b) / math.Max(a, b))
}

func scoreFrequencies(ref, target *Analyzer) float64 {
	var score float64 = 0
	for k := range ref.Frequency {
		score += (ref.Frequency[k] *
			(1 - relativeDifference(ref.Frequency[k], target.Frequency[k])))
	}

	return score
}

// scoreOccurances - Score the occurances of characters in ref and target.
// I.e., if all the characters occuring in ref occur in target, the score is 1.
// If none of the characters occuring in ref occur in target, the score is 0.
// If characters occur in target that do not occur in ref, the score is reduced.
// This is used to defeat cases where the characters in target are a superset
// of the characters in ref, even though the frequency might be similar.
func scoreOccurances(ref, target *Analyzer) float64 {
	// The target set is empty
	if len(target.CCount) == 0 {
		return 0
	}

	return 1 - setRelativeDifference(ref.CCount, target.CCount)
}

// setEqual - considering a, b sets.  Return true if they are equal
func setEqual(a, b map[rune]float64) bool {
	if len(a) != len(b) {
		// The maps are of different lengths, the keys can't be equal
		return false
	}

	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}

	return true
}

// setIntersection - considering a, b sets.  Return the intersection of them
func setIntersection(a, b map[rune]float64) map[rune]float64 {
	res := map[rune]float64{}

	for k := range a {
		if v, ok := b[k]; ok {
			res[k] = v
		}
	}

	return res
}

// setRelativeDifference - return the relative difference of items in sets a, b as a value between 0 - 1.
func setRelativeDifference(a, b map[rune]float64) float64 {
	if setEqual(a, b) {
		return 0
	}

	ci := float64(len(setIntersection(a, b)))
	ca := float64(len(a))
	cb := float64(len(b))

	/* The intersection is empty.  The differenc must be 1. */
	if ci == 0 {
		return 1
	}

	return math.Max((ca-ci), (cb-ci)) / math.Max(ca, cb)
}
