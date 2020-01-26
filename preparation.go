package pinger4

import (
	"context"
	"io"
	"net"
	"sync"

	"umenosuke.net/labelinglog"
)

// AddTarget is AddTarget
func (thisPinger *Pinger) AddTarget(ipAddress string, comment string) {
	select {
	case <-thisPinger.isStarted:
		thisPinger.logger.Log(labelinglog.FlgWarn, "Pinger has already started")
		return
	default:
	}

	binIPAddress := net.ParseIP(ipAddress)
	if binIPAddress == nil {
		resolveIPAddress, err := net.ResolveIPAddr("ip4", ipAddress)
		if err != nil {
			thisPinger.logger.Log(labelinglog.FlgWarn, "parseIP fail : "+err.Error())
			return
		}

		binIPAddress = resolveIPAddress.IP
	}
	if binIPAddress != nil {
		targetID := NetIP2BinIPv4Address(binIPAddress)

		if _, ok := thisPinger.targets[targetID]; !ok {
			thisPinger.targets[targetID] = icmpTarget{
				id:           targetID,
				ipAddress:    ipAddress,
				comment:      comment,
				binIPAddress: binIPAddress,
				netIPAddr:    &net.IPAddr{IP: binIPAddress},
			}
			thisPinger.targetsOrder = append(thisPinger.targetsOrder, targetID)

			list := make([]int64, thisPinger.config.StatisticsCountsNum)
			for i := range list {
				list[i] = 1
			}
			thisPinger.statisticsData.targets[targetID] = &sData{
				Res:   list,
				Index: 0,
			}

			var cancelFuncs [responseListNum]context.CancelFunc
			for i := range cancelFuncs {
				cancelFuncs[i] = func() {}
			}
			thisPinger.timeouterList[targetID] = &struct {
				sync.Mutex
				cancelFunc [responseListNum]context.CancelFunc
			}{
				cancelFunc: cancelFuncs,
			}

			var req [responseListNum]struct {
				isReceived bool
			}
			thisPinger.reqList[targetID] = &struct {
				sync.Mutex
				req [responseListNum]struct {
					isReceived bool
				}
			}{
				req: req,
			}
		} else {
			thisPinger.logger.Log(labelinglog.FlgWarn, "add skip already added : "+ipAddress)
		}
	} else {
		thisPinger.logger.Log(labelinglog.FlgError, "parseIP fail : "+ipAddress)
	}
}

//SetLogEnableLevel a
func (thisPinger *Pinger) SetLogEnableLevel(targetLevelFlgs labelinglog.LogLevel) {
	select {
	case <-thisPinger.isStarted:
		thisPinger.logger.Log(labelinglog.FlgWarn, "Pinger has already started")
		return
	default:
	}

	thisPinger.logger.SetEnableLevel(targetLevelFlgs)
}

//SetLogWriter a
func (thisPinger *Pinger) SetLogWriter(targetLevelFlgs labelinglog.LogLevel, writer io.Writer) {
	select {
	case <-thisPinger.isStarted:
		thisPinger.logger.Log(labelinglog.FlgWarn, "Pinger has already started")
		return
	default:
	}

	thisPinger.logger.SetIoWriter(targetLevelFlgs, writer)
}
