package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func HttpBodyParse(contentType string, body []byte) (map[string]interface{}, error) {
	ctx := make(map[string]interface{})

	switch contentType {
	case "text/plain", "text/html":
		ctx["message"] = string(body)
		return ctx, nil
	case "application/json":
		err := json.Unmarshal(body, &ctx)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("unmarshal json error:\n%s", string(body)))
		}
		return ctx, nil
	default:
		return nil, errors.Errorf("Unknown Content-Type: %s", contentType)
	}
}

func ContextPath(ctx interface{}, path string) (value interface{}, isTable bool) {
	if path == "" {
		return ctx, false
	}
	pathSegments := strings.Split(path, ".")

	seg := pathSegments[0]
	next := strings.Join(pathSegments[1:], ".")
	switch v := ctx.(type) {
	case map[string]interface{}:
		return ContextPath(v[seg], next)
	case []interface{}:
		if seg == "@" {
			table := make([]interface{}, len(v))
			for idx, item := range v {
				itemValue, _ := ContextPath(item, next)
				table[idx] = itemValue
			}
			return table, true
		} else if seg == "#" {
			return len(v), false
		} else {
			n, na := strconv.Atoi(seg)
			if na != nil || n >= len(v) {
				return nil, false
			}
			return ContextPath(v[n], next)
		}
	}

	return nil, false
}
