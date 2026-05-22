package query_builder

import "log/slog"

type Style string

const (
	StylePongo2  Style = "pongo2"
	StyleGrafana Style = "grafana"
)

func Build(style Style, query string, queryContext map[string]any) ([]byte, error) {
	if queryContext == nil {
		queryContext = map[string]any{}
	}

	slog.Debug("Building query",
		"templateStyle", string(style),
		"rawQuery", query,
		"variables", queryContext,
		"variableCount", len(queryContext),
	)

	var (
		result []byte
		err    error
	)

	switch style {
	case StyleGrafana:
		result, err = buildGrafana(query, queryContext)
	default:
		result, err = GetQuery(query, queryContext)
	}

	if err != nil {
		slog.Debug("Build query failed",
			"templateStyle", string(style),
			"rawQuery", query,
			"error", err,
		)
		return nil, err
	}

	slog.Debug("Built query",
		"templateStyle", string(style),
		"rawQuery", query,
		"executedQuery", string(result),
		"variables", queryContext,
	)

	return result, nil
}
