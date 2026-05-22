package query_builder

import (
	"strings"
	"testing"

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
