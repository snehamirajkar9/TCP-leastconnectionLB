package forwarder

import (
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	healthCheckTimeout  = 1 * time.Second
	healthCheckInterval = 10 * time.Second
)

func (u *upstream) checkHealthy() {
	conn, err := net.DialTimeout("tcp", u.addr, healthCheckTimeout)

	if err != nil {
		u.setUnhealthy()
	} else {
		u.setHealthy()
	}
	if conn != nil {
		conn.Close()
	}
}

func (u *upstream) startHealthchecks() chan bool {
	ticker := time.NewTicker(healthCheckInterval)

	stop := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-ticker.C:
				u.checkHealthy()
			case <-stop:
				ticker.Stop()
			}
		}
	}()

	return stop
}

func (u *upstream) setUnhealthy() {
	swapped := u.healthy.CompareAndSwap(true, false)
	if swapped {
		log.Debug().Str("upstream", u.addr).Msg("unhealthy")
	}
}

func (u *upstream) setHealthy() {
	swapped := u.healthy.CompareAndSwap(false, true)
	if swapped {
		log.Debug().Str("upstream", u.addr).Msg("healthy")
	}
}

func (u *upstream) isUnhealthy() bool {
	return !u.healthy.Load()
}
