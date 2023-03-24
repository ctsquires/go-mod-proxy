package index

import (
	"context"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/go-mod-proxy/go-mod-proxy/go/internal/pkg/service/storage"
)

const storageGoModObjNamePrefix = "gomod/"

type Service struct {
	storage storage.Storage
}

type ModuleIndex struct {
	Module    string
	Version   string
	Timestamp time.Time
}

const defaultLimit = 2000

func NewService(storage storage.Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) GetIndex(ctx context.Context, since time.Time, limit int) ([]ModuleIndex, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	index := make([]ModuleIndex, 0, limit)
	var pageToken string
	for {
		var objList *storage.ObjectList
		objList, err := s.storage.ListObjects(ctx, storage.ObjectListOptions{
			NamePrefix: storageGoModObjNamePrefix,
			PageToken:  pageToken,
		})
		if err != nil {
			return nil, err
		}
		for _, o := range objList.Objects {
			if o.CreatedTime.Before(since) {
				continue
			}
			o.Name = strings.TrimPrefix(o.Name, storageGoModObjNamePrefix)
			module, version, ok := strings.Cut(o.Name, "@")
			if !ok {
				continue
			}
			index = append(index, ModuleIndex{
				Module:    module,
				Version:   version,
				Timestamp: o.CreatedTime,
			})
		}
		pageToken = objList.NextPageToken
		if pageToken == "" {
			break
		}
	}

	// Chronological order means first to last
	slices.SortFunc(index, func(i, j ModuleIndex) bool {
		return i.Timestamp.Before(j.Timestamp)
	})
	if len(index) < limit {
		return index, nil
	}
	return index[:limit], nil
}
