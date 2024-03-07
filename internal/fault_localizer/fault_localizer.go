package fl

import (
	"slices"

	"github.com/99pouria/go-apr/internal/fault_localizer/faults"
	"github.com/99pouria/go-apr/internal/projectenv"
	"github.com/sirupsen/logrus"
)

func LocalizeFaults(env *projectenv.Environment) {
	f := faults.GetFaults(env)

	for _, fault := range f {

		ok, err := fault.Check()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"fault": fault.Description(),
			}).Error("Check failed")
			continue
		}

		if ok {
			logrus.WithField("fault", fault.Description()).Info("No fault")
			continue
		}

		logrus.WithField("fault", fault.Description()).Warn("Fault detected, Trying to fix fault")

		// store testcases that failed
		results := env.RunTestCases(false)
		var failedTestcases []int
		for _, result := range results {
			if !result.Ok {
				failedTestcases = append(failedTestcases, result.ID)
			}
		}

		if err := fault.Fix(); err != nil {
			logrus.WithField("fault", fault.Description()).Info("Can not fix fault")
			// trying to revet changes applied by Fix method
			fault.Revert()
		}

		// rerun with testcases TODO: import comment and refactor variables names
		results = env.RunTestCases(false)
		var newFailedTestcases []int
		for _, result := range results {
			if !result.Ok {
				newFailedTestcases = append(newFailedTestcases, result.ID)
			}
		}

		// check result
		improved := false

		if len(newFailedTestcases) == 0 {
			logrus.WithField("fault", fault.Description()).Info("All testcases passes!")
			return
		}

		for _, id := range newFailedTestcases {
			if !slices.Contains(failedTestcases, id) {
				improved = true
			}
		}

		if improved {
			logrus.WithField("fault", fault.Description()).Info("Some testcases passes after using the patch")
		} else {
			logrus.WithField("fault", fault.Description()).Warn("Patch wasn't effective. Reverting changes")
			// TODO: handle error
			fault.Revert()
		}
	}
}
