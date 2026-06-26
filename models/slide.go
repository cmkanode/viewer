package models

type Slide struct {
	ID        int
	Name      string
	Image     []byte
	Completed bool
	Tags      []string
}
