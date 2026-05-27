package query_builder

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild_DefaultUsesPongo2(t *testing.T) {
	result, err := Build("", "SELECT {{ field }} FROM {{ table }}", map[string]any{
		"field": "cpu",
		"table": "metrics",
	})
	require.NoError(t, err)

	assert.Equal(t, "SELECT cpu FROM metrics", strings.TrimSpace(string(result)))
}

func TestBuild_GrafanaVariables(t *testing.T) {
	result, err := Build(StyleGrafana,
		`up{job=~"$job", pod=~"[[pod:regex]]", missing="$missing"} + $1`,
		map[string]any{
			"job": "production",
			"pod": "tcp-bridge-1|tcp-bridge-2",
		},
	)
	require.NoError(t, err)

	assert.Equal(t,
		`up{job=~"production", pod=~"(tcp-bridge-1|tcp-bridge-2)", missing="$missing"} + $1`,
		string(result),
	)
}

func TestBuild_GrafanaKeepsUnsupportedAndFieldPath(t *testing.T) {
	result, err := Build(StyleGrafana,
		`${job:unknown} ${job.field} ${job.field.path} $job`,
		map[string]any{"job": "production"},
	)
	require.NoError(t, err)

	assert.Equal(t, `${job:unknown} ${job.field} ${job.field.path} production`, string(result))
}

func TestBuild_GrafanaFormats(t *testing.T) {
	ctx := map[string]any{
		"hosts": []any{"api.1", "api-2"},
		"name":  "O'Reilly",
	}

	cases := []struct {
		name  string
		query string
		want  string
	}{
		{"glob", "${hosts}", "{api.1,api-2}"},
		{"csv", "${hosts:csv}", "api.1,api-2"},
		{"raw", "${hosts:raw}", "api.1,api-2"},
		{"pipe", "${hosts:pipe}", "api.1|api-2"},
		{"regex", "${hosts:regex}", `(api\.1|api-2)`},
		{"singlequote", "${name:singlequote}", `'O\'Reilly'`},
		{"sqlstring", "${name:sqlstring}", `'O''Reilly'`},
		{"json", "${hosts:json}", `["api.1","api-2"]`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Build(StyleGrafana, tc.query, ctx)
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(result))
		})
	}
}

func TestBuild_GrafanaTimeAliasesAndInterval(t *testing.T) {
	result, err := Build(StyleGrafana,
		`$__from $__to ${__from:date:seconds} ${__to:date:seconds} $__interval $__interval_ms`,
		map[string]any{
			"__start_time":  float64(1700000000),
			"__end_time":    float64(1700003600),
			"__interval":    "30s",
			"__interval_ms": 30000,
		},
	)
	require.NoError(t, err)

	assert.Equal(t, `1700000000000 1700003600000 1700000000 1700003600 30s 30000`, string(result))
}

func TestBuild_GrafanaKeepsUnsupportedTimeFormats(t *testing.T) {
	result, err := Build(StyleGrafana,
		`${__from:date} ${__from:date:YYYY-MM}`,
		map[string]any{"__start_time": float64(1700000000)},
	)
	require.NoError(t, err)

	assert.Equal(t, `${__from:date} ${__from:date:YYYY-MM}`, string(result))
}

func TestBuild_GrafanaRateIntervalFallback(t *testing.T) {
	result, err := Build(StyleGrafana,
		`increase(metric[$__rate_interval]) $__rate_interval_ms`,
		map[string]any{
			"__interval": "30s",
		},
	)
	require.NoError(t, err)

	assert.Equal(t, `increase(metric[2m]) 120000`, string(result))
}

func TestBuild_GrafanaRateIntervalWithScrapeInterval(t *testing.T) {
	result, err := Build(StyleGrafana,
		`increase(metric[$__rate_interval]) $__rate_interval_ms`,
		map[string]any{
			"__interval":        "30s",
			"__scrape_interval": "30s",
		},
	)
	require.NoError(t, err)

	assert.Equal(t, `increase(metric[2m]) 120000`, string(result))
}

func TestBuild_GrafanaRateIntervalWithoutInterval(t *testing.T) {
	result, err := Build(StyleGrafana,
		`increase(metric[$__rate_interval]) $__rate_interval_ms`,
		map[string]any{},
	)
	require.NoError(t, err)

	assert.Equal(t, `increase(metric[2m]) 120000`, string(result))
}

func TestBuild_GrafanaRateIntervalKeepsExplicitValues(t *testing.T) {
	result, err := Build(StyleGrafana,
		`increase(metric[$__rate_interval]) $__rate_interval_ms`,
		map[string]any{
			"__interval":         "30s",
			"__rate_interval":    "2m",
			"__rate_interval_ms": 120000,
		},
	)
	require.NoError(t, err)

	assert.Equal(t, `increase(metric[2m]) 120000`, string(result))
}

func TestBuild_GrafanaRateIntervalBackfillsMilliseconds(t *testing.T) {
	result, err := Build(StyleGrafana,
		`increase(metric[$__rate_interval]) $__rate_interval_ms`,
		map[string]any{
			"__rate_interval": "2m",
		},
	)
	require.NoError(t, err)

	assert.Equal(t, `increase(metric[2m]) 120000`, string(result))
}

func TestBuild_GrafanaRateIntervalBackfillsDuration(t *testing.T) {
	result, err := Build(StyleGrafana,
		`increase(metric[$__rate_interval]) $__rate_interval_ms`,
		map[string]any{
			"__rate_interval_ms": 90000,
		},
	)
	require.NoError(t, err)

	assert.Equal(t, `increase(metric[1m30s]) 90000`, string(result))
}

func TestParsePromDuration(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  time.Duration
	}{
		{"numeric seconds", 30, 30 * time.Second},
		{"numeric string seconds", "30", 30 * time.Second},
		{"seconds", "30s", 30 * time.Second},
		{"minutes and seconds", "1m30s", 90 * time.Second},
		{"hours", "1h", time.Hour},
		{"days", "1d", 24 * time.Hour},
		{"weeks", "1w", 7 * 24 * time.Hour},
		{"mixed days", "1d15s", 24*time.Hour + 15*time.Second},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parsePromDuration(tc.value)
			require.True(t, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParsePromDurationInvalid(t *testing.T) {
	cases := []struct {
		name  string
		value any
	}{
		{"empty", ""},
		{"text", "abc"},
		{"unknown unit", "1x"},
		{"negative", "-1s"},
		{"nil", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := parsePromDuration(tc.value)
			assert.False(t, ok)
		})
	}
}

func TestFormatPromDuration(t *testing.T) {
	cases := []struct {
		name  string
		value time.Duration
		want  string
	}{
		{"subsecond clamps to second", 500 * time.Millisecond, "1s"},
		{"seconds", 45 * time.Second, "45s"},
		{"minute", time.Minute, "1m"},
		{"minute seconds", 90 * time.Second, "1m30s"},
		{"hour", time.Hour, "1h"},
		{"hour minute", time.Hour + time.Minute, "1h1m"},
		{"day", 24 * time.Hour, "1d"},
		{"day seconds", 24*time.Hour + 15*time.Second, "1d15s"},
		{"week", 7 * 24 * time.Hour, "1w"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, formatPromDuration(tc.value))
		})
	}
}
