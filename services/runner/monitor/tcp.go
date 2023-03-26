package monitor

import (
	"monitr/services/common/types"
	"net"
	"time"
)

func (m *Monitor) runTCP() (*types.Result, map[string]any) {
	r := types.NewResult(m.Name, m.Target, m.ID)

	var err error

	timeout := time.Duration(5) * time.Second

	timeoutProp := m.Properties["timeout"]
	if timeoutProp != "" {
		timeout, err = time.ParseDuration(timeoutProp)
		if err != nil {
			return types.NewFailedResult(m.Name, m.Target, m.ID, err), nil
		}
	}

	dialer := net.Dialer{Timeout: timeout}
	start := time.Now()

	conn, err := dialer.Dial("tcp", m.Target)
	if err != nil {
		return types.NewFailedResult(m.Name, m.Target, m.ID, err), nil
	}

	r.Value = int(time.Since(start).Milliseconds())

	outputs := map[string]any{
		"respTime": r.Value,
		"address":  conn.RemoteAddr().String(),
	}

	defer conn.Close()

	return r, outputs
}
