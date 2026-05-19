package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the simulator configuration
type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Port    InterfaceInfo  `yaml:"port"`
	Devices []DeviceConfig `yaml:"devices"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
}

// DeviceConfig represents a single STB device configuration
type DeviceConfig struct {
	HostID   string `yaml:"hostId"`
	MacAddr  string `yaml:"macAddr"`
	CmMac    string `yaml:"cmMac"`
	StbIP    string `yaml:"stbIp"`
	CmIP     string `yaml:"cmIp"`
	StbModel string `yaml:"stbModel"`
	MwVer    string `yaml:"mwVer"`
	LocalVer string `yaml:"localVer"`
	CloudVer string `yaml:"cloudVer"`
}

// InterfaceInfo contains port information for each interface
type InterfaceInfo struct {
	Periodic                 int `yaml:"periodic"`                   // 주기전송 (10분)
	Daily                    int `yaml:"daily"`                      // 일일전송
	QualityMeasurement       int `yaml:"quality_measurement"`        // 품질계측전송
	Diagnostic               int `yaml:"diagnostic"`                 // 자가진단전송
	NetworkQualityTransition int `yaml:"network_quality_transition"` // 망품질전환전송
}

// Global configuration
var (
	config     *Config
	configLock sync.RWMutex
)

// PeriodicData represents 주기전송 (10분마다 전송)
type PeriodicData struct {
	HostID      string  `json:"hostId"`
	MacAddr     string  `json:"macAddr"`
	CmMac       string  `json:"cmMac"`
	StbIP       string  `json:"stbIp"`
	CmIP        string  `json:"cmIp"`
	StbModel    string  `json:"stbModel"`
	MwVer       string  `json:"mwVer"`
	LocalVer    string  `json:"localVer"`
	CloudVer    string  `json:"cloudVer"`
	LoggingTime string  `json:"loggingTime"`
	SendingTime string  `json:"sendingTime"`
	ChSid       any     `json:"chSid,omitempty"`
	ChNum       any     `json:"chNum,omitempty"`
	ChName      *string `json:"chName,omitempty"`
	ChPrg       *string `json:"chPrg,omitempty"`
	ChFreq      *string `json:"chFreq,omitempty"`
	ChMode      *string `json:"chMode,omitempty"`
	PwrLvl      *string `json:"pwrLvl,omitempty"`
	Snr         *string `json:"snr,omitempty"`
	SigWeak     *string `json:"sigWeak,omitempty"`
	SigWeakCnt  any     `json:"sigWeakCnt,omitempty"`
	StbState    *string `json:"stbState,omitempty"`
	RunningTime *string `json:"runningTime,omitempty"`
}

// DailyData represents 일일전송 (하루 1회)
type DailyData struct {
	HostID        string  `json:"hostId"`
	MacAddr       string  `json:"macAddr"`
	CmMac         string  `json:"cmMac"`
	StbIP         string  `json:"stbIp"`
	CmIP          string  `json:"cmIp"`
	StbModel      string  `json:"stbModel"`
	MwVer         string  `json:"mwVer"`
	LocalVer      string  `json:"localVer"`
	CloudVer      string  `json:"cloudVer"`
	LoggingTime   string  `json:"loggingTime"`
	SendingTime   string  `json:"sendingTime"`
	LimitAge      any     `json:"limitAge,omitempty"`
	TvLock        *string `json:"tvLock,omitempty"`
	SkipCh        *string `json:"skipCh,omitempty"`
	EasyBuying    *string `json:"easyBuying,omitempty"`
	FavCh         *string `json:"favCh,omitempty"`
	ZappingAd     any     `json:"zappingAd,omitempty"`
	MiniEpg       any     `json:"miniEpg,omitempty"`
	MiniEpgAd     *string `json:"miniEpgAd,omitempty"`
	TvCaption     any     `json:"tvCaption,omitempty"`
	TvImpaired    *string `json:"tvImpaired,omitempty"`
	BarkerCh      *string `json:"barkerCh,omitempty"`
	VodView       *string `json:"vodView,omitempty"`
	VodRelay      *string `json:"vodRelay,omitempty"`
	Resolution    any     `json:"resolution,omitempty"`
	AudioMode     any     `json:"audioMode,omitempty"`
	HdmiCec       *string `json:"hdmiCec,omitempty"`
	Hdcp          *string `json:"hdcp,omitempty"`
	Hdr           any     `json:"hdr,omitempty"`
	MobilePay     *string `json:"mobilePay,omitempty"`
	MorningAlarm  *string `json:"morningAlarm,omitempty"`
	BootMenu      *string `json:"bootMenu,omitempty"`
	PmsOn         *string `json:"pmsOn,omitempty"`
	OneAdOn       *string `json:"oneAdOn,omitempty"`
	AudioLang     *string `json:"audioLang,omitempty"`
	StandbyMode   *string `json:"standbyMode,omitempty"`
	SavePwr       any     `json:"savePwr,omitempty"`
	VoiceGuide    *string `json:"voiceGuide,omitempty"`
	RunningTime   *string `json:"runningTime,omitempty"`
	LimitContents *string `json:"limitContents,omitempty"`
}

// QualityMeasurementData represents 품질계측전송 (Sleep 모드 진입)
type QualityMeasurementData struct {
	HostID      string           `json:"hostId"`
	MacAddr     string           `json:"macAddr"`
	CmMac       string           `json:"cmMac"`
	StbIP       string           `json:"stbIp"`
	CmIP        string           `json:"cmIp"`
	StbModel    string           `json:"stbModel"`
	MwVer       string           `json:"mwVer"`
	LocalVer    string           `json:"localVer"`
	CloudVer    string           `json:"cloudVer"`
	LoggingTime string           `json:"loggingTime"`
	Channels    []ChannelQuality `json:"channels"`
}

// ChannelQuality represents channel quality information
type ChannelQuality struct {
	ChSid  any    `json:"chSid"`
	ChNum  any    `json:"chNum"`
	ChFreq any    `json:"chFreq"`
	ChMode string `json:"chMode"`
	PwrLvl any    `json:"pwrLvl"`
	Snr    any    `json:"snr"`
}

// DiagnosticData represents 자가진단전송 (핫키 입력 시)
// DiagnosticData represents UDP 자가진단전송 (핫키 *106OK 입력 시)
// Note: TCP 자가진단 제어 (원격 요청)는 별도로 구현 필요 (serverIp:port:8801)
type DiagnosticData struct {
	HostID      string `json:"hostId"`
	MacAddr     string `json:"macAddr"`
	CmMac       string `json:"cmMac"`
	StbIP       string `json:"stbIp"`
	CmIP        string `json:"cmIp"`
	StbModel    string `json:"stbModel"`
	MwVer       string `json:"mwVer"`
	LocalVer    string `json:"localVer"`
	CloudVer    string `json:"cloudVer"`
	LoggingTime string `json:"loggingTime"`
	ChSid       any    `json:"chSid"`
	ChNum       any    `json:"chNum"`
	ChFreq      any    `json:"chFreq"`
	ChMode      string `json:"chMode"`
	PwrLvl      any    `json:"pwrLvl"`
	Snr         any    `json:"snr"`
}

// NetworkQualityTransitionData represents 망품질전환전송 (QAM > 8VSB)
type NetworkQualityTransitionData struct {
	HostID       string  `json:"hostId"`
	MacAddr      string  `json:"macAddr"`
	CmMac        string  `json:"cmMac"`
	StbIP        string  `json:"stbIp"`
	CmIP         string  `json:"cmIp"`
	StbModel     string  `json:"stbModel"`
	MwVer        string  `json:"mwVer"`
	LocalVer     string  `json:"localVer"`
	CloudVer     string  `json:"cloudVer"`
	LoggingTime  string  `json:"loggingTime"`
	ChSid        any     `json:"chSid"`
	ChNum        any     `json:"chNum"`
	ChName       *string `json:"chName,omitempty"`
	ChQamFreq    any     `json:"chQamFreq"`    // v2.5: qamChFreq → chQamFreq
	ChQamMode    string  `json:"chQamMode"`    // v2.5: qamChMode → chQamMode
	ChQamPwrLvl  string  `json:"chQamPwrLvl"`  // v2.5: qamChPwrLvl → chQamPwrLvl
	ChQamSnr     string  `json:"chQamSnr"`     // v2.5: qamChSnr → chQamSnr
	Ch8vsbFreq   any     `json:"ch8vsbFreq"`   // v2.5: vsbChFreq → ch8vsbFreq
	Ch8vsbMode   string  `json:"ch8vsbMode"`   // v2.5: vsbChMode → ch8vsbMode
	Ch8vsbPwrLvl any     `json:"ch8vsbPwrLvl"` // v2.5: vsbChPwrLvl → ch8vsbPwrLvl
	Ch8vsbSnr    any     `json:"ch8vsbSnr"`    // v2.5: vsbChSnr → ch8vsbSnr
}

func main() {
	configFile := flag.String("config", "", "Path to configuration file (JSON)")
	flag.Parse()

	if *configFile == "" {
		log.Fatal("Usage: simulator -config <config-file.yml>")
	}

	// Load configuration
	if err := loadConfig(*configFile); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Loaded configuration with %d devices", len(config.Devices))
	log.Printf("Server IP: %s", config.Server.Host)
	log.Printf("Interfaces: Periodic=%d, Daily=%d, QualityMeasurement=%d, Diagnostic=%d, NetworkQualityTransition=%d",
		config.Port.Periodic,
		config.Port.Daily,
		config.Port.QualityMeasurement,
		config.Port.Diagnostic,
		config.Port.NetworkQualityTransition)

	// Start simulators for each device
	var wg sync.WaitGroup
	for _, device := range config.Devices {
		wg.Add(1)
		go func(dev DeviceConfig) {
			defer wg.Done()
			runDeviceSimulator(dev)
		}(device)
	}

	log.Println("All device simulators started. Press Ctrl+C to stop.")
	wg.Wait()
}

// loadConfig loads configuration from YAML file
func loadConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	configLock.Lock()
	defer configLock.Unlock()

	config = &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config YAML: %w", err)
	}

	return nil
}

// runDeviceSimulator runs all interface simulators for a single device
func runDeviceSimulator(device DeviceConfig) {
	log.Printf("[%s] Starting device simulator", device.HostID)

	// Start periodic transmission (every 10 minutes)
	go sendPeriodic(device, 10*time.Minute)

	// Start daily transmission (once per day at midnight)
	go sendDaily(device)

	// Start quality measurement (every 1 minute for simulation)
	go sendQualityMeasurement(device, 1*time.Minute)

	// Start diagnostic (every 1 minute for simulation)
	go sendDiagnostic(device, 1*time.Minute)

	// Start network quality transition (every 1 minute for simulation)
	go sendNetworkQualityTransition(device, 1*time.Minute)

	// Keep alive
	select {}
}

// sendPeriodic sends periodic data every interval
func sendPeriodic(device DeviceConfig, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Send immediately on start
	sendPeriodicData(device)

	for range ticker.C {
		sendPeriodicData(device)
	}
}

func sendPeriodicData(device DeviceConfig) {
	now := time.Now()

	chName := "EBS HD"
	chPrg := "일단 해봐요 생방송"
	chFreq := "550MHz"
	chMode := "256QAM"
	pwrLvl := fmt.Sprintf("%.1fdBmV", rand.Float64()*24-12) // -12 ~ 12
	snr := fmt.Sprintf("%.1fdB", rand.Float64()*10+30)      // 30 ~ 40
	sigWeak := "N"
	stbState := "watching"
	runningTime := fmt.Sprintf("%dday %dhour %dmin %dsec", rand.IntN(30), rand.IntN(24), rand.IntN(60), rand.IntN(60))

	data := PeriodicData{
		HostID:      device.HostID,
		MacAddr:     device.MacAddr,
		CmMac:       device.CmMac,
		StbIP:       device.StbIP,
		CmIP:        device.CmIP,
		StbModel:    device.StbModel,
		MwVer:       device.MwVer,
		LocalVer:    device.LocalVer,
		CloudVer:    device.CloudVer,
		LoggingTime: now.Format("2006/01/02 15:04"),
		SendingTime: now.Format("2006/01/02 15:04"),
		ChName:      &chName,
		ChPrg:       &chPrg,
		ChFreq:      &chFreq,
		ChMode:      &chMode,
		PwrLvl:      &pwrLvl,
		Snr:         &snr,
		SigWeak:     &sigWeak,
		StbState:    &stbState,
		RunningTime: &runningTime,
	}

	switch rand.IntN(2) {
	case 0:
		data.ChSid = "185"
		data.ChNum = "1"
		data.SigWeakCnt = "1"
	case 1:
		data.ChSid = 100
		data.ChNum = 2
		data.SigWeakCnt = 2
	}

	configLock.RLock()
	port := config.Port.Periodic
	serverIP := config.Server.Host
	configLock.RUnlock()

	if err := sendUDP(serverIP, port, data); err != nil {
		log.Printf("[%s] Failed to send periodic data: %v", device.HostID, err)
	} else {
		log.Printf("[%s] Sent periodic data to %s:%d", device.HostID, serverIP, port)
	}
}

// sendDaily sends daily data once per day
func sendDaily(device DeviceConfig) {
	// Send immediately on start
	sendDailyData(device)

	// Calculate next midnight
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	// Wait until midnight
	time.Sleep(time.Until(nextMidnight))

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Send at midnight
	sendDailyData(device)

	for range ticker.C {
		sendDailyData(device)
	}
}

func sendDailyData(device DeviceConfig) {
	now := time.Now()

	tvLock := "Off"
	skipCh := "11,10,15"
	easyBuying := "NoOpt"
	favCh := "1,3,5"
	miniEpgAd := "On"
	tvImpaired := "Off"
	barkerCh := "Y"
	vodView := "Poster"
	vodRelay := "On"
	hdmiCec := "false"
	hdcp := "on"
	mobilePay := "N"
	morningAlarm := "repeatSetting:1,channelSetting:111,channelNum:11,channelName:MBC,timeSetting:0700"
	bootMenu := "On"
	pmsOn := "On"
	oneAdOn := "true"
	audioLang := "kor"
	standbyMode := "Off"
	voiceGuide := "On"
	runningTime := fmt.Sprintf("%dday %dhour %dmin %dsec", rand.IntN(30), rand.IntN(24), rand.IntN(60), rand.IntN(60))
	limitContents := "Hide"

	data := DailyData{
		HostID:        device.HostID,
		MacAddr:       device.MacAddr,
		CmMac:         device.CmMac,
		StbIP:         device.StbIP,
		CmIP:          device.CmIP,
		StbModel:      device.StbModel,
		MwVer:         device.MwVer,
		LocalVer:      device.LocalVer,
		CloudVer:      device.CloudVer,
		LoggingTime:   now.Format("2006/01/02 15:04"),
		SendingTime:   now.Format("2006/01/02 15:04"),
		TvLock:        &tvLock,
		SkipCh:        &skipCh,
		EasyBuying:    &easyBuying,
		FavCh:         &favCh,
		MiniEpgAd:     &miniEpgAd,
		TvImpaired:    &tvImpaired,
		BarkerCh:      &barkerCh,
		VodView:       &vodView,
		VodRelay:      &vodRelay,
		HdmiCec:       &hdmiCec,
		Hdcp:          &hdcp,
		MobilePay:     &mobilePay,
		MorningAlarm:  &morningAlarm,
		BootMenu:      &bootMenu,
		PmsOn:         &pmsOn,
		OneAdOn:       &oneAdOn,
		AudioLang:     &audioLang,
		StandbyMode:   &standbyMode,
		VoiceGuide:    &voiceGuide,
		RunningTime:   &runningTime,
		LimitContents: &limitContents,
	}

	switch rand.IntN(2) {
	case 0:
		data.LimitAge = any("0")
		data.ZappingAd = &[]string{"1", "2", "3"}[rand.IntN(3)]
		data.MiniEpg = &[]string{"1", "2", "3"}[rand.IntN(3)]
		data.TvCaption = &[]string{"1", "2"}[rand.IntN(2)]
		data.Resolution = &[]string{"1", "2"}[rand.IntN(2)]
		data.AudioMode = &[]string{"1", "2"}[rand.IntN(2)]
		data.Hdr = &[]string{"1", "2"}[rand.IntN(2)]
		data.SavePwr = &[]string{"100000", "200000"}[rand.IntN(2)]
	case 1:
		data.LimitAge = any(10)
		data.ZappingAd = &[]int{10, 20, 30}[rand.IntN(3)]
		data.MiniEpg = &[]int{10, 20, 30}[rand.IntN(3)]
		data.TvCaption = &[]int{10, 20}[rand.IntN(2)]
		data.Resolution = &[]int{10, 20}[rand.IntN(2)]
		data.AudioMode = &[]int{20, 30}[rand.IntN(2)]
		data.Hdr = &[]int{10, 20}[rand.IntN(2)]
		data.SavePwr = &[]int{300000, 600000}[rand.IntN(2)]
	}

	configLock.RLock()
	port := config.Port.Daily
	serverIP := config.Server.Host
	configLock.RUnlock()

	if err := sendUDP(serverIP, port, data); err != nil {
		log.Printf("[%s] Failed to send daily data: %v", device.HostID, err)
	} else {
		log.Printf("[%s] Sent daily data to %s:%d", device.HostID, serverIP, port)
	}
}

// sendQualityMeasurement sends quality measurement data
func sendQualityMeasurement(device DeviceConfig, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Send immediately on start
	sendQualityMeasurementData(device)

	for range ticker.C {
		sendQualityMeasurementData(device)
	}
}

func sendQualityMeasurementData(device DeviceConfig) {
	now := time.Now()

	// Generate sample channel quality data
	channels := []ChannelQuality{
		{
			ChSid:  "185",
			ChNum:  "3",
			ChFreq: "550",
			ChMode: "256QAM",
			PwrLvl: "1",
			Snr:    "1",
		},
		{
			ChSid:  100,
			ChNum:  10,
			ChFreq: 600,
			ChMode: "256QAM",
			PwrLvl: 2,
			Snr:    2,
		},
	}

	data := QualityMeasurementData{
		HostID:      device.HostID,
		MacAddr:     device.MacAddr,
		CmMac:       device.CmMac,
		StbIP:       device.StbIP,
		CmIP:        device.CmIP,
		StbModel:    device.StbModel,
		MwVer:       device.MwVer,
		LocalVer:    device.LocalVer,
		CloudVer:    device.CloudVer,
		LoggingTime: now.Format("2006/01/02 15:04"),
		Channels:    channels,
	}

	configLock.RLock()
	port := config.Port.QualityMeasurement
	serverIP := config.Server.Host
	configLock.RUnlock()

	if err := sendUDP(serverIP, port, data); err != nil {
		log.Printf("[%s] Failed to send quality measurement data: %v", device.HostID, err)
	} else {
		log.Printf("[%s] Sent quality measurement data to %s:%d", device.HostID, serverIP, port)
	}
}

// sendDiagnostic sends diagnostic data
func sendDiagnostic(device DeviceConfig, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Send immediately on start
	sendDiagnosticData(device)

	for range ticker.C {
		sendDiagnosticData(device)
	}
}

func sendDiagnosticData(device DeviceConfig) {
	now := time.Now()

	data := DiagnosticData{
		HostID:      device.HostID,
		MacAddr:     device.MacAddr,
		CmMac:       device.CmMac,
		StbIP:       device.StbIP,
		CmIP:        device.CmIP,
		StbModel:    device.StbModel,
		MwVer:       device.MwVer,
		LocalVer:    device.LocalVer,
		CloudVer:    device.CloudVer,
		LoggingTime: now.Format("2006/01/02 15:04"),
		ChMode:      "256QAM",
	}

	switch rand.IntN(2) {
	case 0:
		data.ChSid = "185"
		data.ChNum = "1"
		data.ChFreq = "1"
		data.PwrLvl = "1"
		data.Snr = "1"
	case 1:
		data.ChSid = 100
		data.ChNum = 2
		data.ChFreq = 2
		data.PwrLvl = 2
		data.Snr = 2
	}

	configLock.RLock()
	port := config.Port.Diagnostic
	serverIP := config.Server.Host
	configLock.RUnlock()

	if err := sendUDP(serverIP, port, data); err != nil {
		log.Printf("[%s] Failed to send diagnostic data: %v", device.HostID, err)
	} else {
		log.Printf("[%s] Sent diagnostic data to %s:%d", device.HostID, serverIP, port)
	}
}

// sendNetworkQualityTransition sends network quality transition data
func sendNetworkQualityTransition(device DeviceConfig, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Send immediately on start
	sendNetworkQualityTransitionData(device)

	for range ticker.C {
		sendNetworkQualityTransitionData(device)
	}
}

func sendNetworkQualityTransitionData(device DeviceConfig) {
	now := time.Now()

	chName := "EBS HD"

	data := NetworkQualityTransitionData{
		HostID:       device.HostID,
		MacAddr:      device.MacAddr,
		CmMac:        device.CmMac,
		StbIP:        device.StbIP,
		CmIP:         device.CmIP,
		StbModel:     device.StbModel,
		MwVer:        device.MwVer,
		LocalVer:     device.LocalVer,
		CloudVer:     device.CloudVer,
		LoggingTime:  now.Format("2006/01/02 15:04"),
		ChSid:        "185",
		ChNum:        "3",
		ChName:       &chName,
		ChQamFreq:    "550MHz",                                      // v2.5 updated
		ChQamMode:    "256QAM",                                      // v2.5 updated
		ChQamPwrLvl:  fmt.Sprintf("%.1fdBmV", rand.Float64()*24-12), // v2.5 updated
		ChQamSnr:     fmt.Sprintf("%.1fdB", rand.Float64()*10+30),   // v2.5 updated
		Ch8vsbFreq:   "560MHz",                                      // v2.5 updated
		Ch8vsbMode:   "8VSB",                                        // v2.5 updated
		Ch8vsbPwrLvl: fmt.Sprintf("%.1fdBmV", rand.Float64()*24-12), // v2.5 updated
		Ch8vsbSnr:    fmt.Sprintf("%.1fdB", rand.Float64()*10+30),   // v2.5 updated
	}

	switch rand.IntN(2) {
	case 0:
		data.ChSid = "185"
		data.ChNum = "1"
		data.ChQamFreq = "1"
		data.Ch8vsbFreq = "1"
		data.Ch8vsbPwrLvl = "1"
		data.Ch8vsbSnr = "1"
	case 1:
		data.ChSid = 100
		data.ChNum = 2
		data.ChQamFreq = 2
		data.Ch8vsbFreq = 2
		data.Ch8vsbPwrLvl = 2
		data.Ch8vsbSnr = 2
	}

	configLock.RLock()
	port := config.Port.NetworkQualityTransition
	serverIP := config.Server.Host
	configLock.RUnlock()

	if err := sendUDP(serverIP, port, data); err != nil {
		log.Printf("[%s] Failed to send network quality transition data: %v", device.HostID, err)
	} else {
		log.Printf("[%s] Sent network quality transition data to %s:%d", device.HostID, serverIP, port)
	}
}

// sendUDP sends data via UDP to the specified address (JSON format)
func sendUDP(serverIP string, port int, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	addr := net.JoinHostPort(serverIP, fmt.Sprintf("%d", port))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write UDP: %w", err)
	}

	return nil
}
