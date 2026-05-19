package external

import "errors"

var (
	ErrorNotSupported               = errors.New("not supported")
	ErrorNotSupportedDatasourceType = errors.New("not supported datasource type")
	ErrorNotSupportedDataServer     = errors.New("not supported data server")
	ErrorNotSupportedStreamServer   = errors.New("not supported stream server")
	ErrorNotSupportedPanelKind      = errors.New("not supported panel kind")

	ErrorNotConnected = errors.New("not connected")

	ErrorNotImplemented = errors.New("not implemented")

	ErrorNotExist                 = errors.New("not exist")
	ErrorNotExistFile             = errors.New("not exist file")
	ErrorNotExistDatasource       = errors.New("not exist datasource")
	ErrorNotExistPluginJSONSchema = errors.New("not exist plugin json schema")
	ErrorNotExistModelText        = errors.New("not exist model text")
	ErrorNotExistTableName        = errors.New("not exist table name")

	ErrorInvalidType                    = errors.New("invalid type")
	ErrorInvalidQuery                   = errors.New("invalid query")
	ErrorInvalidServeType               = errors.New("invalid serve type")
	ErrorInvalidServerType              = errors.New("invalid server type")
	ErrorInvalidServerSchema            = errors.New("invalid server schema")
	ErrorInvalidConfigType              = errors.New("invalid config type")
	ErrorInvalidPermissionType          = errors.New("invalid permission type")
	ErrorInvalidDatasourceType          = errors.New("invalid datasource type")
	ErrorInvalidReadSeekerType          = errors.New("invalid ReadSeeker type")
	ErrorInvalidDataServerRPCClientType = errors.New("invalid DataServerRPCClient type")
	ErrorInvalidChannel                 = errors.New("invalid channel")
	ErrorInvalidTaskName                = errors.New("invalid task name")
	ErrorInvalidDataFormat              = errors.New("invalid data format")
	ErrorInvalidDatasourceID            = errors.New("invalid datasource id")
	ErrorInvalidPanelID                 = errors.New("invalid panel id")

	ErrorNoSuchAgent      = errors.New("no such agent")
	ErrorNoSuchPolicy     = errors.New("no such policy")
	ErrorNoSuchPermission = errors.New("no such permission")

	ErrorEmptySubject = errors.New("empty subject")
	ErrorEmptyAction  = errors.New("empty action")
	ErrorEmptyKind    = errors.New("empty kind")

	ErrorMissingConfig     = errors.New("missing config")
	ErrorMissingPermission = errors.New("missing permission")

	ErrorMigrationCompleted = errors.New("migration complete - not error")
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func (e ErrorResponse) ToJson() string {
	return `{"message":"` + e.Message + `"}`
}

func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{Message: message}
}
