package udp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/panjf2000/gnet/v2"
	"ntels.com/pharos/core/external/clickhouse_interface"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	catv_common "ntels.com/pharos/extensions/catv/business/pkg/common"
)

const (
	TransmissionTypeUnknown                  = "unknown"
	TransmissionTypeDaily                    = "daily"
	TransmissionTypePeriodic                 = "periodic"
	TransmissionTypeDiagnostic               = "diagnostic"
	TransmissionTypeQualityMeasurement       = "quality_measurement"
	TransmissionTypeNetworkQualityTransition = "network_quality_transition"
)

type Table interface {
	// 기본 자료형이 아니면서 필수가 아닌 필드로인해 nil 처리를 위해 배치 강제
	convertBatchFormat() (map[string]any, error)

	getTableName() string
	setTimestamp(timestamp orm.Datetime)
}

func NewEventHandler(config common.Config, transmissionType string) *eventHandler {
	return &eventHandler{
		config:           config,
		transmissionType: transmissionType,
	}
}

type eventHandler struct {
	config           common.Config
	transmissionType string

	gnet.BuiltinEventEngine
}

func (h *eventHandler) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	return onTraffic(h.config, conn, h.transmissionType)
}

func toString(value any) any {
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return v
	}
}

func parseData(transmissionType string, data []byte) (Table, []error) {
	var parseErrors []error

	var dataMap = make(map[string]any)
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return nil, []error{err}
	}

	for key, value := range dataMap {
		switch key {
		case "loggingTime", "sendingTime":
			timeStr, ok := dataMap[key].(string)
			if !ok {
				parseErrors = append(parseErrors, fmt.Errorf("failed to assert time field %s as string", key))
				continue
			}

			var err error
			if dataMap[key+"Utc"], err = catv_common.ParseKSTDatetime(timeStr); err != nil {
				slog.Error("Failed to parse time field in data during parsing", "field", key, "value", timeStr, "error", err)
				parseErrors = append(parseErrors, fmt.Errorf("failed to parse time field %s: %w", key, err))
				continue
			}
		case "runningTime":
			timeStr, ok := dataMap[key].(string)
			if !ok {
				parseErrors = append(parseErrors, fmt.Errorf("failed to assert runningTime field as string"))
				continue
			}

			var err error
			if dataMap[key+"Sec"], err = catv_common.ParseRunningTimeSec(timeStr); err != nil {
				slog.Error("Failed to parse runningTime field in data during parsing", "field", key, "value", timeStr, "error", err)
				parseErrors = append(parseErrors, fmt.Errorf("failed to parse runningTime field %s: %w", key, err))
				continue
			}
		case "runningTimeSec":
			// runningTime 처리 중 추가된 uint64 값이 default의 toString()으로 문자열 변환되지 않도록 skip
		case "channels":
			var channelsMap []map[string]any

			if channelsBytes, err := json.Marshal(value); err != nil {
				slog.Error("Failed to marshal channels data for parsing", "error", err, "value", value)
				parseErrors = append(parseErrors, fmt.Errorf("failed to marshal channels data: %w", err))
				return nil, parseErrors
			} else if err := json.Unmarshal(channelsBytes, &channelsMap); err != nil {
				slog.Error("Failed to unmarshal channels data for parsing", "error", err, "value", string(channelsBytes))
				parseErrors = append(parseErrors, fmt.Errorf("failed to unmarshal channels data: %w", err))
				return nil, parseErrors
			}

			for key, channel := range channelsMap {
				for fieldKey, fieldValue := range channel {
					channelsMap[key][fieldKey] = toString(fieldValue)
				}
			}

			dataMap["channels"] = channelsMap
		default:
			dataMap[key] = toString(value)
		}
	}

	var table Table
	switch transmissionType {
	case TransmissionTypeDaily:
		table = &Daily{}
	case TransmissionTypePeriodic:
		table = &Periodic{}
	case TransmissionTypeDiagnostic:
		table = &Diagnostic{}
	case TransmissionTypeQualityMeasurement:
		table = &QualityMeasurement{}
	case TransmissionTypeNetworkQualityTransition:
		table = &NetworkQualityTransition{}
	default:
		slog.Error("Unknown transmission type for parsing", "type", transmissionType)
		parseErrors = append(parseErrors, fmt.Errorf("unknown transmission type: %s", transmissionType))
		return nil, parseErrors
	}

	if bytesData, err := json.Marshal(dataMap); err != nil {
		slog.Error("Failed to marshal data after converting specified keys to string", "error", err, "data", dataMap)
		parseErrors = append(parseErrors, fmt.Errorf("failed to marshal data after converting specified keys to string: %w", err))
		return nil, parseErrors
	} else if err := json.Unmarshal(bytesData, table); err != nil {
		slog.Error("Failed to unmarshal data into destination struct after converting specified keys to string", "error", err, "data", string(bytesData), "destination_struct", fmt.Sprintf("%T", table))
		parseErrors = append(parseErrors, fmt.Errorf("failed to unmarshal data into destination struct after converting specified keys to string: %w", err))
		return nil, parseErrors
	}

	return table, parseErrors
}

func insert[T Table](table T, config common.Config) error {
	var batch = clickhouse_interface.ClickhouseBatch{}
	if data, err := table.convertBatchFormat(); err != nil {
		slog.Error("Failed to convert table data to batch format", "error", err, "table", table)
		return err
	} else if err := batch.AddBatchJson(data); err != nil {
		slog.Error("Failed to add table data to ClickHouse batch", "error", err, "table", table)
		return err
	} else if err := batch.Insert(config.Catv.Database, table.getTableName()); err != nil {
		slog.Error("Failed to insert table data into ClickHouse", "error", err, "table", table.getTableName())
		return err
	}

	return nil
}

func onTraffic(config common.Config, conn gnet.Conn, transmissionType string) (action gnet.Action) {
	var data []byte
	if buffer, err := conn.Next(-1); err != nil {
		slog.Error("Failed to read data from connection", "error", err, "type", transmissionType)
		return gnet.None
	} else {
		data = bytes.Clone(buffer)
	}

	var timestamp = orm.Datetime{Time: time.Now().UTC()}

	table, errs := parseData(transmissionType, data)
	if table != nil {
		table.setTimestamp(timestamp)
		if err := insert(table, config); err != nil {
			slog.Error("Failed to insert data into ClickHouse", "error", err, "type", transmissionType)
			errs = append(errs, err)
		}
	}

	var errsStr = make([]string, 0, len(errs))
	for _, err := range errs {
		errsStr = append(errsStr, err.Error())
	}

	var raw = Raw{Timestamp: timestamp, TransmissionType: transmissionType, RawData: string(data), Errors: errsStr}
	if err := raw.insert(config); err != nil {
		slog.Error("Failed to insert raw table data", "error", err, "type", transmissionType)
	}

	return gnet.None
}
