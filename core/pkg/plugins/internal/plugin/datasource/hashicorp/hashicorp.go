package hashicorp

import (
	"github.com/hashicorp/go-plugin"
)

const (
	PluginTypeDataServer   = "DataServer"
	PluginTypeStreamServer = "StreamServer"
)

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}
