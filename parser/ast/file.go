package ast

type IdlFile struct {
	Pos *Pos `json:"pos"` // file position

	Assigns map[*Pos]*Assignment `json:"assigns"` // idl property assignment
	Import  map[*Pos]*Import     `json:"imports"`
	Models  map[*Pos]*Model      `json:"models"`
	Rests   map[*Pos]*Rest       `json:"rests"`
	Grpcs   map[*Pos]*Grpc       `json:"grpcs"`
	Wss     map[*Pos]*Ws         `json:"wss"`
	Raws    map[*Pos]*Raw        `json:"raws"`

	Stmts []IStatement `json:"stmts"` // all statement in sequence
}

func NewIdlFile() (file *IdlFile) {
	return &IdlFile{
		Assigns: map[*Pos]*Assignment{},
		Import:  map[*Pos]*Import{},
		Models:  map[*Pos]*Model{},
		Rests:   map[*Pos]*Rest{},
		Grpcs:   map[*Pos]*Grpc{},
		Wss:     map[*Pos]*Ws{},
		Raws:    map[*Pos]*Raw{},
		Stmts:   []IStatement{},
	}
}
