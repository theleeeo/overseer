package app

import (
	"context"
	"errors"
	"overseer/repo"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Environment struct {
	Id    int32  `json:"id"`
	Name  string `json:"name"`
	Order int32  `json:"order"`
}

type Application struct {
	Id    int32  `json:"id"`
	Name  string `json:"name"`
	Order int32  `json:"order"`
}

type Instance struct {
	Id            int32  `json:"id"`
	EnvironmentId int32  `json:"environment_id"`
	ApplicationId int32  `json:"application_id"`
	Name          string `json:"name"`
}

type Deployment struct {
	InstanceId int32     `json:"instance_id"`
	Version    string    `json:"version"`
	DeployedAt time.Time `json:"deployed_at"`
}

type App struct {
	db *repo.Queries
}

func New(db *repo.Queries) *App {
	return &App{db: db}
}

func (a *App) ListApplications(ctx context.Context) ([]Application, error) {
	apps, err := a.db.ListApplications(ctx)
	if err != nil {
		return nil, err
	}

	var result []Application
	for _, app := range apps {
		result = append(result, Application{
			Id:    app.ID,
			Name:  app.Name,
			Order: app.SortOrder,
		})
	}

	return result, nil
}

func (a *App) CreateApplication(ctx context.Context, name string) (Application, error) {
	app, err := a.db.CreateApplication(ctx, name)
	if err != nil {
		return Application{}, err
	}
	return Application{
		Id:    app.ID,
		Name:  app.Name,
		Order: app.SortOrder,
	}, nil
}

func (a *App) UpdateApplication(ctx context.Context, id int32, name string) (Application, error) {
	app, err := a.db.UpdateApplication(ctx, repo.UpdateApplicationParams{
		ID:   id,
		Name: name,
	})
	if err != nil {
		return Application{}, err
	}
	return Application{
		Id:    app.ID,
		Name:  app.Name,
		Order: app.SortOrder,
	}, nil
}

func (a *App) DeleteApplication(ctx context.Context, id int32) error {
	return a.db.DeleteApplication(ctx, id)
}

func (a *App) ReorderApplications(ctx context.Context, ids []int32) error {
	if err := a.db.ReorderApplications(ctx, ids); err != nil {
		return err
	}

	return nil
}

func (a *App) ListEnvironments(ctx context.Context) ([]Environment, error) {
	envs, err := a.db.ListEnvironments(ctx)
	if err != nil {
		return nil, err
	}

	var result []Environment
	for _, env := range envs {
		result = append(result, Environment{
			Id:    env.ID,
			Name:  env.Name,
			Order: env.SortOrder,
		})
	}

	return result, nil
}

func (a *App) CreateEnvironment(ctx context.Context, name string) (Environment, error) {
	env, err := a.db.CreateEnvironment(ctx, name)
	if err != nil {
		return Environment{}, err
	}
	return Environment{
		Id:    env.ID,
		Name:  env.Name,
		Order: env.SortOrder,
	}, nil
}

func (a *App) UpdateEnvironment(ctx context.Context, id int32, name string) (Environment, error) {
	env, err := a.db.UpdateEnvironment(ctx, repo.UpdateEnvironmentParams{
		ID:   id,
		Name: name,
	})
	if err != nil {
		return Environment{}, err
	}
	return Environment{
		Id:    env.ID,
		Name:  env.Name,
		Order: env.SortOrder,
	}, nil
}

func (a *App) DeleteEnvironment(ctx context.Context, id int32) error {
	return a.db.DeleteEnvironment(ctx, id)
}

func (a *App) ReorderEnvironments(ctx context.Context, ids []int32) error {
	if err := a.db.ReorderEnvironments(ctx, ids); err != nil {
		return err
	}

	return nil
}

type CreateInstanceParameters struct {
	EnvironmentId int32
	ApplicationId int32
	Name          string
}

func (a *App) CreateInstance(ctx context.Context, params CreateInstanceParameters) (int32, error) {
	if params.EnvironmentId == 0 {
		return 0, errors.New("environment id is required")
	}

	if params.ApplicationId == 0 {
		return 0, errors.New("application id is required")
	}

	if params.Name == "" {
		return 0, errors.New("name is required")
	}

	return a.db.CreateInstance(ctx, repo.CreateInstanceParams{
		EnvironmentID: params.EnvironmentId,
		ApplicationID: params.ApplicationId,
		Name:          params.Name,
	})
}

type UpdateInstanceParameters struct {
	Id   int32
	Name string
}

func (a *App) UpdateInstance(ctx context.Context, params UpdateInstanceParameters) error {
	if params.Id == 0 {
		return errors.New("instance id is required")
	}

	if params.Name == "" {
		return errors.New("name is required")
	}

	return a.db.UpdateInstance(ctx, repo.UpdateInstanceParams{
		ID:   params.Id,
		Name: params.Name,
	})
}

type ListInstancesParameters struct {
	Name string
}

func (a *App) ListInstances(ctx context.Context, params ListInstancesParameters) ([]Instance, error) {
	instances, err := a.db.ListInstances(ctx, params.Name)
	if err != nil {
		return nil, err
	}

	var result []Instance
	for _, i := range instances {
		result = append(result, Instance{
			Id:            i.ID,
			EnvironmentId: i.EnvironmentID,
			ApplicationId: i.ApplicationID,
			Name:          i.Name,
		})
	}

	return result, nil
}

type ListDeploymentsParameters struct{}

func (a *App) ListDeployments(ctx context.Context, params ListDeploymentsParameters) ([]Deployment, error) {
	deployments, err := a.db.ListDeployments(ctx)
	if err != nil {
		return nil, err
	}

	var result []Deployment
	for _, d := range deployments {
		result = append(result, Deployment{
			InstanceId: d.InstanceID,
			Version:    d.Version,
			DeployedAt: d.DeployedAt.Time,
		})
	}

	return result, nil
}

type RegisterDeploymentParams struct {
	InstanceId int32
	Version    string
	DeployedAt time.Time
}

func (a *App) RegisterDeployment(ctx context.Context, params RegisterDeploymentParams) error {
	if params.InstanceId == 0 {
		return errors.New("instance id is required")
	}

	if params.Version == "" {
		return errors.New("version is required")
	}

	if params.DeployedAt.IsZero() {
		params.DeployedAt = time.Now().UTC()
	}

	return a.db.RegisterDeployment(ctx, repo.RegisterDeploymentParams{
		ID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
		InstanceID: params.InstanceId,
		Version:    params.Version,
		DeployedAt: pgtype.Timestamptz{Time: params.DeployedAt, Valid: true},
	})
}
