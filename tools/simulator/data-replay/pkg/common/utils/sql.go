package utils

import (
	"encoding/base64"
	"fmt"
)

func GetDsnFromEncoded(dbType string, user string, address string, port int, dbName string, options map[string]string) (string, error) {
	decodedUser, err := base64.StdEncoding.DecodeString(user)
	if err != nil {
		return "", err
	}

	return GetDsn(dbType, string(decodedUser), address, port, dbName, options), nil
}

func GetDsn(dbType string, user string, address string, port int, dbName string, options map[string]string) string {
	optionString := ""
	for k, v := range options {
		if optionString == "" {
			optionString += "?"
		} else {
			optionString += "&"
		}
		optionString += fmt.Sprintf("%s=%s", k, v)
	}

	switch dbType {
	case "postgres":
		return fmt.Sprintf("postgresql://%s@%s:%d/%s%s", user, address, port, dbName, optionString)
	case "mysql":
		return fmt.Sprintf("%s@tcp(%s:%d)/%s%s", user, address, port, dbName, optionString)
	case "mssql":
		return fmt.Sprintf("sqlserver://%s@%s:%d/%s%s", user, address, port, dbName, optionString)
	default:
		return ""
	}
}
