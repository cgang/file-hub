package model

type Resource struct {
	ReposName string
	ReposID   int
	OwnerID   int
	Path      string
}

func (r *Resource) String() string {
	return r.ReposName + "/" + r.Path
}
