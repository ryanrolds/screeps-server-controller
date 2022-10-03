package screeps

import (
	_ "embed"
)

//go:embed private-server-config.yaml.tmpl
var PrivateServerConfigTemplate string
