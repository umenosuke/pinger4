package pinger4

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"

	"umenosuke.net/labelinglog"
)

// AddTarget is AddTarget
func (thisPinger *Pinger) AddTarget(ipAddress string, comment string) error {
	select {
	case <-thisPinger.isStarted:
		msg := "Pinger has already started"
		thisPinger.logger.Log(labelinglog.FlgWarn, msg)
		return errors.New(msg)
	default:
	}

	binIPAddress := net.ParseIP(ipAddress)
	if binIPAddress == nil {
		resolveIPAddress, err := net.ResolveIPAddr("ip4", ipAddress)
		if err != nil {
			msg := "parseIP fail : " + err.Error()
			thisPinger.logger.Log(labelinglog.FlgWarn, msg)
			return errors.New(msg)
		}

		binIPAddress = resolveIPAddress.IP
	}

	if binIPAddress == nil {
		msg := "parseIP fail : " + ipAddress
		thisPinger.logger.Log(labelinglog.FlgWarn, msg)
		return errors.New(msg)
	}

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
		msg := "add skip already added : " + ipAddress
		thisPinger.logger.Log(labelinglog.FlgWarn, msg)
		return errors.New(msg)
	}

	return nil
}

//SetLogEnableLevel a
func (thisPinger *Pinger) SetLogEnableLevel(targetLevelFlgs labelinglog.LogLevel) error {
	select {
	case <-thisPinger.isStarted:
		msg := "Pinger has already started"
		thisPinger.logger.Log(labelinglog.FlgWarn, msg)
		return errors.New(msg)
	default:
	}

	thisPinger.logger.SetEnableLevel(targetLevelFlgs)
	return nil
}

//SetLogWriter a
func (thisPinger *Pinger) SetLogWriter(targetLevelFlgs labelinglog.LogLevel, writer io.Writer) error {
	select {
	case <-thisPinger.isStarted:
		msg := "Pinger has already started"
		thisPinger.logger.Log(labelinglog.FlgWarn, msg)
		return errors.New(msg)
	default:
	}

	thisPinger.logger.SetIoWriter(targetLevelFlgs, writer)
	return nil
}
