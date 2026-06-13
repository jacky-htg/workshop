package model

type Access struct {
	ID       int
	ParentID *int
	Name     string
	Alias    string
}

type AccessTree struct {
	ID        int
	Name      string
	Alias     string
	Childrens []Access
}
