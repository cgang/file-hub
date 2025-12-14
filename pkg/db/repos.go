package db

import (
	"context"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/uptrace/bun"
)

type ReposModel struct {
	bun.BaseModel `bun:"table:repositories"`
	*model.Repository
}

func wrapRepos(mo *model.Repository) *ReposModel {
	return &ReposModel{Repository: mo}
}

func unwrapReposes(mos []*ReposModel) []*model.Repository {
	repos := make([]*model.Repository, len(mos))
	for i, mo := range mos {
		repos[i] = mo.Repository
	}
	return repos
}

func newRepos(id int) *ReposModel {
	return &ReposModel{Repository: &model.Repository{ID: id}}
}

func CreateRepository(ctx context.Context, mo *model.Repository) error {
	_, err := db.NewInsert().Model(wrapRepos(mo)).Exec(ctx)
	return err
}

func GetRepositoryByID(ctx context.Context, id int) (*model.Repository, error) {
	mo := newRepos(id)
	err := db.NewSelect().Model(mo).WherePK().Scan(ctx)
	if err != nil {
		return nil, err
	}
	return mo.Repository, nil
}

func GetRepositoryByName(ctx context.Context, name string) (*model.Repository, error) {
	mo := &ReposModel{}
	err := db.NewSelect().Model(mo).Where("name = ?", name).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return mo.Repository, nil
}

func ListRepositories(ctx context.Context, userID int) ([]*model.Repository, error) {
	var mos []*ReposModel
	err := db.NewSelect().Model(&mos).Where("owner_id = ?", userID).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return unwrapReposes(mos), nil
}
