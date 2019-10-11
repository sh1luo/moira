package selfstate

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/moira-alert/moira/notifier/selfstate/heartbeat"

	"gopkg.in/tomb.v2"

	"github.com/moira-alert/moira"
	"github.com/moira-alert/moira/notifier"
	w "github.com/moira-alert/moira/worker"
)

var defaultCheckInterval = time.Second * 10

const selfStateLockName = "moira-self-state-monitor"
const selfStateLockTTL = time.Second * 15

// SelfCheckWorker checks what all notifier services works correctly and send message when moira don't work
type SelfCheckWorker struct {
	Logger     moira.Logger
	Database   moira.Database
	Notifier   notifier.Notifier
	Config     Config
	tomb       tomb.Tomb
	Heartbeats []heartbeat.Heartbeater
}

func (selfCheck *SelfCheckWorker) selfStateChecker(stop <-chan struct{}) error {
	selfCheck.Logger.Info("Moira Notifier Self State Monitor started")

	checkTicker := time.NewTicker(defaultCheckInterval)
	defer checkTicker.Stop()

	nextSendErrorMessage := time.Now().Unix()

	selfCheck.Heartbeats = make([]heartbeat.Heartbeater, 0, 5)
	if heartbeat := heartbeat.GetDatabase(selfCheck.Config.RedisDisconnectDelaySeconds, selfCheck.Logger, selfCheck.Database); heartbeat != nil {
		selfCheck.Heartbeats = append(selfCheck.Heartbeats, heartbeat)
	}

	if heartbeat := heartbeat.GetFilter(selfCheck.Config.LastMetricReceivedDelaySeconds, selfCheck.Logger, selfCheck.Database); heartbeat != nil {
		selfCheck.Heartbeats = append(selfCheck.Heartbeats, heartbeat)
	}

	if heartbeat := heartbeat.GetGraphiteLocalChecker(selfCheck.Config.LastCheckDelaySeconds, selfCheck.Logger, selfCheck.Database); heartbeat != nil {
		selfCheck.Heartbeats = append(selfCheck.Heartbeats, heartbeat)
	}

	if heartbeat := heartbeat.GetGraphiteRemoteChecker(selfCheck.Config.LastRemoteCheckDelaySeconds, selfCheck.Logger, selfCheck.Database); heartbeat != nil {
		selfCheck.Heartbeats = append(selfCheck.Heartbeats, heartbeat)
	}

	if heartbeat := heartbeat.GetNotifier(selfCheck.Logger, selfCheck.Database); heartbeat != nil {
		selfCheck.Heartbeats = append(selfCheck.Heartbeats, heartbeat)
	}

	for {
		select {
		case <-stop:
			selfCheck.Logger.Info("Moira Notifier Self State Monitor stopped")
			return nil
		case <-checkTicker.C:
			nextSendErrorMessage = selfCheck.check(time.Now().Unix(), nextSendErrorMessage)
		}
	}
}

// Start self check worker
func (selfCheck *SelfCheckWorker) Start() error {
	if !selfCheck.Config.Enabled {
		selfCheck.Logger.Debugf("Moira Self State Monitoring disabled")
		return nil
	}
	senders := selfCheck.Notifier.GetSenders()
	if err := selfCheck.Config.checkConfig(senders); err != nil {
		selfCheck.Logger.Errorf("Can't configure Moira Self State Monitoring: %s", err.Error())
		return nil
	}

	selfCheck.tomb.Go(func() error {
		w.NewWorker(
			"Moira Self State Monitoring",
			selfCheck.Logger,
			selfCheck.Database.NewLock(selfStateLockName, selfStateLockTTL),
			selfCheck.selfStateChecker,
		).Run(selfCheck.tomb.Dying())
		return nil
	})

	return nil
}

// Stop self check worker and wait for finish
func (selfCheck *SelfCheckWorker) Stop() error {
	if !selfCheck.Config.Enabled {
		return nil
	}
	senders := selfCheck.Notifier.GetSenders()
	if err := selfCheck.Config.checkConfig(senders); err != nil {
		return nil
	}

	selfCheck.tomb.Kill(nil)
	return selfCheck.tomb.Wait()
}

func (selfCheck *SelfCheckWorker) handleCheckServices(nowTS int64) []moira.NotificationEvent {
	var events []moira.NotificationEvent

	for _, heartbeat := range selfCheck.Heartbeats {
		currentValue, needSend, err := heartbeat.Check(nowTS)
		if err != nil {
			selfCheck.Logger.Error(err)
		}
		if !needSend {
			continue
		}
		events = append(events, generateNotificationEvent(heartbeat.GetErrorMessage(), currentValue))
		if heartbeat.NeedTurnOffNotifier() {
			selfCheck.setNotifierState(moira.SelfStateERROR)
		}
		if !heartbeat.NeedToCheckOthers() {
			break
		}
	}
	return events
}

func (selfCheck *SelfCheckWorker) sendNotification(events []moira.NotificationEvent, nowTS int64) int64 {
	eventsJSON, _ := json.Marshal(events)
	selfCheck.Logger.Errorf("Health check. Send package of %v notification events: %s", len(events), eventsJSON)
	selfCheck.sendErrorMessages(events)
	return nowTS + selfCheck.Config.NoticeIntervalSeconds
}

func (selfCheck *SelfCheckWorker) check(nowTS int64, nextSendErrorMessage int64) int64 {
	if nextSendErrorMessage < nowTS {
		events := selfCheck.handleCheckServices(nowTS)
		if len(events) > 0 {
			nextSendErrorMessage = selfCheck.sendNotification(events, nowTS)
		}
	}
	return nextSendErrorMessage
}

func (selfCheck *SelfCheckWorker) sendErrorMessages(events []moira.NotificationEvent) {
	var sendingWG sync.WaitGroup

	for _, adminContact := range selfCheck.Config.Contacts {
		pkg := notifier.NotificationPackage{
			Contact: moira.ContactData{
				Type:  adminContact["type"],
				Value: adminContact["value"],
			},
			Trigger: moira.TriggerData{
				Name:       "Moira health check",
				ErrorValue: float64(0),
			},
			Events:     events,
			DontResend: true,
		}
		selfCheck.Notifier.Send(&pkg, &sendingWG)
		sendingWG.Wait()
	}
}

func generateNotificationEvent(message string, currentValue int64) moira.NotificationEvent {
	val := float64(currentValue)
	return moira.NotificationEvent{
		Timestamp: time.Now().Unix(),
		OldState:  moira.StateNODATA,
		State:     moira.StateERROR,
		Metric:    message,
		Value:     &val,
	}
}

func (selfCheck *SelfCheckWorker) setNotifierState(state string) {
	err := selfCheck.Database.SetNotifierState(state)
	if err != nil {
		selfCheck.Logger.Errorf("Can't set notifier state: %v", err)
	}
}
