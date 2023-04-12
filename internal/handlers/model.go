package handlers

type Method struct {
	PathPart   string `json:"path_part" validate:"required"`
	MethodType string `json:"method_type" validate:"required"`
}

type Specification struct {
	Socket  string   `json:"socket" validate:"required,url"`
	Methods []Method `json:"methods" validate:"required"`
}
