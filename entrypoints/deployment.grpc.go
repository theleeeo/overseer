package entrypoints

import (
	"context"
	deploymentpb "overseer/api-go/deployment/v1"
	"overseer/app"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	resp, err := d.app.ListDeployments(ctx)
	if err != nil {
		return nil, err
	}

	var pbDeployments []*deploymentpb.Deployment
	for _, dep := range resp {
		pbDeployments = append(pbDeployments, &deploymentpb.Deployment{
			EnvironmentId: int32(dep.EnvironmentId),
			ApplicationId: int32(dep.ApplicationId),
			Version:       dep.Version,
			DeployedAt:    timestamppb.New(dep.DeployedAt),
		})
	}

	return &deploymentpb.ListResponse{
		Deployments: pbDeployments,
		Pagination: &deploymentpb.ResponsePagination{
			Total: int32(len(resp)),
		},
	}, nil
}

// Register implements deployment.DeploymentServiceServer.
func (d *DeploymentServer) Register(ctx context.Context, req *deploymentpb.RegisterRequest) (*deploymentpb.RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
