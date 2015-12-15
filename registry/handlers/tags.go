package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/gorilla/handlers"
)

// tagsDispatcher constructs the tags handler api endpoint.
func tagsDispatcher(ctx *Context, r *http.Request) http.Handler {
	tagsHandler := &tagsHandler{
		Context: ctx,
	}

	return handlers.MethodHandler{
		"GET": http.HandlerFunc(tagsHandler.GetTags),
	}
}

// tagsHandler handles requests for lists of tags under a repository name.
type tagsHandler struct {
	*Context
}

type tagsAPIResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// GetTags returns a json list of tags for a specific image name.
func (th *tagsHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	tagService := th.Repository.Tags(th)

	var allTags []string
	var tagCount int
	last := ""
	var done bool
	for !done {
		tags := make([]string, 64)
		tc, err := tagService.Enumerate(th, tags, last)
		tagCount += tc
		switch err {
		case nil:
			allTags = append(allTags, tags...)
			last = allTags[tagCount-1]
			continue
		case io.EOF:
			allTags = append(allTags, tags...)
			done = true
		default:
			switch err.(type) {
			case distribution.ErrRepositoryUnknown:
				th.Errors = append(th.Errors, v2.ErrorCodeNameUnknown.WithDetail(map[string]string{"name": th.Repository.Name().String()}))
			default:
				th.Errors = append(th.Errors, errcode.ErrorCodeUnknown.WithDetail(err))
			}
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	enc := json.NewEncoder(w)
	if err := enc.Encode(tagsAPIResponse{
		Name: th.Repository.Name().String(),
		Tags: allTags[0:tagCount],
	}); err != nil {
		th.Errors = append(th.Errors, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}
}
