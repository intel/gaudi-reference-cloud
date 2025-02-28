// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"fmt"
)

func (c *CloudConfig) SetStunnelConf(clusterAddr string) error {
	confContent := fmt.Sprintf(`pid = /var/run/stunnel4/stunnel.pid
socket = r:TCP_NODELAY=1

[nfs4]
client = yes
accept = localhost:2049
connect = %s
ciphers = ALL
sslVersionMin = TLSv1.2
sslVersionMax = TLSv1.3
verifyChain = yes
CAPath = /etc/ssl/certs
`, clusterAddr)

	c.AddWriteFileWithPermissions("/etc/stunnel/stunnel.conf", confContent, "0644")
	// step to restart stunnel4
	c.AddRunCmd("[ -x /usr/bin/stunnel ] && sudo systemctl restart stunnel4")
	return nil
}
