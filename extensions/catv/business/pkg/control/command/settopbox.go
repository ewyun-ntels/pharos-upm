package command

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"slices"
	"strings"
	"time"

	"ntels.com/pharos/core/pkg/common"
	catv_common "ntels.com/pharos/extensions/catv/business/pkg/common"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

type Settopbox struct {
	config            common.Config
	scheduleResultRaw tables.StbControlScheduleResultRaw

	ctx          context.Context
	cipherSuites []uint16

	connectionTimeout        time.Duration
	sendTimeout              time.Duration
	portExhaustionRetryCount int
	portExhaustionRetryDelay time.Duration
	totalRetryCount          int
	totalRetryDelay          time.Duration
}

func NewSettopbox(config common.Config, ctx context.Context, stbControlScheduleResultRaw tables.StbControlScheduleResultRaw) *Settopbox {
	connectionTimeout, err := time.ParseDuration(config.Catv.Control.Timeout.Connection)
	if err != nil {
		slog.Error("Invalid connection timeout, using default 3s", "input", config.Catv.Control.Timeout.Connection, "error", err)
		connectionTimeout = 3 * time.Second
	}
	sendTimeout, err := time.ParseDuration(config.Catv.Control.Timeout.Send)
	if err != nil {
		slog.Error("Invalid send timeout, using default 5s", "input", config.Catv.Control.Timeout.Send, "error", err)
		sendTimeout = 5 * time.Second
	}
	portExhaustionRetryDelay, err := time.ParseDuration(config.Catv.Control.Retry.PortExhaustion.Delay)
	if err != nil {
		slog.Error("Invalid port exhaustion retry delay, using default 20s", "input", config.Catv.Control.Retry.PortExhaustion.Delay, "error", err)
		portExhaustionRetryDelay = 20 * time.Second
	}
	totalRetryDelay, err := time.ParseDuration(config.Catv.Control.Retry.Total.Delay)
	if err != nil {
		slog.Error("Invalid total retry delay, using default 5s", "input", config.Catv.Control.Retry.Total.Delay, "error", err)
		totalRetryDelay = 5 * time.Second
	}

	// Direct TCP dial with context: pool not used (unique IP per device = 0% reuse)
	// STB 장비는 TLS 서버이나 구형 cipher suite(RSA 키 교환 계열)만 지원함
	// Go 기본 cipher는 ECDHE 계열만 포함 → handshake failure
	// InsecureCipherSuites 포함 전체 cipher 목록을 활성화하여 협상 가능하게 함
	allCiphers := func() []uint16 {
		ids := make([]uint16, 0)
		for _, s := range tls.CipherSuites() {
			ids = append(ids, s.ID)
		}
		for _, s := range tls.InsecureCipherSuites() {
			ids = append(ids, s.ID)
		}
		return ids
	}()

	return &Settopbox{
		config:            config,
		scheduleResultRaw: stbControlScheduleResultRaw,

		ctx:          ctx,
		cipherSuites: allCiphers,

		connectionTimeout:        connectionTimeout,
		sendTimeout:              sendTimeout,
		portExhaustionRetryCount: config.Catv.Control.Retry.PortExhaustion.Count,
		portExhaustionRetryDelay: portExhaustionRetryDelay,
		totalRetryCount:          config.Catv.Control.Retry.Total.Count,
		totalRetryDelay:          totalRetryDelay,
	}
}

func (c *Settopbox) controlDetail() StbControlResponse {
	var response []byte
	var err error

	if c.config.Catv.Control.Simulation {
		response, err = c.simulation()
	} else {
		response, err = c.notSimulation()
	}
	if err != nil {
		slog.Error("Settopbox control failed", "stb_mdl_nm", c.scheduleResultRaw.StbMdlNm, "cm_mac_addr", c.scheduleResultRaw.CmMacAddr, "stb_mac_addr", c.scheduleResultRaw.StbMacAddr, "error", err)
		return StbControlResponse{
			Code:    control_common.ResponseCodeError,
			Message: err.Error(),
		}
	} else {
		return c.parseResponse(c.scheduleResultRaw.WorkType, response)
	}
}

