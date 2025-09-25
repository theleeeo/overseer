package entrypoints

import (
	"encoding/json"
	"net/http"
	"overseer/app"
	"strconv"
)

// TODO: Error handling
func RegisterRestHandlers(mux *http.ServeMux, a *app.App) {
	mux.HandleFunc("/applications", func(w http.ResponseWriter, r *http.Request) {
		apps, err := a.ListApplications(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(apps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonData)
	})

	mux.HandleFunc("POST /applications", func(w http.ResponseWriter, r *http.Request) {
		var newApp app.Application
		if err := json.NewDecoder(r.Body).Decode(&newApp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		createdApp, err := a.CreateApplication(r.Context(), newApp.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(createdApp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonData)
	})

	mux.HandleFunc("PUT /applications/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var updatedApp app.Application
		if err := json.NewDecoder(r.Body).Decode(&updatedApp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updatedAppResult, err := a.UpdateApplication(r.Context(), id, updatedApp.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(updatedAppResult)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonData)
	})

	mux.HandleFunc("POST /applications/reorder", func(w http.ResponseWriter, r *http.Request) {
		var newOrder []int32
		if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := a.ReorderApplications(r.Context(), newOrder); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("DELETE /applications/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := a.DeleteApplication(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /environments", func(w http.ResponseWriter, r *http.Request) {
		envs, err := a.ListEnvironments(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(envs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonData)
	})

	mux.HandleFunc("POST /environments", func(w http.ResponseWriter, r *http.Request) {
		var newEnv app.Environment
		if err := json.NewDecoder(r.Body).Decode(&newEnv); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		createdEnv, err := a.CreateEnvironment(r.Context(), newEnv.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(createdEnv)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonData)
	})

	mux.HandleFunc("PUT /environments/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var updatedEnv app.Environment
		if err := json.NewDecoder(r.Body).Decode(&updatedEnv); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updatedEnvResult, err := a.UpdateEnvironment(r.Context(), id, updatedEnv.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(updatedEnvResult)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonData)
	})

	mux.HandleFunc("POST /environments/reorder", func(w http.ResponseWriter, r *http.Request) {
		var newOrder []int32
		if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := a.ReorderEnvironments(r.Context(), newOrder); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("DELETE /environments/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := a.DeleteEnvironment(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	// Should return a json where the root keys are environment names and the environment-objects have their keys being application names and their values being the corresponding AppInstance objects.
	mux.HandleFunc("GET /deployments", func(w http.ResponseWriter, r *http.Request) {
		deployments, err := a.ListDeployments(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(deployments)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonData)
	})
}
