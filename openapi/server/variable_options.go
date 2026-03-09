package server

import v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

// VariableOption configures a server variable.
type VariableOption func(*v3.ServerVariable)

// VariableEnum sets the allowed values for the server variable.
func VariableEnum(enum ...string) VariableOption {
	return func(sv *v3.ServerVariable) {
		sv.Enum = enum
	}
}

// VariableDescription sets the description for the server variable.
func VariableDescription(description string) VariableOption {
	return func(sv *v3.ServerVariable) {
		sv.Description = description
	}
}
