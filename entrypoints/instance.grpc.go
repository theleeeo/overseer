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

func (d *InstanceServer) Create(context.Context, *instancepb.CreateRequest) (*instancepb.CreateResponse, error) {
	panic("unimplemented")
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
