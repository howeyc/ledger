// Package period provides a set of missing Time Range to Go. It is based on PHP Period http://period.thephpleague.com/
// period cover all basic operations regardings time ranges.
//
// Highlights
//
// Treats Time Range as immutable value objects
// Exposes many named constructors to ease time range creation
// Covers all basic manipulations related to time range
// Fully documented
// Framework-agnostic
//
// Questions?
//
// studiofrenetic/period was created by Studio Frentic. Find us on Twitter at @StudioFrenetic
//
// Examples
//
//     // Accessing time range properties
//     p, err := period.CreateFromMonth(2015, 3) // {2015-01-12 00:00:00 +0000 UTC 2015-01-19 00:00:00 +0000 UTC}
//     if err != nil {
//         log.Fatal(err)
//     }
//
//     // Comparing time ranges
//     p, err := period.CreateFromWeek(2015, 3)
//     if err != nil {
//         log.Fatal(err)
//     }
//     alt := period.CreateFromDuration(time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC), time.Duration(24*7)*time.Hour) // {2015-01-14 00:00:00 +0000 UTC 2015-01-21 00:00:00 +0000 UTC}
//     sameDuration := p.SameDurationAs(alt) // true
//
//     p := period.CreateFromDuration(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), (time.Duration(2) * time.Hour))
//     shouldContains := time.Date(2015, 1, 1, 0, 30, 0, 0, time.UTC)
//     contains := p.Contains(shouldContains)
//
//
//     // Modifying time ranges
//     p.Next()
//
package period
