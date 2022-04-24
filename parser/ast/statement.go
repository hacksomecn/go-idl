package ast

type Pos struct {
	Package  string `json:"package"`
	FileName string `json:"file_name"`
	Name     string `json:"name"`
	FilePath string `json:"file_path"`
}

// Statement definition statement
type Statement struct {
	Expr string // expr string
	Pos  *Pos   // statement pos
}

type IStatement interface {
}

type Assignment struct {
	Statement
}

type Comment struct {
	Statement
}

type Import struct {
	Statement
}

type Decorator struct {
	Statement
}

type Model struct {
	Statement
}

type Rest struct {
	Statement
}

type Grpc struct {
	Statement
}

type Ws struct {
	Statement
}

type Raw struct {
	Statement
}
