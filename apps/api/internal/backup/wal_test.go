package backup

import (
	"reflect"
	"testing"
)

func TestParseWALSegment(t *testing.T) {
	seg, err := ParseWALSegment("000000010000000000000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seg.Timeline != 1 || seg.Log != 0 || seg.Segment != 1 {
		t.Fatalf("got %+v", seg)
	}
	if seg.String() != "000000010000000000000001" {
		t.Fatalf("round trip = %q", seg.String())
	}
}

func TestParseWALSegmentRejectsBadNames(t *testing.T) {
	bad := []string{
		"",
		"00000001",
		"00000001000000000000000Z",
		"000000010000000000000001.archive",
	}
	for _, name := range bad {
		if _, err := ParseWALSegment(name); err == nil {
			t.Errorf("expected %q to be rejected", name)
		}
	}
}

func TestWALSegmentNext(t *testing.T) {
	// 16MB segments: 256 per 4GiB log.
	perLog := segmentsPerLog(16 * 1024 * 1024)
	if perLog != 256 {
		t.Fatalf("segmentsPerLog = %d, want 256", perLog)
	}

	seg := WALSegment{Timeline: 1, Log: 0, Segment: 5}
	if got := seg.Next(perLog); got.Segment != 6 || got.Log != 0 {
		t.Fatalf("next = %v, want log 0 segment 6", got)
	}

	last := WALSegment{Timeline: 1, Log: 2, Segment: 255}
	if got := last.Next(perLog); got.Segment != 0 || got.Log != 3 {
		t.Fatalf("rollover = %v, want log 3 segment 0", got)
	}
}

func TestFindGaps(t *testing.T) {
	perLog := segmentsPerLog(16 * 1024 * 1024)

	t.Run("contiguous segments have no gaps", func(t *testing.T) {
		segments := []WALSegment{
			{1, 0, 1}, {1, 0, 2}, {1, 0, 3},
		}
		if gaps := FindGaps(segments, 16*1024*1024); len(gaps) != 0 {
			t.Fatalf("expected no gaps, got %v", gaps)
		}
	})

	t.Run("a missing middle segment is detected", func(t *testing.T) {
		segments := []WALSegment{
			{1, 0, 1}, {1, 0, 4},
		}
		gaps := FindGaps(segments, 16*1024*1024)
		want := []string{
			WALSegment{1, 0, 2}.String(),
			WALSegment{1, 0, 3}.String(),
		}
		if !reflect.DeepEqual(gaps, want) {
			t.Fatalf("gaps = %v, want %v", gaps, want)
		}
	})

	t.Run("rollover to next log is not a gap", func(t *testing.T) {
		segments := []WALSegment{
			{1, 0, uint32(perLog - 1)},
			{1, 1, 0},
		}
		if gaps := FindGaps(segments, 16*1024*1024); len(gaps) != 0 {
			t.Fatalf("rollover should not be a gap, got %v", gaps)
		}
	})
}

func TestSortWALSegments(t *testing.T) {
	segments := []WALSegment{
		{1, 0, 3}, {1, 0, 1}, {2, 0, 0}, {1, 1, 0},
	}
	SortWALSegments(segments)

	want := []WALSegment{
		{1, 0, 1}, {1, 0, 3}, {1, 1, 0}, {2, 0, 0},
	}
	if !reflect.DeepEqual(segments, want) {
		t.Fatalf("sorted = %v, want %v", segments, want)
	}
}

func TestWALPrefix(t *testing.T) {
	if got := walPrefix("prod"); got != "prod/wal" {
		t.Fatalf("got %q", got)
	}
	if got := walPrefix("prod/"); got != "prod/wal" {
		t.Fatalf("got %q", got)
	}
}
