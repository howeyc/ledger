Go Period
============

[![Author](http://img.shields.io/badge/author-@studiofrenetic-blue.svg?style=flat-square)](https://twitter.com/studiofrenetic)
[![GoDoc](https://godoc.org/github.com/studiofrenetic/period?status.svg)](https://godoc.org/github.com/studiofrenetic/period)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE.md)


Period provides a set of missing Time Range to Go, it cover all basic operations regardings time ranges.
A Go port of [thephpleague/period](https://github.com/thephpleague/period) based on [Resolving Feature Envy in the Domain](http://verraes.net/2014/08/resolving-feature-envy-in-the-domain/) by Mathias Verraes and extends the concept to cover all basic operations regarding time ranges.

## Highlights

- Treats Time Range as time objects
- Exposes many named constructors to ease time range creation
- Covers all basic manipulations related to time range

Install `Period` using go get.

```
$ go get github.com/studiofrenetic/period
```

Doc
-------
https://godoc.org/github.com/studiofrenetic/period

Examples
-------
```Go
// Accessing time range properties
p, err := period.CreateFromMonth(2015, 3) // {2015-01-12 00:00:00 +0000 UTC 2015-01-19 00:00:00 +0000 UTC}
if err != nil {
    log.Fatal(err)
}

// Comparing time ranges
p, err := period.CreateFromWeek(2015, 3)
if err != nil {
    log.Fatal(err)
}
alt := period.CreateFromDuration(time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC), time.Duration(24*7)*time.Hour) // {2015-01-14 00:00:00 +0000 UTC 2015-01-21 00:00:00 +0000 UTC}
sameDuration := p.SameDurationAs(alt) // true

p := period.CreateFromDuration(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), (time.Duration(2) * time.Hour))
shouldContains := time.Date(2015, 1, 1, 0, 30, 0, 0, time.UTC)
contains := p.Contains(shouldContains)

// Modifying time ranges
p.Next()
```

Testing
-------

```bash
$ go test -v
```

Roadmap
-------
#### Constructor
- [x] Period{}
- [X] CreateFromYear
- [x] CreateFromSemester
- [x] CreateFromQuarter
- [x] CreateFromMonth
- [X] CreateFromWeek
- [X] CreateFromDuration
- [X] CreateFromDurationBeforeEnd

## Comparing Periods
#### Comparing endpoints
- [X] Contains(another_period)
- [X] Overlaps
- [X] SameValueAs
- [X] Abuts
- [X] IsBefore
- [X] IsAfter
- [X] Diff

#### Comparing durations
- [X] DurationGreaterThan
- [X] DurationLessThan
- [X] SameDurationAs
- [X] CompareDuration
- [X] DurationDiff
- [X] TimestampDurationDiff

## Modifying Periods
#### Using endpoints
- [X] StartingOn
- [X] EndingOn

#### Using duration
- [X] WithDuration
- [X] Add
- [X] Sub
- [X] Next
- [X] Previous

#### Using Period objects
- [X] Merge
- [X] Intersect
- [X] Gap

#### Doc
- [ ] API
- [ ] Accessing time range properties
- [ ] Iterate over a time range
- [ ] Comparing time ranges
- [ ] Modifying time ranges

------

Contributing
============

Please feel free to submit issues, fork the repository and send pull requests!

Contributions are welcome and will be fully credited. Please see [CONTRIBUTING](CONTRIBUTING.md) for details.

------

Licence
=======
Copyright (c) 2015 Studio Frenetic

Please consider promoting this project if you find it useful.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