func (c *Settopbox) Control(retry bool) StbControlResponse {
	if !slices.Contains(AllWorkTypes, c.scheduleResultRaw.WorkType) {
		return StbControlResponse{
			Code:    control_common.ResponseCodeError,
			Message: fmt.Sprintf("invalid settopbox request: unsupported work_type '%s'", c.scheduleResultRaw.WorkType),
		}
	}

	var lastRespone StbControlResponse

	for attempt := range c.totalRetryCount {
		select {
		case <-c.ctx.Done():
			return StbControlResponse{
				Code:    control_common.ResponseCodeError,
				Message: "context cancelled",
			}
		default:
		}

		if attempt > 0 {
			backoff := c.totalRetryDelay * time.Duration(attempt)
			slog.Debug("Retrying stb control command",
				"attempt", attempt+1,
				"max_attempts", c.totalRetryCount,
				"stb_mdl_nm", c.scheduleResultRaw.StbMdlNm,
				"cm_mac_addr", c.scheduleResultRaw.CmMacAddr,
				"stb_mac_addr", c.scheduleResultRaw.StbMacAddr,
				"backoff_seconds", backoff.Seconds())

			select {
			case <-c.ctx.Done():
				return StbControlResponse{
					Code:    control_common.ResponseCodeError,
					Message: "context cancelled during retry",
				}
			case <-time.After(backoff):
			}
		}

		lastRespone = c.controlDetail()
		if retry && c.isRetryableError(lastRespone.Message) {
			slog.Warn("Retryable error occurred during stb control command, will retry",
				"attempt", attempt+1,
				"max_attempts", c.totalRetryCount,
				"stb_mdl_nm", c.scheduleResultRaw.StbMdlNm,
				"cm_mac_addr", c.scheduleResultRaw.CmMacAddr,
				"stb_mac_addr", c.scheduleResultRaw.StbMacAddr,
				"error_message", lastRespone.Message)
			continue
		}

		return lastRespone
	}

	return StbControlResponse{
		Code:    control_common.ResponseCodeError,
		Message: fmt.Sprintf("failed after %d attempts: %v", c.totalRetryCount, lastRespone.Message),
	}
}

func (c *Settopbox) simulation() ([]byte, error) {
	codes := []string{control_common.ResponseCodeSuccess, control_common.ResponseCodeFailure, control_common.ResponseCodeError}
	code := codes[rand.IntN(len(codes))]
	failureMessages := []string{"fail", "wrong parameter", "wrong work type", "no data", "no action", "not supported"}
	errorMessages := []string{
		"failed to connect to " + c.getConnectionAddress() + ": context deadline exceeded",
		"no response received from settopbox",
		"failed after 1 attempts: communication error with 3C:62:00:81:0D:10: read tcp 10.42.0.145:59944->" + c.getConnectionAddress() + ": read: connection reset by peer",
		"communication error with 54:FA:3E:C2:0A:21: remote error: tls: illegal parameter",
	}

	var message string
	switch code {
	case control_common.ResponseCodeSuccess:
		message = "success"
	case control_common.ResponseCodeFailure:
		message = failureMessages[rand.IntN(len(failureMessages))]
	case control_common.ResponseCodeError:
		return nil, errors.New(errorMessages[rand.IntN(len(errorMessages))])
	}

	var response any
	switch c.scheduleResultRaw.WorkType {
	case WorkTypeStbRequestInfo:
		now := time.Now()
		timestamp := now.Format("2006/01/02 15:04")
		hostId := c.scheduleResultRaw.StbMacAddr
		if len(hostId) > 10 {
			hostId = hostId[:10]
		}

		response = map[string]string{
			"hostId":        pickRandom([]string{hostId, "1A8020DF41", "2B9031EF52"}),
			"macAddr":       c.scheduleResultRaw.StbMacAddr,
			"cmMac":         c.scheduleResultRaw.CmMacAddr,
			"stbIp":         pickRandom([]string{"10.43.81.14", "10.43.82.25", "10.43.83.36"}),
			"cmIp":          c.scheduleResultRaw.CmIpAddr,
			"stbModel":      pickRandom([]string{"THX-U300", "THX-U400", "HC100", "UC2000", "UC2600", "SX730C"}),
			"mwVer":         pickRandom([]string{"3.1.46", "3.1.47", "3.2.01", "4.0.12"}),
			"localVer":      pickRandom([]string{"1.0.6.03", "1.0.7.02", "1.1.0.01"}),
			"cloudVer":      pickRandom([]string{"1.6.12", "1.6.13", "1.7.01"}),
			"loggingTime":   timestamp,
			"sendingTime":   timestamp,
			"limitAge":      pickRandom([]string{"0", "1", "2", "3", "4"}),
			"tvLock":        pickRandom([]string{"Off", "0", "1", "2"}),
			"skipCh":        pickRandom([]string{"", "11,10,15", "3,7,11"}),
			"easyBuying":    pickRandom([]string{"On", "Off", "NoOpt"}),
			"favCh":         pickRandom([]string{"", "11,10,15", "1,2,3,4,5"}),
			"zappingAd":     pickRandom([]string{"0", "1"}),
			"miniEpg":       pickRandom([]string{"0", "3", "5", "10"}),
			"miniEpgAd":     pickRandom([]string{"On", "Off"}),
			"tvCaption":     pickRandom([]string{"0", "1", "2", "3", "4", "5", "6"}),
			"tvImpaired":    pickRandom([]string{"On", "Off"}),
			"barkerCh":      pickRandom([]string{"N", "Y"}),
			"vodView":       pickRandom([]string{"Poster", "Text"}),
			"vodRelay":      pickRandom([]string{"On", "Off"}),
			"resolution":    pickRandom([]string{"0", "1", "2", "3", "4", "5"}),
			"audioMode":     pickRandom([]string{"2", "3"}),
			"hdmiCec":       pickRandom([]string{"true", "false"}),
			"hdcp":          pickRandom([]string{"on", "off"}),
			"hdr":           pickRandom([]string{"0", "1"}),
			"mobilePay":     pickRandom([]string{"Y", "N"}),
			"morningAlarm":  pickRandom([]string{"repeatSetting:0", "repeatSetting:1,channelSetting:111,channelNum:11,channelName:MBC,timeSetting:0700"}),
			"bootMenu":      pickRandom([]string{"On", "Off"}),
			"pmsOn":         pickRandom([]string{"On", "Off"}),
			"oneAdOn":       pickRandom([]string{"true", "false"}),
			"audioLang":     pickRandom([]string{"kor", "eng", "jpn", "chi", "fre", "ger", "spa", "ara", "por", "ita", "rus"}),
			"standbyMode":   pickRandom([]string{"On", "Off"}),
			"savePwr":       pickRandom([]string{"0", "300000", "10800000"}),
			"voiceGuide":    pickRandom([]string{"On|1", "On|2", "On|3", "Off|1"}),
			"chNum":         pickRandom([]string{"3", "11", "15", "VOD", "DATA"}),
			"chSid":         pickRandom([]string{"185", "251", "312", "401"}),
			"chName":        pickRandom([]string{"EBS HD", "MBC HD", "SBS HD", "VOD", "Youtube"}),
			"chPrg":         pickRandom([]string{"일단 해봐요 생방송 오후 1시", "뉴스데스크", "주말 드라마"}),
			"chFreq":        pickRandom([]string{"741MHz", "747MHz", "753MHz", "240MHz"}),
			"chMode":        pickRandom([]string{"8VSB", "256QAM"}),
			"pwrLvl":        pickRandom([]string{"12dBmV", "5dBmV", "0dBmV", "-3dBmV", "-12dBmV"}),
			"snr":           pickRandom([]string{"33dB", "35dB", "38dB", "40dB", "42dB"}),
			"sigWeak":       pickRandom([]string{"Y", "N"}),
			"sigWeakCnt":    pickRandom([]string{"0", "1", "2", "5"}),
			"volume":        pickRandom([]string{"00", "05", "10", "15", "20"}),
			"homeState":     pickRandom([]string{"show", "hide"}),
			"stbState":      pickRandom([]string{"watching", "standby"}),
			"runningTime":   pickRandom([]string{"1day 1hour 1min 1sec", "2day 5hour 32min 45sec", "0day 12hour 5min 23sec"}),
			"limitContents": pickRandom([]string{"Protect", "Hide", "Show"}),
		}
	default:
		response = StbControlResponse{Code: code, Message: message}
	}

	if responseBytes, err := json.Marshal(response); err != nil {
		return nil, fmt.Errorf("failed to marshal simulation response: %v", err)
	} else {
		return responseBytes, nil
	}
}

