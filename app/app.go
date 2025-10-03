package app

import (
	"context"
	"overseer/repo"
	"time"
)

type Environment struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type Application struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type Deployment struct {
	EnvironmentId int       `json:"environment_id"`
	ApplicationId int       `json:"application_id"`
	Version       string    `json:"version"`
	DeployedAt    time.Time `json:"deployed_at"`
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
			Id:    int(app.ID),
			Name:  app.Name,
			Order: int(app.SortOrder),
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
		Id:    int(app.ID),
		Name:  app.Name,
		Order: int(app.SortOrder),
	}, nil
}

func (a *App) UpdateApplication(ctx context.Context, id int, name string) (Application, error) {
	app, err := a.db.UpdateApplication(ctx, repo.UpdateApplicationParams{
		ID:   int32(id),
		Name: name,
	})
	if err != nil {
		return Application{}, err
	}
	return Application{
		Id:    int(app.ID),
		Name:  app.Name,
		Order: int(app.SortOrder),
	}, nil
}

func (a *App) DeleteApplication(ctx context.Context, id int) error {
	return a.db.DeleteApplication(ctx, int32(id))
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
			Id:    int(env.ID),
			Name:  env.Name,
			Order: int(env.SortOrder),
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
		Id:    int(env.ID),
		Name:  env.Name,
		Order: int(env.SortOrder),
	}, nil
}

func (a *App) UpdateEnvironment(ctx context.Context, id int, name string) (Environment, error) {
	env, err := a.db.UpdateEnvironment(ctx, repo.UpdateEnvironmentParams{
		ID:   int32(id),
		Name: name,
	})
	if err != nil {
		return Environment{}, err
	}
	return Environment{
		Id:    int(env.ID),
		Name:  env.Name,
		Order: int(env.SortOrder),
	}, nil
}

func (a *App) DeleteEnvironment(ctx context.Context, id int) error {
	return a.db.DeleteEnvironment(ctx, int32(id))
}

func (a *App) ReorderEnvironments(ctx context.Context, ids []int32) error {
	if err := a.db.ReorderEnvironments(ctx, ids); err != nil {
		return err
	}

	return nil
}

func (a *App) ListDeployments(ctx context.Context) ([]Deployment, error) {
	deployments, err := a.db.ListDeploymentsFlat(ctx)
	if err != nil {
		return nil, err
	}

	var result []Deployment
	for _, d := range deployments {
		result = append(result, Deployment{
			EnvironmentId: int(d.EnvironmentID),
			ApplicationId: int(d.ApplicationID),
			Version:       d.Version,
			DeployedAt:    d.DeployedAt.Time,
		})
	}

	return result, nil
}
