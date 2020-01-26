package pinger4

import (
	"context"
	"strconv"
	"sync"
	"time"

	"umenosuke.net/labelinglog"
)

// Run is Pinger start
func (thisPinger *Pinger) Run(ctx context.Context) {
	select {
	case <-thisPinger.isStarted:
		thisPinger.logger.Log(labelinglog.FlgWarn, "Pinger has already started")
		return
	default:
		close(thisPinger.isStarted)
	}

	defer thisPinger.logger.Log(labelinglog.FlgNotice, "finish Pinger")
	thisPinger.logger.Log(labelinglog.FlgNotice, "start Pinger")

	timeoutLimit := thisPinger.config.IntervalMillisec * responseListNum / 2
	if timeoutLimit < thisPinger.config.TimeoutMillisec {
		thisPinger.logger.Log(labelinglog.FlgWarn, "timeout too long, change timeout ["+strconv.FormatInt(thisPinger.config.TimeoutMillisec, 10)+" to "+strconv.FormatInt(timeoutLimit, 10)+"]")
		thisPinger.config.TimeoutMillisec = timeoutLimit
	}

	childCtx, childCtxCancel := context.WithCancel(context.Background())
	defer childCtxCancel()
	thisPinger.cancelFunc = childCtxCancel

	wgChild := sync.WaitGroup{}
	{
		thisPinger.logger.Log(labelinglog.FlgDebug, "start statistics")
		wgChild.Add(1)
		go thisPinger.statistics(childCtx, &wgChild)

		thisPinger.logger.Log(labelinglog.FlgDebug, "start listener")
		wgChild.Add(1)
		go thisPinger.listener(childCtx, &wgChild)

		thisPinger.logger.Log(labelinglog.FlgDebug, "start broker")
		wgChild.Add(1)
		go thisPinger.broker(childCtx, &wgChild)
	}

	if thisPinger.config.DebugEnable {
		wgChild.Add(1)
		go thisPinger.debugStatus(childCtx, &wgChild)
	}

	select {
	case <-ctx.Done():
		thisPinger.logger.Log(labelinglog.FlgDebug, "stop request from parent")
	case <-childCtx.Done():
		thisPinger.logger.Log(labelinglog.FlgDebug, "stop request from child")
		thisPinger.logger.Log(labelinglog.FlgError, "may be fatal error (´・ω・`)")
	}

	thisPinger.logger.Log(labelinglog.FlgDebug, "stop request to all chlid")
	{
		thisPinger.cancelFunc()

		c := make(chan struct{})
		go (func() {
			wgChild.Wait()
			close(c)
		})()

		thisPinger.logger.Log(labelinglog.FlgNotice, "waiting for termination ("+strconv.FormatInt(terminateTimeOutSec, 10)+"sec)")
		select {
		case <-c:
			thisPinger.logger.Log(labelinglog.FlgNotice, "terminated successfully")
		case <-time.After(time.Duration(terminateTimeOutSec) * time.Second):
			thisPinger.logger.Log(labelinglog.FlgError, "forced termination")
		}

		thisPinger.logger.Log(labelinglog.FlgDebug, "chIcmpResponse        "+strconv.Itoa(len(thisPinger.chIcmpResponse)))
		thisPinger.logger.Log(labelinglog.FlgDebug, "chIcmpResponseTimeout "+strconv.Itoa(len(thisPinger.chIcmpResponseTimeout)))
		thisPinger.logger.Log(labelinglog.FlgDebug, "chIcmpResult          "+strconv.Itoa(len(thisPinger.chIcmpResult)))
	}
}
