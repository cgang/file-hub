package model

// Resource represents a file or directory within a repository
type Resource struct {
	Repo *Repository
	Path string
}

func (r *Resource) String() string {
	return r.Repo.Name + r.Path
}
