package db

import (
	"context"
	"slices"
	"strings"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/uptrace/bun"
)

// ShareModel represents a share object for database operations
type ShareModel struct {
	bun.BaseModel `bun:"table:shares"`
	*model.Share  `bun:",inherit"`
}

func wrapShare(mo *model.Share) *ShareModel {
	return &ShareModel{Share: mo}
}

func unwrapShares(mos []*ShareModel) []*model.Share {
	shares := make([]*model.Share, len(mos))
	for i, mo := range mos {
		shares[i] = mo.Share
	}
	return shares
}

func newShare(id int) *ShareModel {
	return &ShareModel{Share: &model.Share{ID: id}}
}

func CreateShare(ctx context.Context, mo *model.Share) error {
	_, err := db.NewInsert().Model(wrapShare(mo)).Exec(ctx)
	return err
}

func GetShareByID(ctx context.Context, id int) (*model.Share, error) {
	mo := newShare(id)
	err := db.NewSelect().Model(mo).WherePK().Scan(ctx)
	if err != nil {
		return nil, err
	}
	return mo.Share, nil
}

func GetShareForObject(ctx context.Context, userID int, res *model.Resource) (*model.Share, error) {
	var mos []*ShareModel
	err := db.NewSelect().Model(&mos).Where("user_id = ? AND repos_id = ?", userID, res.ReposID).Scan(ctx)
	if err != nil {
		return nil, err
	}

	// Sort shares by path length in descending order to match the most specific share first
	slices.SortFunc(mos, func(a, b *ShareModel) int {
		return len(b.Path) - len(a.Path)
	})

	for _, mo := range mos {
		if strings.HasPrefix(res.Path, mo.Path) {
			return mo.Share, nil
		}
	}

	return nil, nil
}

func GetSharesByUserID(ctx context.Context, userID int) ([]*model.Share, error) {
	var mos []*ShareModel
	err := db.NewSelect().Model(&mos).Where("user_id = ?", userID).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return unwrapShares(mos), nil
}

func DeleteShareByID(ctx context.Context, id int) error {
	mo := newShare(id)
	_, err := db.NewDelete().Model(mo).WherePK().Exec(ctx)
	return err
}
