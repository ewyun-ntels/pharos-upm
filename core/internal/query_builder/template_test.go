package query_builder

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckTimestamp(t *testing.T) {
	examples := []string{
		" 5 Seconds ago",
		" 6 Seconds ago ",
		" 6  Seconds ago ",
		" 36000  Seconds ago ",
		"1 Months from now",
		"5 Days ago",
		"10 Hours from now",
		"3 Years ago",
		"7 Weeks from now",
		"2 Min",
	}

	for _, example := range examples {
		number, unit, err := parseExpression(example)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			fmt.Printf("Parsed: %d, %s\n", number, unit)
		}
	}
}

func TestParseAndSetTime(t *testing.T) {
	context := map[string]any{
		"start_time": "5 Seconds ago",
		"end_time":   "10 Minutes from now",
	}

	parseAndSetTime(context, "start_time")
	parseAndSetTime(context, "end_time")

	for key, value := range context {
		fmt.Printf("%s: %v\n", key, value)
	}
}

func TestParseAndSetDuration(t *testing.T) {
	checkTimes := []string{
		"5 Seconds ago",
		"10 Minutes from now",
		"2 Hours ago",
		"1 Days ago",
		"3 Weeks from now",
		"1 Months ago",
	}
	for _, checkTime := range checkTimes {
		timestamp, err := calculateUnixTimestamp(checkTime)
		assert.NoError(t, err)

		fmt.Printf("Timestamp for '%s': %d\n", checkTime, timestamp)
	}
}

// ─── parseAndSetTime: 숫자 타입 처리 ──────────────────────────────────────────

func TestParseAndSetTime_NumericTypes(t *testing.T) {
	// JSON number는 Go에서 float64로 파싱되므로, 백엔드가 int64로 변환해야 함
	// 변환 안 되면 pongo2가 "1.772970183e+09" 같은 과학적 표기법으로 렌더링
	const fixedUnix = int64(1772970183)

	cases := []struct {
		name  string
		input any
	}{
		{"int", int(fixedUnix)},
		{"int32", int32(fixedUnix)},
		{"int64", int64(fixedUnix)},
		// float32는 32비트 정밀도 한계로 큰 Unix timestamp를 근사 저장하므로 제외
		// 프론트엔드(JavaScript)는 float64만 사용하여 실제 발생하지 않는 케이스
		{"float64", float64(fixedUnix)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := map[string]any{"__start_time": tc.input}
			parseAndSetTime(ctx, "__start_time")

			got, ok := ctx["__start_time"].(int64)
			require.True(t, ok, "expected int64, got %T (%v)", ctx["__start_time"], ctx["__start_time"])
			assert.Equal(t, fixedUnix, got)
		})
	}
}

func TestParseAndSetTime_StringType(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  func() int64
	}{
		{
			name:  "Now",
			input: "Now",
			want:  func() int64 { return time.Now().Unix() },
		},
		{
			name:  "1 Hours ago",
			input: "1 Hours ago",
			want:  func() int64 { return time.Now().Add(-1 * time.Hour).Unix() },
		},
		{
			name:  "30 Minutes ago",
			input: "30 Minutes ago",
			want:  func() int64 { return time.Now().Add(-30 * time.Minute).Unix() },
		},
		{
			name:  "7 Days ago",
			input: "7 Days ago",
			want:  func() int64 { return time.Now().AddDate(0, 0, -7).Unix() },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := map[string]any{"__start_time": tc.input}
			parseAndSetTime(ctx, "__start_time")

			got, ok := ctx["__start_time"].(int64)
			require.True(t, ok, "expected int64, got %T", ctx["__start_time"])

			delta := math.Abs(float64(tc.want() - got))
			assert.LessOrEqual(t, delta, float64(5), "delta too large: %.0fs", delta)
		})
	}
}

func TestParseAndSetTime_InvalidType_ValueUnchanged(t *testing.T) {
	// 처리 불가 타입은 값이 변경되지 않아야 함 (bool 등)
	ctx := map[string]any{"__start_time": true}
	parseAndSetTime(ctx, "__start_time")
	assert.Equal(t, true, ctx["__start_time"])
}

