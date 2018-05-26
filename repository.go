package main

import (
	"context"
	"database/sql"
	"github.com/agxp/cloudflix/video-hosting-svc/proto"
	"github.com/minio/minio-go"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type Repository interface {
	Search(p opentracing.SpanContext, query string, page uint64) ([]*video_host.GetVideoInfoResponse, error)
}

type SearchRepository struct {
	s3     *minio.Client
	pg     *sql.DB
	tracer *opentracing.Tracer
	vh     video_host.HostClient
}

func (repo *SearchRepository) Search(parent opentracing.SpanContext, query string, page uint64) ([]*video_host.GetVideoInfoResponse, error) {
	sp, _ := opentracing.StartSpanFromContext(context.Background(), "Search_Repo", opentracing.ChildOf(parent))
	defer sp.Finish()

	dbSP, _ := opentracing.StartSpanFromContext(context.Background(), "PG_Search", opentracing.ChildOf(sp.Context()))
	defer dbSP.Finish()

	getSearchQuery := `select id from videos where uploaded=true and title like $1 or description like $1 limit 20 offset $2`

	rows, err := repo.pg.Query(getSearchQuery, "%" + query + "%", page*20)
	if err != nil {
		logger.Error(err.Error())
		dbSP.Finish()
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			logger.Error(err.Error())
			dbSP.Finish()
			return nil, err
		}
		logger.Info("search result id", zap.String("id", id))
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		logger.Error(err.Error())
		dbSP.Finish()
		return nil, err
	}

	var data []*video_host.GetVideoInfoResponse

	for _, v := range ids {
		res, err := repo.vh.GetVideoInfo(context.Background(), &video_host.GetVideoInfoRequest{
			Id: v,
		})

		if err != nil {
			logger.Error(err.Error())
			dbSP.Finish()
			return nil, err
		}
		data = append(data, res)
	}

	dbSP.Finish()

	return data, nil
}