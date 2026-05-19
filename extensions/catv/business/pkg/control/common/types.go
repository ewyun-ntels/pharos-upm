package common

const (
	CommandUse       = "control-schedule"
	HttpRelativePath = "/catv/control"

	ResponseCodeSuccess = "1"  // Command executed successfully
	ResponseCodeFailure = "0"  // Command failed (device-side error)
	ResponseCodeError   = "-1" // System error (network, validation, etc.)
)

type StbControlResponse struct {
	Code    string `json:"code"`    // Response code from the device
	Message string `json:"message"` // Response message or data from the device
}
