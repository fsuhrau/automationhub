package models

import (
	"encoding/json"
)

type AppParameterString struct {
	Type    string `json:"type"`
	Default string `json:"default"`
}

type AppParameterOption struct {
	Type         string   `json:"type"`
	Options      []string `json:"options"`
	DefaultValue string   `json:"defaultValue"`
}

type AppParameter struct {
	Model
	AppID uint   `json:"appId"`
	App   *App   `json:"app" gorm:"foreignKey:AppID"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

// Custom UnmarshalJSON for the Type field
func (a *AppParameter) UnmarshalJSON(data []byte) error {
	type Alias AppParameter
	aux := &struct {
		Type interface{} `json:"type"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert the Type field to a string if it's an object
	switch v := aux.Type.(type) {
	case string:
		a.Type = v
	case map[string]interface{}:
		converted, err := json.Marshal(v)
		if err != nil {
			return err
		}
		a.Type = string(converted)
	default:
		return &json.UnmarshalTypeError{
			Value: string(data),
			Type:  nil,
		}
	}

	return nil
}

func (a *AppParameter) MarshalJSON() ([]byte, error) {
	type Alias AppParameter
	aux := &struct {
		Type interface{} `json:"type"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	// Convert the Type field back to its original format
	var typeValue interface{}
	if json.Valid([]byte(a.Type)) {
		// If the Type field is valid JSON, unmarshal it into a map
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(a.Type), &obj); err != nil {
			return nil, err
		}
		typeValue = obj
	} else {
		// Otherwise, treat it as a plain string
		typeValue = a.Type
	}
	aux.Type = typeValue

	return json.Marshal(aux)
}
