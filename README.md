frequency
=========

A library for frequency analysis of UTF-8 encoded text.

You feed the analyzer a corpus of text, be it English, source code, etc.
You can then score strings or byte arrays of content on a scale 0 - 1, depending on
how close they match the frequency and character occurance of the original corpus.

When training the analyzer, a larger corpus is better than a small one.
When scoring data, a larger input is better than a small one.

Install
=======
```
$ go get github.com/solvip/frequency
```

Example
=======
An example program can be found at https://github.com/solvip/frequency-english

Usage
=====

Initializing and training the analyzer, where corpus.txt is a file containing your reference corpus
```
contents, _ := ioutil.ReadFile("corpus.txt")

a := frequency.New()
a.Feed(contents)
```

Scoring content
```
score := a.ScoreString("This is a piece of english text.  The quick brown fox jumped over the lazy dog")
```

Saving the analyzer state
```
a.Save("/path/to/a/file.gob")
```

Restoring the analyzer state
```
a := frequency.New()
err := a.Restore("/path to/previously saved state")
if err != nil {
   ... maybe load data from corpus?  
}
```

License
=======
Copyright 2014 Sölvi Páll Ásgeirsson
The MIT License