func TestParseAndSetTime_KeyNotExist(t *testing.T) {
	ctx := map[string]any{}
	parseAndSetTime(ctx, "__start_time") // should not panic or create key
	_, exists := ctx["__start_time"]
	assert.False(t, exists, "key should not be created when absent")
}

// ─── parseAndSetTimeMs ────────────────────────────────────────────────────────

func TestParseAndSetTimeMs_StringType(t *testing.T) {
	ctx := map[string]any{"__start_time_ms": "1 Hours ago"}
	parseAndSetTimeMs(ctx, "__start_time_ms")

	got, ok := ctx["__start_time_ms"].(int64)
	require.True(t, ok, "expected int64, got %T", ctx["__start_time_ms"])

	want := time.Now().Add(-1*time.Hour).Unix() * 1000
	delta := math.Abs(float64(want - got))
	assert.LessOrEqual(t, delta, float64(5000), "delta too large: %.0fms", delta)
}

func TestParseAndSetTimeMs_NumericKeptAsIs(t *testing.T) {
	// 숫자는 이미 밀리초로 가정 → 그대로 유지
	cases := []struct {
		name  string
		input any
	}{
		{"int64", int64(1772970183000)},
		{"float64", float64(1772970183000)},
		{"int", int(1772970183000)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := map[string]any{"__start_time_ms": tc.input}
			parseAndSetTimeMs(ctx, "__start_time_ms")
			assert.Equal(t, tc.input, ctx["__start_time_ms"], "numeric value should remain unchanged")
		})
	}
}

// ─── GetQuery ─────────────────────────────────────────────────────────────────

func TestGetQuery_Float64NotScientificNotation(t *testing.T) {
	// 핵심 회귀 테스트
	// 수정 전: float64(1772970183) → pongo2가 "1.772970183e+09" 렌더링 → SQL 오류
	// 수정 후: float64(1772970183) → int64(1772970183) → pongo2가 "1772970183" 렌더링
	result, err := GetQuery("{{ __start_time }}", map[string]any{
		"__start_time": float64(1772970183),
	})
	require.NoError(t, err)

	got := strings.TrimSpace(string(result))
	assert.False(t, strings.ContainsAny(got, "eE"), "must not render as scientific notation, got: %q", got)
	assert.Equal(t, "1772970183", got)
}

func TestGetQuery_StringTimeReplaced(t *testing.T) {
	result, err := GetQuery(
		"SELECT * FROM t WHERE time > {{ __start_time }} AND time < {{ __end_time }}",
		map[string]any{
			"__start_time": "1 Hours ago",
			"__end_time":   "Now",
		},
	)
	require.NoError(t, err)

	got := string(result)
	// 문자열 표현식이 숫자로 치환됐는지 확인
	assert.False(t, strings.Contains(got, "Hours ago"), "time expression should be replaced")
	assert.False(t, strings.Contains(got, "Now"), "time expression should be replaced")
}

func TestGetQuery_MultipleTimeParams(t *testing.T) {
	// __start_time, __end_time 동시에 float64로 들어오는 케이스
	result, err := GetQuery(
		"{{ __start_time }} {{ __end_time }}",
		map[string]any{
			"__start_time": float64(1772970183),
			"__end_time":   float64(1772973783),
		},
	)
	require.NoError(t, err)

	got := strings.TrimSpace(string(result))
	// 과학적 표기법 없어야 함 (SQL 키워드 없는 순수 숫자 출력으로 검증)
	assert.False(t, strings.ContainsAny(got, "eE"), "must not render as scientific notation, got: %q", got)
	assert.Contains(t, got, "1772970183")
	assert.Contains(t, got, "1772973783")
}

func TestGetQuery_PlainVariables(t *testing.T) {
	result, err := GetQuery(
		"SELECT {{ field }} FROM {{ table }} LIMIT {{ limit }}",
		map[string]any{
			"field": "cpu",
			"table": "metrics",
			"limit": 100,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "SELECT cpu FROM metrics LIMIT 100", strings.TrimSpace(string(result)))
}

func TestGetQuery_InvalidTemplate(t *testing.T) {
	_, err := GetQuery("{{ unclosed", map[string]any{})
	assert.Error(t, err)
}
