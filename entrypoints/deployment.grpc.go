package entrypoints

import (
	"context"
	deploymentpb "overseer/api-go/deployment/v1"
	"overseer/app"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type DeploymentServer struct {
	app *app.App
}

func NewDeploymentServer(app *app.App) deploymentpb.DeploymentServiceServer {
	return &DeploymentServer{
		app: app,
	}
}

// List implements deployment.DeploymentServiceServer.
func (d *DeploymentServer) List(ctx context.Context, req *deploymentpb.ListRequest) (*deploymentpb.ListResponse, error) {
	resp, err := d.app.ListDeployments(ctx, app.ListDeploymentsParameters{})
	if err != nil {
		return nil, err
	}

	var pbDeployments []*deploymentpb.Deployment
	for _, dep := range resp {
		pbDeployments = append(pbDeployments, &deploymentpb.Deployment{
			InstanceId: dep.InstanceId,
			Version:    dep.Version,
			DeployedAt: timestamppb.New(dep.DeployedAt),
		})
	}

	return &deploymentpb.ListResponse{
		Deployments: pbDeployments,
		Pagination: &deploymentpb.ResponsePagination{
			Total: int32(len(resp)),
		},
	}, nil
}

func (d *DeploymentServer) Register(ctx context.Context, req *deploymentpb.RegisterRequest) (*deploymentpb.RegisterResponse, error) {
	params := app.RegisterDeploymentParams{
		InstanceId: req.InstanceId,
		Version:    req.Version,
	}

	if req.DeployedAt != nil {
		params.DeployedAt = req.DeployedAt.AsTime()
	}

	if err := d.app.RegisterDeployment(ctx, params); err != nil {
		return nil, err
	}

	return &deploymentpb.RegisterResponse{}, nil
}
