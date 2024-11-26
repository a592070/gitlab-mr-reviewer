package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gitlab-mr-reviewer/pkg/logging"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
)

func runMockGitlabServer(logger *logging.ZaprLogger, authorization string) *httptest.Server {
	serveMux := http.NewServeMux()
	logMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.
				WithValues("url", r.URL).
				WithValues("method", r.Method).
				WithValues("contentType", r.Header.Get("Content-Type")).
				Info("[MockGitlabServer]Received request.")
			next.ServeHTTP(w, r)
		})
	}
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("PRIVATE-TOKEN")
			if len(token) == 0 || token != authorization {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	serveMux.Handle("/health", logMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	serveMux.Handle("/auth/health", logMiddleware(authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))))
	serveMux.Handle("GET /api/v4/projects/{projectId}/merge_requests/{mergeRequestId}", logMiddleware(authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectId, err := strconv.Atoi(r.PathValue("projectId"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mergeRequestId, err := strconv.Atoi(r.PathValue("mergeRequestId"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		bytes, err := json.Marshal(map[string]any{
			"id":          mergeRequestId,
			"iid":         mergeRequestId,
			"project_id":  projectId,
			"title":       "feat: This is a mock merge request",
			"description": "",
			"diff_refs": map[string]string{
				"base_sha":  "63c97907ceb628e4fbbc125ca9ce5bd2a1f9566f",
				"head_sha":  "94e7e0bb7144018e544743e1d6f22731f8ddeba1",
				"start_sha": "63c97907ceb628e4fbbc125ca9ce5bd2a1f9566f",
			},
		})
		if err != nil {
			logger.Info(fmt.Sprintf("[MockGitlabServer]Error marshalling response: %s", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if _, err := w.Write(bytes); err != nil {
			logger.Info(fmt.Sprintf("[MockGitlabServer]Error writting response: %s", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}))))
	serveMux.Handle("GET /api/v4/projects/{projectId}/merge_requests/{mergeRequestId}/diffs", logMiddleware(authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := strconv.Atoi(r.PathValue("projectId"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = strconv.Atoi(r.PathValue("mergeRequestId"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		bytes, err := json.Marshal([]map[string]any{
			{
				"diff":         "@@ -0,0 +1,9 @@\n+package mutation\n+\n+import \"github.com/pkg/errors\"\n+\n+var (\n+\tErrorFailedCreateConfigmap = errors.New(\"Failed to create configmap\")\n+\tErrorConfigmapExists       = errors.New(\"Configmap already exists\")\n+\tErrorConfigmapNotFound     = errors.New(\"Configmap not found\")\n+)\n",
				"new_path":     "internal/mutation/errors.go",
				"old_path":     "internal/mutation/errors.go",
				"a_mode":       "0",
				"b_mode":       "100644",
				"new_file":     true,
				"renamed_file": false,
				"deleted_file": false,
			},
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(bytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))))

	serveMux.Handle("POST /api/v4/projects/{projectId}/merge_requests/{mergeRequestId}/notes", logMiddleware(authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectId, err := strconv.Atoi(r.PathValue("projectId"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		mergeRequestId, err := strconv.Atoi(r.PathValue("mergeRequestId"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		requestBodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		logger.Info(fmt.Sprintf("[MockGitlabServer]RequestBody: %s", string(requestBodyBytes)))

		responseBody, err := json.Marshal(map[string]any{
			"body":         string(requestBodyBytes),
			"noteable_iid": mergeRequestId,
			"project_id":   projectId,
		})
		if err != nil {
			logger.Info(fmt.Sprintf("[MockGitlabServer]Error marshalling response: %s", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write(responseBody); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

	}))))

	return httptest.NewServer(serveMux)
}

var _ = ginkgo.Describe("GitlabRepository", ginkgo.Ordered, func() {
	var logger *logging.ZaprLogger
	var r GitlabRepository
	var testServer *httptest.Server
	authorization := "fake-token"
	ginkgo.BeforeAll(func() {
		var err error
		logger, err = logging.NewZaprLogger(false, "info")
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		testServer = runMockGitlabServer(logger, authorization)
		logger.Info(fmt.Sprintf("testServer: %s", testServer.URL))
		r = NewGitlabRepository(logger, testServer.URL, authorization)
	})

	ginkgo.It("Mock Gitlab Server should be running", func() {
		var path string

		path = "/health"
		ginkgo.By(path)
		response, err := http.Get(fmt.Sprintf("%s%s", testServer.URL, path))
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		gomega.Expect(response.StatusCode).To(gomega.Equal(http.StatusOK))

		path = "/auth/health"
		ginkgo.By(path)
		request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", testServer.URL, path), nil)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		request.Header.Set("PRIVATE-TOKEN", authorization)
		response, err = http.DefaultClient.Do(request)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		gomega.Expect(response.StatusCode).To(gomega.Equal(http.StatusOK))

	})
	ginkgo.It("Should be able to GetMergeRequest", func() {
		ctx := context.Background()
		projectId := int32(1)
		mergeRequestId := int32(1)

		ginkgo.By("send request to server")
		mergeRequest, err := r.GetMergeRequest(ctx, projectId, mergeRequestId)

		ginkgo.By("should not fail")
		if err != nil {
			logger.Error(err, "GetMergeRequest failed")
		}
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		gomega.Expect(mergeRequest).ToNot(gomega.BeNil())

		ginkgo.By("validate merge request data")
		gomega.Expect(mergeRequest.Id).To(gomega.Equal(mergeRequestId))
		gomega.Expect(mergeRequest.ProjectId).To(gomega.Equal(projectId))

	})

	ginkgo.It("Should be able to CreateMergeRequestSummary", func() {
		ctx := context.Background()
		projectId := int32(1)
		mergeRequestId := int32(1)

		ginkgo.By("send request to server")
		err := r.CreateMergeRequestSummary(ctx, CreateMergeRequestSummaryInput{
			ProjectId:          projectId,
			MergeRequestId:     mergeRequestId,
			RelativeChangeNote: "## High-level Summary\nThis merge request introduces the `ConfigMapRepository` to streamline operations with Kubernetes' API and refactors the sidecar mutator logic within the Beyla project. These changes improve error handling, modularity, and maintainability while facilitating better interaction with ConfigMaps. Tests are updated accordingly to reflect the new structure.\n\n| Files/Groups                              | Summary                                                                                       |\n|-------------------------------------------|-----------------------------------------------------------------------------------------------|\n| `configmap_repository.go`, `configmap_repository_impl.go`, `errors.go` | Added `ConfigMapRepository` interface and implementation for CRUD operations with Kubernetes ConfigMaps, enhancing logging and error handling. Introduced common error variables for consistent error management. |\n| `beyla_sidecar_mutator.go`, `constants.go` | Refactored sidecar mutator to utilize the new `ConfigMapRepository`, improving separation of concerns and consistency with operations. Replaced hardcoded strings with constants. |\n| `configmap_mutator.go`                    | Removed due to redundancy after implementing the new `ConfigMapRepository`.                   |\n| `pod_webhook_handler.go`, `pod_webhook_handler_test.go` | Adjusted to utilize the refactored sidecar mutator and updated test cases to align with new logic. Updated error handling to give better HTTP responses. |\n| `injector.go`                             | Updated dependencies to inject the new `ConfigMapRepository` into components, in accordance with the refactor. |",
			SummaryNote:        "### Release Notes\n\n**New Feature:**\n- Introduced a `ConfigMapRepository` interface for streamlined operations with Kubernetes' API.\n\n**Refactor:**\n- Refactored Beyla sidecar mutator to enhance error handling and maintainability.\n- Removed `configmap_mutator` due to the introduction of a new repository.\n- Updated dependency injection to incorporate the refactor.\n\n**Test:**\n- Updated test cases to align with new repository-based logic.\n\n> In the realm of pods, new code ignites,  \n> ConfigMaps now soar to greater heights.  \n> Repositories built, with errors in check,  \n> Our Kubernetes sidecars are perfectly decked.  \n> 🚀 With refactor's flair and a feature in stride,  \n> Off to the cluster, our changes glide! 🎉\n",
		})

		ginkgo.By("should not fail")
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

	})

	ginkgo.AfterAll(func() {
		testServer.Close()
	})
})
