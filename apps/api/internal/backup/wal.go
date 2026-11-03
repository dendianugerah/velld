package backup

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// WAL segment files are named TTTTTTTTXXXXXXXXYYYYYYYY:
//
//	TTTTTTTT = timeline (hex)
//	XXXXXXXX = logical log id (hex)
//	YYYYYYYY = segment number within the log (hex)
//
// Velld uses this to list, order, and detect gaps in archived segments, which
// is how it decides whether a point-in-time recovery window is intact.
type WALSegment struct {
	Timeline uint32
	Log      uint32
	Segment  uint32
}

const walSegmentNameLength = 24

// ParseWALSegment parses a 24-character WAL segment file name.
func ParseWALSegment(name string) (WALSegment, error) {
	if len(name) != walSegmentNameLength {
		return WALSegment{}, fmt.Errorf("invalid WAL segment name %q: expected %d hex characters", name, walSegmentNameLength)
	}

	timeline, err := strconv.ParseUint(name[0:8], 16, 32)
	if err != nil {
		return WALSegment{}, fmt.Errorf("invalid timeline in %q: %w", name, err)
	}
	log, err := strconv.ParseUint(name[8:16], 16, 32)
	if err != nil {
		return WALSegment{}, fmt.Errorf("invalid log id in %q: %w", name, err)
	}
	segment, err := strconv.ParseUint(name[16:24], 16, 32)
	if err != nil {
		return WALSegment{}, fmt.Errorf("invalid segment number in %q: %w", name, err)
	}

	return WALSegment{Timeline: uint32(timeline), Log: uint32(log), Segment: uint32(segment)}, nil
}

// String renders the segment back to its canonical file name.
func (w WALSegment) String() string {
	return fmt.Sprintf("%08X%08X%08X", w.Timeline, w.Log, w.Segment)
}

// Next returns the next segment in the same timeline. 255 is the maximum
// segment number before the log id rolls over.
func (w WALSegment) Next(segmentsPerLog uint32) WALSegment {
	if segmentsPerLog == 0 {
		segmentsPerLog = 1
	}
	next := w
	next.Segment++
	if next.Segment >= segmentsPerLog {
		next.Segment = 0
		next.Log++
	}
	return next
}

// segmentsPerLog returns the number of WAL segments in a logical log for the
// given wal_segment_size in bytes. 16MB segments give 4 GiB per log.
func segmentsPerLog(segmentSizeBytes int64) uint32 {
	if segmentSizeBytes <= 0 {
		return 0
	}
	const logBytes = int64(4 * 1024 * 1024 * 1024)
	return uint32(logBytes / segmentSizeBytes)
}

// FindGaps returns the missing segment names between sorted archived segments,
// used to detect an RPO violation. Adjacent segments may roll from the last
// segment of one log into segment 0 of the next; anything else is a gap.
func FindGaps(archived []WALSegment, segmentSizeBytes int64) []string {
	if len(archived) < 2 {
		return nil
	}

	perLog := segmentsPerLog(segmentSizeBytes)
	var gaps []string

	for i := 1; i < len(archived); i++ {
		if archived[i].Timeline != archived[i-1].Timeline {
			continue
		}

		expected := archived[i-1].Next(perLog)
		for expected.String() != archived[i].String() && !sameOrAfter(expected, archived[i]) {
			gaps = append(gaps, expected.String())
			expected = expected.Next(perLog)
			if len(gaps) > 10000 {
				return gaps
			}
		}
	}

	return gaps
}

// sameOrAfter reports whether a is the same segment as b or comes after it.
func sameOrAfter(a, b WALSegment) bool {
	if a.Timeline != b.Timeline {
		return a.Timeline > b.Timeline
	}
	if a.Log != b.Log {
		return a.Log > b.Log
	}
	return a.Segment >= b.Segment
}

// SortWALSegments sorts segments in WAL order (timeline, log, segment).
func SortWALSegments(segments []WALSegment) {
	sort.Slice(segments, func(i, j int) bool {
		return walLess(segments[i], segments[j])
	})
}

func walLess(a, b WALSegment) bool {
	if a.Timeline != b.Timeline {
		return a.Timeline < b.Timeline
	}
	if a.Log != b.Log {
		return a.Log < b.Log
	}
	return a.Segment < b.Segment
}

// walPrefix builds the S3 key prefix for a connection's WAL archive.
func walPrefix(connectionName string) string {
	return strings.TrimSuffix(connectionName, "/") + "/wal"
}
