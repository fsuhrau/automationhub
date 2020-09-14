package device

type Properties struct {
	Name         string
	DeviceID     string
	Type         string
	OS           string
	Architecture string
	App          string
}

func (p *Properties) AreValid() bool {
	if len(p.App) == 0 {
		return false
	}
	return true
}