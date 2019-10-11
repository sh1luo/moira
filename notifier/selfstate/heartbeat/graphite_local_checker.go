package heartbeat

import (
	"github.com/moira-alert/moira"
	"time"
)

type graphiteLocalChecker struct {
	heartbeat
	count int64
}

func GetGraphiteLocalChecker(delay int64, logger moira.Logger, database moira.Database) Heartbeater {
	if delay > 0 {
		return &graphiteLocalChecker{heartbeat: heartbeat{
			logger:              logger,
			database:            database,
			delay:               delay,
			lastSuccessfulCheck: time.Now().Unix(),
		}}
	}
	return nil
}

func (check *graphiteLocalChecker) Check(nowTS int64) (int64, bool, error) {
	checksCont, err := check.database.GetChecksUpdatesCount()
	if err != nil {
		return 0, false, err
	}

	if check.count != checksCont {
		check.count = checksCont
		check.lastSuccessfulCheck = nowTS
		return 0, false, nil
	}

	if check.lastSuccessfulCheck < nowTS-check.delay {
		check.logger.Errorf(templateMoreThanMessage, check.GetErrorMessage(), nowTS-check.lastSuccessfulCheck)
		return nowTS - check.lastSuccessfulCheck, true, nil
	}
	return 0, false, nil
}

func (graphiteLocalChecker) NeedTurnOffNotifier() bool {
	return false
}

func (check graphiteLocalChecker) NeedToCheckOthers() bool {
	checksCont, _ := check.database.GetChecksUpdatesCount()
	return checksCont > 0
}

func (graphiteLocalChecker) GetErrorMessage() string {
	return "Moira-Checker does not check triggers"
}
