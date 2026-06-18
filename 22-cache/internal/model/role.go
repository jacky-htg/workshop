package model

type Role struct {
	ID   int
	Name string

	Accesses []Access
}