func (c *Settopbox) notSimulation() ([]byte, error) {
	commandBytes, err := c.makeCommandBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settopbox request: %v", err)
	}

	conn, err := c.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get settopbox connection: %v", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(c.sendTimeout)); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %v", err)
	}

	if _, err := conn.Write(commandBytes); err != nil {
		return nil, fmt.Errorf("failed to send command: %v", err)
	}

	// STB responds only after the write side is closed (half-close).
	// CloseWrite sends TLS close_notify on the write direction while keeping
	// the read side open so we can receive the response.
	if tlsConn, ok := conn.(*tls.Conn); ok {
		if err := tlsConn.CloseWrite(); err != nil {
			return nil, fmt.Errorf("failed to half-close connection: %v", err)
		}
	}

	bufferPtr := responseBufferPool.Get().(*[]byte)
	buffer := *bufferPtr
	defer responseBufferPool.Put(bufferPtr)

	response := make([]byte, 0, 4096)
	for {
		if n, err := conn.Read(buffer); n > 0 {
			response = append(response, buffer[:n]...)
		} else if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() && len(response) > 0 {
				break
			} else if err.Error() != "EOF" {
				return nil, fmt.Errorf("communication error with %s: %v", c.scheduleResultRaw.CmMacAddr, err)
			}
			break
		}
	}

	return response, nil
}

func (c *Settopbox) makeCommandBytes() ([]byte, error) {
	command := map[string]any{
		"work_type": c.scheduleResultRaw.WorkType,
	}
	if c.scheduleResultRaw.WorkValue != nil {
		command["work_value"] = c.scheduleResultRaw.WorkValue
	}
	return json.Marshal(command)
}

