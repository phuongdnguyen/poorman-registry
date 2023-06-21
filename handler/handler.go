package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/xxxibgdrgnmm/reverse-registry/repository"
	containerregistry "github.com/xxxibgdrgnmm/reverse-registry/services/container-registry"
	"github.com/xxxibgdrgnmm/reverse-registry/utils"
)

type Interface interface {
	V2Handler(c *gin.Context)
	TokenHandler(c *gin.Context)
	ProxyHandler(c *gin.Context)
}

type client struct {
	containerRegistryService containerregistry.Interface
	imageStorage             repository.Interface
	log                      *logrus.Logger
}

type Options struct {
	Log     *logrus.Logger
	Cr      containerregistry.Interface
	Storage repository.Interface
}

func New(opt Options) Interface {
	return &client{log: opt.Log, containerRegistryService: opt.Cr, imageStorage: opt.Storage}
}

func (s *client) V2Handler(ctx *gin.Context) {
	s.log.WithContext(ctx)
	out, _ := http.NewRequest(ctx.Request.Method, "https://cgr.dev/v2/", nil)
	s.log.WithFields(logrus.Fields{
		"method": out.Method,
		"url":    out.URL.String(),
		"header": utils.Redact(ctx.Request.Header),
	}).Info("sending request")

	ctx.Writer.Header().Add("X-Redirected", out.URL.String())

	back, err := http.DefaultClient.Do(out)
	if err != nil {
		s.log.Errorf("error sending request: %v", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer back.Body.Close()

	s.log.WithFields(logrus.Fields{
		"method": out.Method,
		"url":    out.URL.String(),
		"status": back.Status,
		"header": utils.Redact(back.Header),
	}).Info("got response")

	for k, v := range back.Header {
		for _, vv := range v {
			ctx.Writer.Header().Add(k, vv)
		}
	}

	// Ping responses may include a response header to point to where to get a token, that looks like:
	//   Www-Authenticate: Bearer realm="http://cgr.dev/token",service="cgr.dev"
	//
	// In order for the client to be able to use this, we need to rewrite it to
	// point to our token endpoint, not the upstream:
	//   Www-Authenticate: Bearer realm="http://$HOST/token",service="cgr.dev"
	wwwAuth := back.Header.Get("Www-Authenticate")
	if wwwAuth != "" {
		rewrittenWwwAuth := strings.Replace(wwwAuth, `https://cgr.dev/`, fmt.Sprintf(`http://%s/`, ctx.Request.Host), 1)
		ctx.Writer.Header().Set("Www-Authenticate", rewrittenWwwAuth)
	}
	ctx.Writer.WriteHeader(back.StatusCode)
	if _, err := io.Copy(ctx.Writer, back.Body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, fmt.Sprintf("error copying response body %v", err))
	}

}

func (s *client) TokenHandler(ctx *gin.Context) {
	s.log.WithContext(ctx)
	vals := ctx.Request.URL.Query()
	scope := vals.Get("scope")
	scope = strings.Replace(scope, "repository:", "repository:chainguard/", 1)
	vals.Set("scope", scope)

	url := "https://cgr.dev/token?" + vals.Encode()
	out, _ := http.NewRequest(ctx.Request.Method, url, nil)
	out.Header = ctx.Request.Header.Clone()

	s.log.WithFields(logrus.Fields{
		"method": out.Method,
		"url":    out.URL.String(),
		"header": utils.Redact(ctx.Request.Header),
	}).Info("sending request")
	ctx.Header("X-Redirected", out.URL.String())

	back, err := http.DefaultClient.Do(out)
	if err != nil {
		s.log.Errorf("error sending request: %v", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer back.Body.Close()

	s.log.WithFields(logrus.Fields{
		"method": out.Method,
		"url":    out.URL.String(),
		"status": back.Status,
		"header": utils.Redact(back.Header),
	}).Info("got response")

	for k, v := range back.Header {
		for _, vv := range v {
			ctx.Header(k, vv)
		}
	}

	ctx.Status(back.StatusCode)
	if _, err := io.Copy(ctx.Writer, back.Body); err != nil {
		s.log.Errorf("Error copying response body: %v", err)
	}
}

func (s *client) ProxyHandler(ctx *gin.Context) {
	// /v2/nginx/manifests/1.25.1-r0
	a := strings.Split(ctx.Request.URL.Path, "/")
	image := a[2]
	ref := a[4]
	if reference.NameRegexp.MatchString(image) && reference.TagRegexp.MatchString(ref) {
		// nginx:1.25.1-r0
		nameWithTag := image + ":" + ref
		r, err := s.imageStorage.FindByNameTag(nameWithTag)
		if err != nil {
			s.log.Errorf("find name tag %v", err)
		}
		if r.HashedIndex != "" {
			ctx.Writer.Header().Set("Content-Type", "application/vnd.oci.image.index.v1+json")
			ctx.Writer.Header().Set("Docker-Content-Digest", r.HashedIndex)
			ctx.Writer.Header().Set("Content-Length", "0")
			ctx.Status(http.StatusOK)
			s.log.Info("sent response from local db")
			return
		}
	}
	repo := ctx.Param("repo")
	fmt.Printf("repo: %v\n", repo)
	rest := ctx.Param("rest")
	fmt.Printf("rest: %v\n", rest)
	url := fmt.Sprintf("https://cgr.dev/v2/chainguard/%s%s", repo, rest)
	if query := ctx.Request.URL.Query().Encode(); query != "" {
		url += "?" + query
	}
	out, _ := http.NewRequest(ctx.Request.Method, url, nil)
	out.Header = ctx.Request.Header.Clone()

	s.log.WithFields(logrus.Fields{
		"method": out.Method,
		"url":    out.URL.String(),
		"header": utils.Redact(out.Header),
	}).Info("sending request")
	ctx.Header("X-Redirected", out.URL.String())

	back, err := http.DefaultTransport.RoundTrip(out) // Transport doesn't follow redirects.
	if err != nil {
		s.log.Errorf("Error sending request: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	defer back.Body.Close()

	s.log.WithFields(logrus.Fields{
		"method": out.Method,
		"url":    out.URL.String(),
		"status": back.Status,
		"header": utils.Redact(back.Header),
		"body":   back.Body,
	}).Info("got response")
	// Copy response headers.
	for k, v := range back.Header {
		for _, vv := range v {
			ctx.Header(k, vv)
		}
	}

	// Responses may include a header to point to where to get a token, that looks like:
	//   Www-Authenticate: Bearer realm="http://cgr.dev/token",service="cgr.dev"
	//
	// In order for the client to be able to use this, we need to rewrite it to
	// point to our token endpoint, not the upstream:
	//   Www-Authenticate: Bearer realm="http://$HOST/token",service="cgr.dev"
	wwwAuth := back.Header.Get("Www-Authenticate")
	if wwwAuth != "" {
		rewrittenWwwAuth := strings.Replace(wwwAuth, `://cgr.dev/`, fmt.Sprintf(`://%s/`, ctx.Request.Host), 1)
		ctx.Header("Www-Authenticate", rewrittenWwwAuth)
	}

	// List responses may include a response header to support pagination, that looks like:
	//   Link: </v2/chainguard/static/tags/list?n=100&last=blah>; rel="next">
	//
	// In order for the client to be able to use this link, we need to rewrite it to
	// point to the user's requested repo, not the upstream:
	//   Link: </v2/static/repo/tags/list?n=100&last=blah>; rel="next">
	link := back.Header.Get("Link")
	if link != "" {
		rewrittenLink := strings.Replace(link, "/v2/chainguard/", "/v2/", 1)
		ctx.Header("Link", rewrittenLink)
	}

	// If it's a list request, rewrite the response so the name key matches the
	// user's requested repo, otherwise clients will repeatedly request the
	// first page looking for their repo's tags.
	if strings.Contains(ctx.Request.URL.Path, "/tags/list") {
		var lr listResponse
		if err := json.NewDecoder(back.Body).Decode(&lr); err != nil {
			s.log.Errorf("Error decoding list response body: %v", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		// chainguard/nginx -> nginx
		lr.Name = strings.TrimPrefix(lr.Name, "chainguard/")

		// Unset the content-length header from our response, because we're
		// about to rewrite the response to be shorter than the original.
		// This can confuse Cloud Run, which responds with an empty body
		// if the content-length header is wrong in some cases.
		ctx.Header("Content-Length", "")
		ctx.Status(back.StatusCode)
		if err := json.NewEncoder(ctx.Writer).Encode(lr); err != nil {
			s.log.Errorf("Error encoding list response body: %v", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		}

		return
	} else {
		ctx.Status(back.StatusCode)
	}

	// Copy response body.
	if _, err := io.Copy(ctx.Writer, back.Body); err != nil {
		s.log.Errorf("Error copying response body: %v", err)
	}

}

type listResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
