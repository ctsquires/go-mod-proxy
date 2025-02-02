package goindex

import (
	"context"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	"github.com/go-mod-proxy/go-mod-proxy/go/internal/pkg/service/gomodule/gocmd"
	"github.com/go-mod-proxy/go-mod-proxy/go/internal/pkg/service/storage"
)

type Service struct {
	storage storage.Storage
}

type IndexItem struct {
	Path      string
	Version   string
	Timestamp time.Time
}

const defaultLimit = 2000

func NewService(storage storage.Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) GetIndex(ctx context.Context, since time.Time, limit int) ([]IndexItem, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	log.Infof("Retrieving index items: since=%+v, limit=%d", since, limit)
	index := make([]IndexItem, 0, limit)
	var pageToken string
	for {
		objList, err := s.storage.ListObjects(ctx, storage.ObjectListOptions{
			NamePrefix: gocmd.StorageGoModObjNamePrefix + "github.com/BurntSushi/",
			PageToken:  pageToken,
		})
		if err != nil {
			return nil, err
		}
		for _, o := range objList.Objects {
			if o.CreatedTime.Before(since) {
				continue
			}
			o.Name = strings.TrimPrefix(o.Name, gocmd.StorageGoModObjNamePrefix)
			path, version, ok := strings.Cut(o.Name, "@")
			if !ok {
				continue
			}
			index = append(index, IndexItem{
				Path:      path,
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
	slices.SortFunc(index, func(i, j IndexItem) bool {
		return i.Timestamp.Before(j.Timestamp)
	})
	if len(index) < limit {
		return index, nil
	}
	return index[:limit], nil
}
