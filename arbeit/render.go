package arbeit

import "encoding/json"

func (a Arbeitsjahre) Render(indent bool) ([]byte, error) {
	return json.MarshalIndent(a, "", "    ")
}

func (a *Arbeitsjahr) Render(indent bool) ([]byte, error) {
	return json.MarshalIndent(a, "", "    ")
}

func (m *Arbeitsmonat) Render(indent bool) ([]byte, error) {
	if m == nil {
		return []byte("<empty monat>"), nil
	}
	return json.MarshalIndent(m, "", "    ")
}

// func (m Arbeitsmonat) Marshal(indent bool) ([]byte, error) {
// 	return json.MarshalIndent(m, "", "    ")
// }

func (m *Arbeitstag) Render(indent bool) ([]byte, error) {
	if m == nil {
		return []byte("<empty tag>"), nil
	}
	return json.MarshalIndent(m, "", "    ")
}
