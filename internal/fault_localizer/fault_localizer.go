package fl

import (
	"slices"

	issuetracker "github.com/99pouria/go-apr/internal/fault_localizer/issue-tracker"
	"github.com/99pouria/go-apr/internal/projectenv"
	"github.com/99pouria/go-apr/pkg/logger"
)

func LocalizeFaults(env *projectenv.Environment) (isFixed bool) {
	issues := issuetracker.GetIssues(env)

	var (
		fixedIssues []issuetracker.Issue
	)
	isFixed = false

	logger.PrintInBoxCenter("Program repair statarted", 46)

	defer func() {
		logger.PrintInBoxCenter("Program repair finished", 46)

		switch len(fixedIssues) {
		case 0:
			logger.Println("No bugs found for code.")
			if isFixed {
				logger.Printf("All testcases passed. Code is %s\n", logger.Green("BUG-FREE"))
			} else {
				logger.Printf("Testcases execution %s. Can not repair.\n", logger.Red("FAILED"))
			}
		default:
			logger.Println("Following bugs fixed:")
			for index, f := range fixedIssues {
				logger.Printf("\t[%s] %s\n", logger.Green(index+1), f.Description())
			}
			if isFixed {
				logger.Printf("Repaired code passes all tests. Code is %s\n", logger.Green("BUG-FREE"))
			} else {
				logger.Printf("Repaired code passes more tests, but some still %s.\n", logger.Red("FAIL"))
			}
		}
	}()

	for {
		// running testcases for 100 times
		results := env.RunTestCases(false, 100)
		var failedTestcases []int
		for _, result := range results {
			if !result.Ok {
				failedTestcases = append(failedTestcases, result.ID)
			}
		}

		if len(failedTestcases) == 0 {
			isFixed = true
			return
		}

		for {
			if !issues.Next() {
				return
			}
			issue := issues.Get()

			// trying to update file content to have latest modified file
			if err := env.FuncCode.UpdateCodeContentFromPath(); err != nil {
				logger.Warnf("Can not read file content from its source. This error effects result. Error: %s", err.Error())
			}

			ok, err := issue.Check()
			if err != nil {
				logger.Printf("%s Can not check for bugs.\t%s=%s\n", logger.Red("[ERROR]"), logger.Yellow("Description"), issue.Description())
				continue
			}

			if ok {
				logger.Debugf("Issue checked and the code doesn't have it.\t%s=%s\n", logger.Yellow("Description"), issue.Description())
				continue
			}

			logger.Printf("%s Bug detected. Trying to fix it.\t\t%s=%s\n", logger.Blue("[INFO]"), logger.Yellow("Bug description"), issue.Description())

			if err := issue.Fix(); err != nil {
				logger.Printf("%s Can not fix bug: %s\n", logger.Red("[ERROR]"), err.Error())
				// trying to revet changes applied by Fix method
				if err := issue.Revert(); err != nil {
					logger.Debugf("can not revert applied changes %s\n", err.Error())
				}
				continue
			}

			// running testcases 100 times after applying patch to check if code has improved or not
			results = env.RunTestCases(true, 100)
			var newFailedTestcases []int
			for _, result := range results {
				if !result.Ok {
					newFailedTestcases = append(newFailedTestcases, result.ID)
				}
			}

			// TODO: design a better mechanisem to check if code has improved or not
			improved := false
			for _, id := range newFailedTestcases {
				if !slices.Contains(failedTestcases, id) {
					improved = true
				}
			}

			if len(failedTestcases) > len((newFailedTestcases)) {
				improved = true
			}

			if improved {
				logger.Printf("%s Applying patch was effective.\n", logger.Blue("[INFO]"))
				fixedIssues = append(fixedIssues, issue)
				logger.Printf("%s Rerunning testcases to make sure there isn't any other bugs.\n", logger.Blue("[INFO]"))
				break
			} else {
				logger.Printf("%s Applying patch was not effective. Reverting changes.\n", logger.Yellow("[WARN]"))
				if err := issue.Revert(); err != nil {
					logger.Debugf("Can not revert changes that applied: %s\n", err.Error())
				}
			}
		}
	}
}
