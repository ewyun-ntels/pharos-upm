package tables

import (
	"io"
	"strings"
	"time"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

const CRLF = "\r\n"

type WeatherTable interface {
	TableName() string
	Insert(config common.Config) error
	Transform(insertTime orm.Datetime, data string) ([]WeatherTable, error)
	ConvertBatchFormat() (map[string]any, error)

	GetMockData(t time.Time) (string, error)
}

func UTF8ToEUC_KR(data string) ([]byte, error) {
	return io.ReadAll(transform.NewReader(strings.NewReader(data), korean.EUCKR.NewEncoder()))
}

func EUC_KRToUTF8(data string) ([]byte, error) {
	return io.ReadAll(transform.NewReader(strings.NewReader(data), korean.EUCKR.NewDecoder()))
}
