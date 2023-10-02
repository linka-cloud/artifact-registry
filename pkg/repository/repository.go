package repository

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"path/filepath"

	"github.com/opencontainers/go-digest"
)

type Artifact interface {
	io.Reader
	Name() string
	Version() string
	Path() string
	Size() int64
	Digest() digest.Digest
}

type ArtifactInfo interface {
	Name() string
	Version() string
	Path() string
	Size() int64
	Digest() digest.Digest
	Meta() []byte
}

type Repository[T Artifact] interface {
	Index(ctx context.Context, priv string, artifacts ...T) ([]Artifact, error)
	GenerateKeypair() (string, string, error)
	KeyNames() (string, string)
	Name() string
}

type Storage[T Artifact, U Repository[T]] interface {
	Stat(ctx context.Context, file string) (ArtifactInfo, error)
	Open(ctx context.Context, name string) (io.ReadCloser, error)
	Write(ctx context.Context, a T) error
	Delete(ctx context.Context, name string) error
	ServeFile(w http.ResponseWriter, r *http.Request, name string) error
	Key() string
	Close() error
}

var _ Artifact = (*File)(nil)

type File struct {
	name string
	data []byte
	r    io.Reader
}

func NewFile(name string, data []byte) *File {
	return &File{
		name: name,
		data: data,
		r:    bytes.NewReader(data),
	}
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.r.Read(p)
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Version() string {
	return ""
}

func (f *File) Path() string {
	return f.name
}

func (f *File) Size() int64 {
	return int64(len(f.data))
}

func (f *File) Digest() digest.Digest {
	return digest.FromBytes(f.data)
}

type info struct {
	version string
	path    string
	size    int64
	digest  digest.Digest
	meta    []byte
}

func (i *info) Name() string {
	return filepath.Base(i.path)
}

func (i *info) Version() string {
	return i.version
}

func (i *info) Path() string {
	return i.path
}

func (i *info) Size() int64 {
	return i.size
}

func (i *info) Digest() digest.Digest {
	return i.digest
}

func (i *info) Meta() []byte {
	return nil
}
