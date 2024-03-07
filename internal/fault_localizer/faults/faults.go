package faults

import "github.com/99pouria/go-apr/internal/projectenv"

type Fault interface {
	Check() (bool, error)
	Fix() error
	Description() string
	Revert() error
}

func GetFaults(env *projectenv.Environment) (faults []Fault) {

	wgFault := InitWaitGroupFault(env)

	faults = append(faults, wgFault)
	return
}
