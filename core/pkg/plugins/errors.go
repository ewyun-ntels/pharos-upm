package plugins

import "ntels.com/pharos/core/external"

// IsDatasourceUnavailable reports whether the error string indicates that
// the datasource is currently unavailable or unsupported. This allows callers
// to log-and-continue, deferring evaluation until provisioning is ready.
func IsDatasourceUnavailable(err string) bool {
	if err == "" {
		return false
	}
	return err == external.ErrorNotExistDatasource.Error() ||
		err == external.ErrorNotSupportedDatasourceType.Error() ||
		err == external.ErrorNotSupportedDataServer.Error() ||
		err == external.ErrorNotSupported.Error() ||
		err == external.ErrorNotImplemented.Error()
}
