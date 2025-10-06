package entrypoints

import (
	"context"

	environmentpb "overseer/api-go/environment/v1"
	"overseer/app"
)

type EnvironmentServer struct {
	app *app.App
}

func NewEnvironmentServer(app *app.App) environmentpb.EnvironmentServiceServer {
	return &EnvironmentServer{
		app: app,
	}
}

func (e *EnvironmentServer) Create(ctx context.Context, req *environmentpb.CreateRequest) (*environmentpb.CreateResponse, error) {
	resp, err := e.app.CreateEnvironment(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	return &environmentpb.CreateResponse{
		Id: int32(resp.Id),
	}, nil
}

func (e *EnvironmentServer) List(ctx context.Context, req *environmentpb.ListRequest) (*environmentpb.ListResponse, error) {
	envs, err := e.app.ListEnvironments(ctx)
	if err != nil {
		return nil, err
	}

	var pbEnvs []*environmentpb.Environment
	for _, env := range envs {
		pbEnvs = append(pbEnvs, &environmentpb.Environment{
			Id:        int32(env.Id),
			Name:      env.Name,
			SortOrder: int32(env.Order),
		})
	}

	return &environmentpb.ListResponse{
		Environments: pbEnvs,
		Pagination: &environmentpb.ResponsePagination{
			Total: int32(len(envs)),
		},
	}, nil
}

func (e *EnvironmentServer) SetSortOrder(ctx context.Context, req *environmentpb.SetSortOrderRequest) (*environmentpb.SetSortOrderResponse, error) {
	var ids []int32
	for _, id := range req.IdsInOrder {
		ids = append(ids, int32(id))
	}

	if err := e.app.ReorderEnvironments(ctx, ids); err != nil {
		return nil, err
	}

	return &environmentpb.SetSortOrderResponse{}, nil
}

func (e *EnvironmentServer) Update(ctx context.Context, req *environmentpb.UpdateRequest) (*environmentpb.UpdateResponse, error) {
	// Since there right now only exists the name field to update, we can just check if it's nil
	// and return early if it is. In the future, if more fields are added, this should be changed to
	// handle multiple fields properly.
	if req.Name == nil {
		return &environmentpb.UpdateResponse{}, nil
	}

	_, err := e.app.UpdateEnvironment(ctx, req.Id, *req.Name)
	if err != nil {
		return nil, err
	}

	return &environmentpb.UpdateResponse{}, nil
}
