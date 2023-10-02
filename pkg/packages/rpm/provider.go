package rpm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/mux"

	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/repository"
)

const definition = `[%[1]s]
name=%[1]s
baseurl=%[2]s
enabled=1
gpgcheck=1
gpgkey=%[2]s/%[3]s
`

var _ packages.Provider = (*provider)(nil)

func newProvider(_ context.Context, backend string, key []byte) (packages.Provider, error) {
	return &provider{backend: backend, key: key}, nil
}

type provider struct {
	backend string
	cache   sync.Map
	key     []byte
}

func (p *provider) repo(ctx context.Context, name string) (repository.Storage[*Package, *repo], error) {
	// if v, ok := p.cache.Load(name); ok {
	// 	return v.(repository.Storage[*Package, *repo]), nil
	// }
	repo, err := repository.NewStorage[*Package, *repo](ctx, p.backend, name, &repo{}, p.key)
	if err != nil {
		return nil, err
	}
	// p.cache.Store(name, repo)
	return repo, nil
}

func (p *provider) repositoryConfig(w http.ResponseWriter, r *http.Request) {
	ctx := repository.ContextWithAuth(r.Context(), r)
	name := mux.Vars(r)["repo"]
	repo, err := p.repo(ctx, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = repo
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := strings.TrimSuffix(r.Host, "/")
	if u, p, ok := r.BasicAuth(); ok {
		host = fmt.Sprintf("%s:%s@%s", u, p, host)
	}
	url := fmt.Sprintf("%s://%s/%s", scheme, host, strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, ".repo"), "/"))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(definition, strings.NewReplacer("/", "-").Replace(name), url, RepositoryPublicKey)))
}

func (p *provider) repositoryKey(w http.ResponseWriter, r *http.Request) {
	ctx := repository.ContextWithAuth(r.Context(), r)
	name := mux.Vars(r)["repo"]
	repo, err := p.repo(ctx, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rc, err := repo.Open(ctx, RepositoryPublicKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.Copy(w, rc)
}

func (p *provider) uploadPackage(w http.ResponseWriter, r *http.Request) {
	ctx := repository.ContextWithAuth(r.Context(), r)
	name := mux.Vars(r)["repo"]

	var (
		reader io.ReadCloser
		size   int64
	)
	if file, header, err := r.FormFile("file"); err == nil {
		reader, size = file, header.Size
	} else {
		reader, size = r.Body, r.ContentLength
	}
	defer reader.Close()
	repo, err := p.repo(ctx, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg, err := NewPackage(reader, size, repo.Key())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := repo.Write(ctx, pkg); err != nil {
		if errors.Is(err, os.ErrExist) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (p *provider) downloadPackage(w http.ResponseWriter, r *http.Request) {
	ctx := repository.ContextWithAuth(r.Context(), r)
	name := mux.Vars(r)["repo"]
	file := mux.Vars(r)["filename"]
	repo, err := p.repo(ctx, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := repo.ServeFile(w, r, file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *provider) deletePackage(w http.ResponseWriter, r *http.Request) {
	ctx := repository.ContextWithAuth(r.Context(), r)
	name := mux.Vars(r)["repo"]
	file := mux.Vars(r)["filename"]
	repo, err := p.repo(ctx, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := repo.Delete(ctx, file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *provider) repositoryFile(w http.ResponseWriter, r *http.Request) {
	ctx := repository.ContextWithAuth(r.Context(), r)
	name := mux.Vars(r)["repo"]
	file := mux.Vars(r)["filename"]
	repo, err := p.repo(ctx, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := repo.ServeFile(w, r, file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *provider) Route(r *mux.Router) {
	r.HandleFunc("/{repo:.+}.repo", p.repositoryConfig).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/"+RepositoryPublicKey, p.repositoryKey).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/upload", p.uploadPackage).Methods(http.MethodPut)
	r.HandleFunc("/{repo:.+}/repodata/{filename}", p.repositoryFile).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/{filename}", p.downloadPackage).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/{filename}", p.deletePackage).Methods(http.MethodDelete)
}

func init() {
	packages.Register("rpm", newProvider)
}