func (c *Settopbox) getConnection() (net.Conn, error) {
	dialer := tls.Dialer{
		NetDialer: &net.Dialer{Timeout: c.connectionTimeout},
		Config: &tls.Config{ //nolint:gosec
			InsecureSkipVerify: true,             // STB 자체 서명 인증서
			MinVersion:         tls.VersionTLS10, // STB는 TLS 1.0 지원
			CipherSuites:       c.cipherSuites,   // 구형 RSA cipher 포함 전체 활성화
		},
	}

	var conn net.Conn
	var lastErr error

	connectionAddress := c.getConnectionAddress()
	for attempt := range c.portExhaustionRetryCount {
		conn, lastErr = dialer.DialContext(c.ctx, "tcp", connectionAddress)
		if lastErr == nil {
			break
		}

		if c.ctx.Err() != nil {
			return nil, c.ctx.Err()
		}

		if c.isPortExhaustionError(lastErr) && attempt < c.portExhaustionRetryCount-1 {
			backoff := time.Duration(attempt+1) * c.portExhaustionRetryDelay
			slog.Warn("Port exhaustion detected, backing off", "address", connectionAddress, "attempt", attempt+1, "backoff_seconds", backoff.Seconds(), "error", lastErr)

			timer := time.NewTimer(backoff)
			select {
			case <-c.ctx.Done():
				timer.Stop()
				return nil, errors.New("context cancelled during port exhaustion backoff")
			case <-timer.C:
				continue
			}
		}

		break
	}

	if conn == nil {
		slog.Error("Failed to establish connection", "address", connectionAddress, "error", lastErr)
		return nil, fmt.Errorf("failed to connect to %s: %v", connectionAddress, lastErr)
	}

	return conn, nil
}

func (c *Settopbox) getConnectionAddress() string {
	return fmt.Sprintf("%s:%d", c.scheduleResultRaw.SrcIpAddr, c.config.Catv.Control.Port)
}

func (c *Settopbox) parseResponse(workType string, response []byte) StbControlResponse {
	if len(response) == 0 {
		return StbControlResponse{
			Code:    control_common.ResponseCodeError,
			Message: "no response received from settopbox",
		}
	}

	switch workType {
	case WorkTypeSysCheck, WorkTypeStbRequestInfo:
		var responseMap map[string]string
		if err := json.Unmarshal(response, &responseMap); err != nil {
			return StbControlResponse{
				Code:    control_common.ResponseCodeError,
				Message: fmt.Sprintf("failed to unmarshal settopbox response: %v, response: %s", err, string(response)),
			}
		}

		for _, key := range []string{"loggingTime", "sendingTime"} {
			if value, ok := responseMap[key]; ok {
				if utcTime, err := catv_common.ParseKSTDatetime(value); err != nil {
					slog.Warn("Failed to parse time field in response", "field", key, "value", value, "error", err)
					continue
				} else {
					responseMap[key+"Utc"] = utcTime.String()
				}
			}
		}

		if value, ok := responseMap["runningTime"]; ok {
			if sec, err := catv_common.ParseRunningTimeSec(value); err != nil {
				slog.Warn("Failed to parse running time field in response", "field", "runningTime", "value", value, "error", err)
			} else {
				responseMap["runningTimeSec"] = fmt.Sprintf("%d", sec)
			}
		}

		responseBytes, err := json.Marshal(responseMap)
		if err != nil {
			return StbControlResponse{
				Code:    control_common.ResponseCodeError,
				Message: fmt.Sprintf("failed to marshal response map: %v, responseMap: %v", err, responseMap),
			}
		}

		return StbControlResponse{
			Code:    control_common.ResponseCodeSuccess,
			Message: string(responseBytes),
		}
	default:
		var stbControlResponse StbControlResponse
		if err := json.Unmarshal(response, &stbControlResponse); err != nil {
			return StbControlResponse{
				Code:    control_common.ResponseCodeError,
				Message: fmt.Sprintf("failed to unmarshal settopbox response: %v, response: %s", err, string(response)),
			}
		}
		return stbControlResponse
	}
}

func (c *Settopbox) isPortExhaustionError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	return strings.Contains(errMsg, "cannot assign requested address") ||
		strings.Contains(errMsg, "address already in use") ||
		strings.Contains(errMsg, "too many open files")
}

func (c *Settopbox) isRetryableError(message string) bool {
	if message == "" {
		return false
	}

	retryableErrors := []string{
		"timeout",
		"network unreachable",
		"connection reset",
		"broken pipe",
		"no route to host",
	}

	msgLower := strings.ToLower(message)
	for _, errText := range retryableErrors {
		if strings.Contains(msgLower, errText) {
			return true
		}
	}
	return false
}
