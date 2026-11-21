package test

import (
	"strings"

	"github.com/fsuhrau/automationhub/apperr"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Edit is what the test edit page saves: the name, how and where the test
// runs, and for Unity tests what it runs. The json tags are the wire format
// of the page.
type Edit struct {
	Name                  string                       `json:"name"`
	ExecutionType         models.ExecutionType         `json:"executionType"`
	AllDevices            bool                         `json:"allDevices"`
	Devices               []uint                       `json:"devices"`
	UnityTestCategoryType models.UnityTestCategoryType `json:"unityTestCategoryType"`
	Categories            string                       `json:"categories"`
	TestFunctions         []models.UnityTestFunction   `json:"testFunctions"`
	RecordSession         bool                         `json:"recordSession"`
}

// NewTest is what the new test page sends: the test type and the same
// settings the edit page saves.
type NewTest struct {
	Type models.TestType `json:"type"`
	Edit
}

// validate tidies the edit (trimmed name and categories, no repeated
// devices or functions) and says what keeps the test from running.
func (e *Edit) validate(unity bool) error {
	e.Name = strings.TrimSpace(e.Name)
	if e.Name == "" {
		return apperr.Invalid("the test needs a name")
	}
	if e.ExecutionType != models.ConcurrentExecutionType && e.ExecutionType != models.SimultaneouslyExecutionType {
		return apperr.Invalid("unknown execution type %d", e.ExecutionType)
	}
	if err := e.tidyDevices(); err != nil {
		return err
	}
	if !unity {
		return nil
	}
	e.Categories = appsvc.NormalizeTags(e.Categories)
	e.TestFunctions = uniqueFunctions(e.TestFunctions)
	return e.validateSelection()
}

// tidyDevices drops repeated devices; all devices means no picked ones.
func (e *Edit) tidyDevices() error {
	seen := map[uint]bool{}
	devices := e.Devices[:0:0]
	for _, d := range e.Devices {
		if !seen[d] {
			seen[d] = true
			devices = append(devices, d)
		}
	}
	e.Devices = devices
	if e.AllDevices {
		e.Devices = nil
	} else if len(e.Devices) == 0 {
		return apperr.Invalid("pick at least one device, or run on all devices")
	}
	return nil
}

// validateSelection checks that the Unity test selection selects something.
func (e *Edit) validateSelection() error {
	switch e.UnityTestCategoryType {
	case models.AllTest:
	case models.AllOfCategory:
		if e.Categories == "" {
			return apperr.Invalid("name at least one category")
		}
	case models.SelectedTestsOnly:
		if len(e.TestFunctions) == 0 {
			return apperr.Invalid("pick at least one test")
		}
	default:
		return apperr.Invalid("unknown Unity test selection %d", e.UnityTestCategoryType)
	}
	return nil
}

// uniqueFunctions trims the functions and drops those without a method and
// repeated ones.
func uniqueFunctions(in []models.UnityTestFunction) []models.UnityTestFunction {
	type key struct{ assembly, class, method string }
	known := map[key]bool{}
	functions := in[:0:0]
	for _, f := range in {
		k := key{strings.TrimSpace(f.Assembly), strings.TrimSpace(f.Class), strings.TrimSpace(f.Method)}
		if k.method == "" || known[k] {
			continue
		}
		known[k] = true
		functions = append(functions, models.UnityTestFunction{Assembly: k.assembly, Class: k.class, Method: k.method})
	}
	return functions
}
