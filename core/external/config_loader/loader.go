package config_loader

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
	"ntels.com/pharos/core/pkg/common"
)

func checkFileReadable(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return err
	}

	// #nosec G304 -- 어떤 상황에서 돌지 모르기때문에 일단 disable
	f, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return nil
}

func Load(configFile []string) (common.Config, error) {
	config := common.Config{}

	if len(configFile) == 0 {
		slog.Error("config_loader config file is empty")
		return common.Config{}, errors.New("config_loader config file is empty")
	}

	viper.SetConfigFile(configFile[0])
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("config_loader file read failed", "error", err)
		return config, err
	}

	for _, file := range configFile[1:] {
		err := checkFileReadable(file)
		if err != nil {
			return common.Config{}, err
		}

		// #nosec G304 -- 어떤 상황에서 돌지 모르기때문에 일단 disable
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}

		err = viper.MergeConfig(f)
		if err != nil {
			slog.Error("config_loader merge config failed", "error", err)
			return common.Config{}, err
		}

		_ = f.Close()
	}

	//viper.SetConfigFile(configFile)
	viper.AutomaticEnv()

	// .으로 구분된 환경변수를 _로 변경
	// [변경 이유]
	// - 환경변수는 보통 POSIX 표준을 따르며, key는 문자(A-Z, a-z), 숫자(0-9), 또는 _로 구성
	// - 일부 쉘(bash, zsh)이나 도구에서 .이 포함된 key를 지원하지 않거나, 구문 분석에서 문제 발생 가능
	// - 설정시 하단과 같은 에러 발생
	//   % export ABC.abc=a
	//   export: not valid in this context: ABC.abc
	//   % export ABC_abc=a
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	if err := viper.Unmarshal(&config); err != nil {
		slog.Error("config_loader decode failed", "error", err)
		return config, err
	}

	// store globally for packages that need runtime access (e.g., TLS settings)
	common.SetGlobalConfig(config)

	return config, nil
}
