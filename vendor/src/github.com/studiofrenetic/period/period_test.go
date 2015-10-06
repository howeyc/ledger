package period

import (
	"github.com/stretchr/testify/assert"
	// "github.com/studiofrenetic/period"
	// "fmt"
	"testing"
	"time"
)

func TestCreateFromWeek(t *testing.T) {
	StartWeek = time.Monday
	p, err := CreateFromWeek(2015, 1)
	if err != nil {
		assert.Fail(t, "Error create period from week", "OutOfRangeError")
	}

	startRequested, _ := time.Parse(YMDHIS, "2014-12-29 00:00:00")
	endRequested, _ := time.Parse(YMDHIS, "2015-01-05 00:00:00")

	assert.Equal(t, p.Start, startRequested, "they should be equal")
	assert.Equal(t, p.End, endRequested, "they should be equal")
}

func TestCreateFromMonth(t *testing.T) {
	p, err := CreateFromMonth(2015, 3)
	if err != nil {
		assert.Fail(t, "Error create period from week", "OutOfRangeError")
	}

	startRequested, _ := time.Parse(YMDHIS, "2015-03-01 00:00:00")
	endRequested, _ := time.Parse(YMDHIS, "2015-04-01 00:00:00")

	assert.Equal(t, p.Start, startRequested, "they should be equal")
	assert.Equal(t, p.End, endRequested, "they should be equal")
}

