package entrypoints

import (
	"context"
	instancepb "overseer/api-go/instance/v1"
	"overseer/app"
)

type InstanceServer struct {
	app *app.App
}

func NewInstanceServer(app *app.App) instancepb.InstanceServiceServer {
	return &InstanceServer{
		app: app,
	}
}

func (d *InstanceServer) Create(ctx context.Context, req *instancepb.CreateRequest) (*instancepb.CreateResponse, error) {
	resp, err := d.app.CreateInstance(ctx, app.CreateInstanceParameters{
		Name:          req.Name,
		EnvironmentId: req.EnvironmentId,
		ApplicationId: req.ApplicationId,
	})
	if err != nil {
		return nil, err
	}

	return &instancepb.CreateResponse{
		Id: resp,
	}, nil
}

func (d *InstanceServer) Update(ctx context.Context, req *instancepb.UpdateRequest) (*instancepb.UpdateResponse, error) {
	if err := d.app.UpdateInstance(ctx, app.UpdateInstanceParameters{
		Id:   req.Id,
		Name: req.Name,
	}); err != nil {
		return nil, err
	}

	return &instancepb.UpdateResponse{}, nil
}

func (d *InstanceServer) List(ctx context.Context, req *instancepb.ListRequest) (*instancepb.ListResponse, error) {
	resp, err := d.app.ListInstances(ctx, app.ListInstancesParameters{
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}

	var pbInstances []*instancepb.Instance
	for _, dep := range resp {
		pbInstances = append(pbInstances, &instancepb.Instance{
			Id:            int32(dep.Id),
			EnvironmentId: int32(dep.EnvironmentId),
			ApplicationId: int32(dep.ApplicationId),
			Name:          dep.Name,
		})
	}

	return &instancepb.ListResponse{
		Instances: pbInstances,
		Pagination: &instancepb.ResponsePagination{
			Total: int32(len(resp)),
		},
	}, nil
}

func (d *InstanceServer) Delete(ctx context.Context, req *instancepb.DeleteRequest) (*instancepb.DeleteResponse, error) {
	if err := d.app.DeleteInstance(ctx, req.Id); err != nil {
		return nil, err
	}

	return &instancepb.DeleteResponse{}, nil
}
