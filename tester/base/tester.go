package base

import (
	"fmt"
	"github.com/gofrs/uuid"
)

type TestRunner struct {

}

func (tr *TestRunner) NewSessionID() string {
	u, _ := uuid.NewV4()
	return fmt.Sprintf("%s", u)
}

