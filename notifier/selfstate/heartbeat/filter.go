package heartbeat

import (
	"github.com/moira-alert/moira"
	"time"
)

type filter struct {
	heartbeat
	count int64
}

func GetFilter(delay int64, logger moira.Logger, database moira.Database) Heartbeater {
	if delay > 0 {
		return &filter{heartbeat: heartbeat{
			logger:              logger,
			database:            database,
			delay:               delay,
			lastSuccessfulCheck: time.Now().Unix(),
		}}
	}
	return nil
}

func (check *filter) Check(nowTS int64) (int64, bool, error) {
	metricsCount, err := check.database.GetMetricsUpdatesCount()
	if err != nil {
		return 0, false, err
	}

	if check.count != metricsCount {
		check.count = metricsCount
		check.lastSuccessfulCheck = nowTS
		return 0, false, nil
	}

	if check.lastSuccessfulCheck < nowTS-check.heartbeat.delay {
		check.logger.Errorf(templateMoreThanMessage, check.GetErrorMessage(), nowTS-check.heartbeat.lastSuccessfulCheck)
		return nowTS - check.heartbeat.lastSuccessfulCheck, true, nil
	}
	return 0, false, nil
}

func (filter) NeedTurnOffNotifier() bool {
	return true
}

func (check filter) NeedToCheckOthers() bool {
	metricsCount, _ := check.database.GetMetricsUpdatesCount()
	return metricsCount > 0
}

func (filter) GetErrorMessage() string {
	return "Moira-Filter does not receive metrics"
}