func TestCreateFromQuarter(t *testing.T) {
	p, err := CreateFromQuarter(2015, 1)
	if err != nil {
		assert.Fail(t, "Error create period from quarter", "OutOfRangeError")
	}

	expected := Period{
		time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2015, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	assert.Equal(t, expected, p)
}

func TestCreateFromSemester(t *testing.T) {
	p, err := CreateFromSemester(2015, 1)
	if err != nil {
		assert.Fail(t, "Error create period from semester", "OutOfRangeError")
	}

	expected := Period{
		time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2015, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	assert.Equal(t, expected, p)
}

func TestCreateFromDuration(t *testing.T) {
	from := time.Date(2015, 5, 25, 0, 0, 0, 0, time.UTC)
	duration := time.Duration(24*7) * time.Hour

	p := CreateFromDuration(from, duration)

	expected := Period{
		time.Date(2015, 5, 25, 0, 0, 0, 0, time.UTC),
		time.Date(2015, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	assert.Equal(t, expected, p)
}

func TestCreateFromDurationBeforeEnd(t *testing.T) {

	from := time.Date(2015, 5, 25, 0, 0, 0, 0, time.UTC)
	duration := time.Duration(24*7) * time.Hour

	p := CreateFromDurationBeforeEnd(from, duration)

	expected := Period{
		time.Date(2015, 5, 18, 0, 0, 0, 0, time.UTC),
		time.Date(2015, 5, 25, 0, 0, 0, 0, time.UTC),
	}

	assert.Equal(t, expected, p)
}

func TestNext(t *testing.T) {
	from := time.Date(2015, 5, 25, 0, 0, 0, 0, time.UTC)
	duration := time.Duration(24*2) * time.Hour

	p := CreateFromDuration(from, duration)

	p.Next()

	expected := Period{
		time.Date(2015, 5, 27, 0, 0, 0, 0, time.UTC),
		time.Date(2015, 5, 29, 0, 0, 0, 0, time.UTC),
	}

	assert.Equal(t, expected, p)
}

func TestPrevious(t *testing.T) {
	from := time.Date(2015, 5, 25, 0, 0, 0, 0, time.UTC)
	duration := time.Duration(24*2) * time.Hour

	p := CreateFromDuration(from, duration)

	p.Previous()

	expected := Period{
		time.Date(2015, 5, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2015, 5, 25, 0, 0, 0, 0, time.UTC),
	}

	assert.Equal(t, expected, p)
}

func TestDiff(t *testing.T) {
	period, err := CreateFromYear(2013)
	if err != nil {
		t.Fatalf("err : %v\n", err)
	}
	t.Logf("period: %s", period)

	alt := CreateFromDuration(time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC), time.Duration(24*7)*time.Hour)
	if err != nil {
		t.Fatalf("err : %v\n", err)
	}
	t.Logf("alt: %s", alt)

	diff, err := alt.Diff(period)
	if err != nil {
		t.Fatalf("err : %v\n", err)
	}

	t.Logf("diff: %s", diff)
}

func TestSameValueAs(t *testing.T) {
	from := time.Date(2015, 5, 25, 0, 0, 0, 0, time.UTC)
	duration := time.Duration(24*2) * time.Hour

	p := CreateFromDuration(from, duration)

	altPeriod := CreateFromDuration(from, duration)

	var isSame bool = p == altPeriod

	t.Logf("p: %s, is same: %t", p, isSame)

}

func TestDiffWithEqualsPeriod(t *testing.T) {
	period, _ := CreateFromYear(2015)
	alt := CreateFromDuration(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), (time.Duration(365*24) * time.Hour))
	diff, _ := alt.Diff(period)
	t.Logf("diff: %v\nalt: %v", diff, alt)
}

func TestDiffWithPeriodSharingOneEndpoints(t *testing.T) {
	period, _ := CreateFromYear(2015)
	alt := CreateFromDuration(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), (time.Duration(90*24) * time.Hour))
	diff, _ := alt.Diff(period)
	t.Logf("%#v", diff)
}

func TestDiffWithOverlapsPeriod(t *testing.T) {
	period := CreateFromDuration(time.Date(2015, 1, 1, 10, 0, 0, 0, time.UTC), (time.Duration(3) * time.Hour))
	alt := CreateFromDuration(time.Date(2015, 1, 1, 11, 0, 0, 0, time.UTC), (time.Duration(3) * time.Hour))
	diff, _ := alt.Diff(period)

	assert.Len(t, diff, 2, "The size of slice is not 2")
	assert.Equal(t, (time.Duration(1) * time.Hour), diff[0].GetDurationInterval(), "first diff should be 1 hour as time.Duration")
	assert.Equal(t, (time.Duration(1) * time.Hour), diff[1].GetDurationInterval(), "second diff should be 1 hour as time.Duration")

}

func TestDurationDiff(t *testing.T) {
	period := CreateFromDuration(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), (time.Duration(1) * time.Hour))
	alt := CreateFromDuration(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), (time.Duration(2) * time.Hour))
	diff := period.DurationDiff(alt)

	assert.Equal(t, (time.Duration(-1) * time.Hour), diff, "Should be 1 hour diff")
}

func TestContains(t *testing.T) {
	period := CreateFromDuration(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), (time.Duration(2) * time.Hour))
	shouldContains := time.Date(2015, 1, 1, 0, 30, 0, 0, time.UTC)
	contains := period.Contains(shouldContains)

	assert.Equal(t, true, contains, "Should be true")
}

func TestBefore(t *testing.T) {
	period := CreateFromDuration(time.Date(2015, 1, 1, 13, 0, 0, 0, time.UTC), (time.Duration(2) * time.Hour))
	alt := CreateFromDuration(time.Date(2015, 1, 1, 15, 30, 0, 0, time.UTC), (time.Duration(2) * time.Hour))

	isBefore := period.Before(alt)

	assert.Equal(t, true, isBefore, "Should be true")
}

func TestAfter(t *testing.T) {
	period := CreateFromDuration(time.Date(2015, 1, 1, 13, 0, 0, 0, time.UTC), (time.Duration(2) * time.Hour))
	alt := CreateFromDuration(time.Date(2015, 1, 1, 1, 11, 0, 0, time.UTC), (time.Duration(2) * time.Hour))

	isAfter := period.After(alt)

	assert.Equal(t, true, isAfter, "Should be true")
}
