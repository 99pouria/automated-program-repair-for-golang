package issuetracker

import (
	"github.com/99pouria/go-apr/internal/projectenv"
)

type Issue interface {
	Check() (bool, error)
	Fix() error
	Description() string
	Revert() error
}

type Issues struct {
	issues []Issue
}

func GetIssues(env *projectenv.Environment) *Issues {
	return &Issues{issues: []Issue{InitWaitGroupFault(env)}}
}

func (i *Issues) Next() bool {
	return !(len(i.issues) == 0)
}

func (i *Issues) Get() Issue {
	if len(i.issues) == 0 {
		return nil
	}

	newI := i.issues[0]
	i.issues = i.issues[1:]

	return newI
}
