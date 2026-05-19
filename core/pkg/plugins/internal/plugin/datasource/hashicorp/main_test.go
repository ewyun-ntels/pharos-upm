package hashicorp

import (
	"testing"

	"github.com/hashicorp/go-plugin"
)

func TestHandshakeConfig_Constants(t *testing.T) {
	// HandshakeConfig의 상수 값들이 올바르게 설정되어 있는지 확인
	if HandshakeConfig.ProtocolVersion != 1 {
		t.Errorf("Expected ProtocolVersion 1, got %d", HandshakeConfig.ProtocolVersion)
	}

	if HandshakeConfig.MagicCookieKey != "BASIC_PLUGIN" {
		t.Errorf("Expected MagicCookieKey 'BASIC_PLUGIN', got %s", HandshakeConfig.MagicCookieKey)
	}

	if HandshakeConfig.MagicCookieValue != "hello" {
		t.Errorf("Expected MagicCookieValue 'hello', got %s", HandshakeConfig.MagicCookieValue)
	}
}

func TestPluginType_Constants(t *testing.T) {
	// 플러그인 타입 상수들이 올바르게 정의되어 있는지 확인
	if PluginTypeDataServer != "DataServer" {
		t.Errorf("Expected PluginTypeDataServer 'DataServer', got %s", PluginTypeDataServer)
	}

	if PluginTypeStreamServer != "StreamServer" {
		t.Errorf("Expected PluginTypeStreamServer 'StreamServer', got %s", PluginTypeStreamServer)
	}
}

func TestMain_PluginSet_Creation(t *testing.T) {
	// 기본적인 플러그인 셋이 생성되는지 확인
	pluginSet := plugin.PluginSet{
		PluginTypeDataServer: &DataServerPlugin{
			DataServer: &mockDataServer{},
		},
		PluginTypeStreamServer: &StreamServerPlugin{
			StreamServer: &mockStreamServer{},
		},
	}

	if pluginSet[PluginTypeDataServer] == nil {
		t.Error("Expected DataServerPlugin to be set")
	}

	if pluginSet[PluginTypeStreamServer] == nil {
		t.Error("Expected StreamServerPlugin to be set")
	}
}
