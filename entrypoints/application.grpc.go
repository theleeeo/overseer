package entrypoints

import (
	"context"

	applicationpb "overseer/api-go/application/v1"
	"overseer/app"
)

type ApplicationServer struct {
	app *app.App
}

func NewApplicationServer(app *app.App) applicationpb.ApplicationServiceServer {
	return &ApplicationServer{
		app: app,
	}
}

func (e *ApplicationServer) Create(ctx context.Context, req *applicationpb.CreateRequest) (*applicationpb.CreateResponse, error) {
	resp, err := e.app.CreateApplication(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	return &applicationpb.CreateResponse{
		Id: int32(resp.Id),
	}, nil
}

func (e *ApplicationServer) List(ctx context.Context, req *applicationpb.ListRequest) (*applicationpb.ListResponse, error) {
	apps, err := e.app.ListApplications(ctx)
	if err != nil {
		return nil, err
	}

	var pbApps []*applicationpb.Application
	for _, app := range apps {
		pbApps = append(pbApps, &applicationpb.Application{
			Id:        int32(app.Id),
			Name:      app.Name,
			SortOrder: int32(app.Order),
		})
	}

	return &applicationpb.ListResponse{
		Applications: pbApps,
		Pagination: &applicationpb.ResponsePagination{
			Total: int32(len(apps)),
		},
	}, nil
}

func (e *ApplicationServer) SetSortOrder(ctx context.Context, req *applicationpb.SetSortOrderRequest) (*applicationpb.SetSortOrderResponse, error) {
	var ids []int32
	for _, id := range req.IdsInOrder {
		ids = append(ids, int32(id))
	}

	if err := e.app.ReorderApplications(ctx, ids); err != nil {
		return nil, err
	}

	return &applicationpb.SetSortOrderResponse{}, nil
}

func (e *ApplicationServer) Update(ctx context.Context, req *applicationpb.UpdateRequest) (*applicationpb.UpdateResponse, error) {
	// Since there right now only exists the name field to update, we can just check if it's nil
	// and return early if it is. In the future, if more fields are added, this should be changed to
	// handle multiple fields properly.
	if req.Name == nil {
		return &applicationpb.UpdateResponse{}, nil
	}

	_, err := e.app.UpdateApplication(ctx, int(req.Id), *req.Name)
	if err != nil {
		return nil, err
	}

	return &applicationpb.UpdateResponse{}, nil
}
