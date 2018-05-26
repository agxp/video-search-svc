package main

import (
	pb "github.com/agxp/cloudflix/video-search-svc/proto"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type service struct {
	repo   Repository
	tracer *opentracing.Tracer
	logger *zap.Logger
}

func (srv *service) Search(ctx context.Context, req *pb.Request, res *pb.SearchResponse) error {
	sp, _ := opentracing.StartSpanFromContext(ctx, "Search_Service")

	logger.Info("Request for Search_Service received")
	defer sp.Finish()

	rsp, err := srv.repo.Search(sp.Context(), req.Query, req.Page)
	if err != nil {
		logger.Error("failed Search", zap.Error(err))
		return err
	}

	res.Data = rsp
	return nil
}
