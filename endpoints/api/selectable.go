package api

type Selectable struct {
	ID   string
	Name string
}
type Selectables []*Selectable

func (s Selectables) Len() int      { return len(s) }
func (s Selectables) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ Selectables }

func (s ByName) Less(i, j int) bool { return s.Selectables[i].Name < s.Selectables[j].Name }