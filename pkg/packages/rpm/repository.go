// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package rpm

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"io"
	"time"

	"go.linka.cloud/artifact-registry/pkg/buffer"
	"go.linka.cloud/artifact-registry/pkg/codec"
	"go.linka.cloud/artifact-registry/pkg/crypt/openpgp"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

type RepoChecksum struct {
	Value string `xml:",chardata"`
	Type  string `xml:"type,attr"`
}

type RepoLocation struct {
	Href string `xml:"href,attr"`
}

type RepoData struct {
	Type         string       `xml:"type,attr"`
	Checksum     RepoChecksum `xml:"checksum"`
	OpenChecksum RepoChecksum `xml:"open-checksum"`
	Location     RepoLocation `xml:"location"`
	Timestamp    int64        `xml:"timestamp"`
	Size         int64        `xml:"size"`
	OpenSize     int64        `xml:"open-size"`
}

type Repomd struct {
	XMLName  xml.Name    `xml:"repomd"`
	Xmlns    string      `xml:"xmlns,attr"`
	XmlnsRpm string      `xml:"xmlns:rpm,attr"`
	Data     []*RepoData `xml:"data"`
}

const (
	RepositoryPublicKey  = "repository.key"
	RepositoryPrivateKey = "private.key"
)

var _ storage.Repository = (*repo)(nil)

type repo struct{}

func (r *repo) Name() string {
	return "rpm"
}

func (r *repo) GenerateKeypair() (private string, public string, err error) {
	return openpgp.GenerateKeypair("Artifact Registry", "RPM Registry", "")
}

func (r *repo) KeyNames() (string, string) {
	return RepositoryPrivateKey, RepositoryPublicKey
}

func (r *repo) Codec() storage.Codec {
	return codec.Funcs[storage.Artifact]{
		Format: "json",
		EncodeFunc: func(v storage.Artifact) ([]byte, error) {
			return json.Marshal(v)
		},
		DecodeFunc: func(b []byte) (storage.Artifact, error) {
			var v Package
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return &v, nil
		},
	}
}

func (r *repo) Index(ctx context.Context, key string, a ...storage.Artifact) ([]storage.Artifact, error) {
	packages := storage.MustAs[*Package](a)
	primary, primaryFile, err := buildPrimary(ctx, packages...)
	if err != nil {
		return nil, err
	}
	filelists, filelistsFile, err := buildFilelists(ctx, packages...)
	if err != nil {
		return nil, err
	}
	other, otherFile, err := buildOther(ctx, packages...)
	if err != nil {
		return nil, err
	}
	files, err := buildRepomd(ctx, key, primary, filelists, other)
	if err != nil {
		return nil, err
	}
	return append(files, primaryFile, filelistsFile, otherFile), nil
}

// https://docs.pulpproject.org/en/2.19/plugins/pulp_rpm/tech-reference/rpm.html#repomd-xml
func buildRepomd(_ context.Context, priv string, data ...*RepoData) ([]storage.Artifact, error) {
	var repomdContent bytes.Buffer
	repomdContent.WriteString(xml.Header)
	if err := encode(&repomdContent, &Repomd{
		Xmlns:    "http://linux.duke.edu/metadata/repo",
		XmlnsRpm: "http://linux.duke.edu/metadata/rpm",
		Data:     data,
	}); err != nil {
		return nil, err
	}

	repomdAscContent := &bytes.Buffer{}
	if err := openpgp.ArmoredDetachSign(repomdAscContent, priv, bytes.NewReader(repomdContent.Bytes())); err != nil {
		return nil, err
	}

	return []storage.Artifact{
		storage.NewFile("repomd.xml", repomdContent.Bytes()),
		storage.NewFile("repomd.xml.asc", repomdAscContent.Bytes()),
	}, nil
}

func buildPrimary(_ context.Context, packages ...*Package) (*RepoData, storage.Artifact, error) {
	type Version struct {
		Epoch   string `xml:"epoch,attr"`
		Version string `xml:"ver,attr"`
		Release string `xml:"rel,attr"`
	}

	type Checksum struct {
		Checksum string `xml:",chardata"`
		Type     string `xml:"type,attr"`
		Pkgid    string `xml:"pkgid,attr"`
	}

	type Times struct {
		File  uint64 `xml:"file,attr"`
		Build uint64 `xml:"build,attr"`
	}

	type Sizes struct {
		Package   int64  `xml:"package,attr"`
		Installed uint64 `xml:"installed,attr"`
		Archive   uint64 `xml:"archive,attr"`
	}

	type Location struct {
		Href string `xml:"href,attr"`
	}

	type EntryList struct {
		Entries []*Entry `xml:"rpm:entry"`
	}

	type Format struct {
		License   string    `xml:"rpm:license"`
		Vendor    string    `xml:"rpm:vendor"`
		Group     string    `xml:"rpm:group"`
		Buildhost string    `xml:"rpm:buildhost"`
		Sourcerpm string    `xml:"rpm:sourcerpm"`
		Provides  EntryList `xml:"rpm:provides"`
		Requires  EntryList `xml:"rpm:requires"`
		Conflicts EntryList `xml:"rpm:conflicts"`
		Obsoletes EntryList `xml:"rpm:obsoletes"`
		Files     []*File   `xml:"file"`
	}

	type Package struct {
		XMLName      xml.Name `xml:"package"`
		Type         string   `xml:"type,attr"`
		Name         string   `xml:"name"`
		Architecture string   `xml:"arch"`
		Version      Version  `xml:"version"`
		Checksum     Checksum `xml:"checksum"`
		Summary      string   `xml:"summary"`
		Description  string   `xml:"description"`
		Packager     string   `xml:"packager"`
		URL          string   `xml:"url"`
		Time         Times    `xml:"time"`
		Size         Sizes    `xml:"size"`
		Location     Location `xml:"location"`
		Format       Format   `xml:"format"`
	}

	type Metadata struct {
		XMLName      xml.Name   `xml:"metadata"`
		Xmlns        string     `xml:"xmlns,attr"`
		XmlnsRpm     string     `xml:"xmlns:rpm,attr"`
		PackageCount int        `xml:"packages,attr"`
		Packages     []*Package `xml:"package"`
	}

	pkgs := make([]*Package, 0, len(packages))
	for _, pd := range packages {

		files := make([]*File, 0, 3)
		for _, f := range pd.FileMetadata.Files {
			if f.IsExecutable {
				files = append(files, f)
			}
		}

		pkgs = append(pkgs, &Package{
			Type:         "rpm",
			Name:         pd.PkgName,
			Architecture: pd.FileMetadata.Architecture,
			Version: Version{
				Epoch:   pd.FileMetadata.Epoch,
				Version: pd.FileMetadata.Version,
				Release: pd.FileMetadata.Release,
			},
			Checksum: Checksum{
				Type:     "sha256",
				Checksum: pd.HashSHA256,
				Pkgid:    "YES",
			},
			Summary:     pd.VersionMetadata.Summary,
			Description: pd.VersionMetadata.Description,
			Packager:    pd.FileMetadata.Packager,
			URL:         pd.VersionMetadata.ProjectURL,
			Time: Times{
				File:  pd.FileMetadata.FileTime,
				Build: pd.FileMetadata.BuildTime,
			},
			Size: Sizes{
				Package:   pd.FileSize,
				Installed: pd.FileMetadata.InstalledSize,
				Archive:   pd.FileMetadata.ArchiveSize,
			},
			Location: Location{
				Href: pd.Path(),
			},
			Format: Format{
				License:   pd.VersionMetadata.License,
				Vendor:    pd.FileMetadata.Vendor,
				Group:     pd.FileMetadata.Group,
				Buildhost: pd.FileMetadata.BuildHost,
				Sourcerpm: pd.FileMetadata.SourceRpm,
				Provides: EntryList{
					Entries: pd.FileMetadata.Provides,
				},
				Requires: EntryList{
					Entries: pd.FileMetadata.Requires,
				},
				Conflicts: EntryList{
					Entries: pd.FileMetadata.Conflicts,
				},
				Obsoletes: EntryList{
					Entries: pd.FileMetadata.Obsoletes,
				},
				Files: files,
			},
		})
	}

	return newRepoData("primary", &Metadata{
		Xmlns:        "http://linux.duke.edu/metadata/common",
		XmlnsRpm:     "http://linux.duke.edu/metadata/rpm",
		PackageCount: len(pkgs),
		Packages:     pkgs,
	})
}

// https://docs.pulpproject.org/en/2.19/plugins/pulp_rpm/tech-reference/rpm.html#filelists-xml
func buildFilelists(_ context.Context, packages ...*Package) (*RepoData, storage.Artifact, error) { //nolint:dupl
	type Version struct {
		Epoch   string `xml:"epoch,attr"`
		Version string `xml:"ver,attr"`
		Release string `xml:"rel,attr"`
	}

	type Package struct {
		Pkgid        string  `xml:"pkgid,attr"`
		Name         string  `xml:"name,attr"`
		Architecture string  `xml:"arch,attr"`
		Version      Version `xml:"version"`
		Files        []*File `xml:"file"`
	}

	type Filelists struct {
		XMLName      xml.Name   `xml:"filelists"`
		Xmlns        string     `xml:"xmlns,attr"`
		PackageCount int        `xml:"packages,attr"`
		Packages     []*Package `xml:"package"`
	}

	pkgs := make([]*Package, 0, len(packages))
	for _, pd := range packages {
		pkgs = append(pkgs, &Package{
			Pkgid:        pd.HashSHA256,
			Name:         pd.PkgName,
			Architecture: pd.FileMetadata.Architecture,
			Version: Version{
				Epoch:   pd.FileMetadata.Epoch,
				Version: pd.FileMetadata.Version,
				Release: pd.FileMetadata.Release,
			},
			Files: pd.FileMetadata.Files,
		})
	}

	return newRepoData("filelists", &Filelists{
		Xmlns:        "http://linux.duke.edu/metadata/other",
		PackageCount: len(pkgs),
		Packages:     pkgs,
	})
}

func newRepoData(filetype string, obj any) (*RepoData, storage.Artifact, error) {
	content, _ := buffer.NewHashedBuffer()
	gzw := gzip.NewWriter(content)
	wc := &writtenCounter{}
	h := sha256.New()

	w := io.MultiWriter(gzw, wc, h)
	_, _ = w.Write([]byte(xml.Header))

	if err := encode(w, obj); err != nil {
		return nil, nil, err
	}

	if err := gzw.Close(); err != nil {
		return nil, nil, err
	}

	data, err := io.ReadAll(content)
	if err != nil {
		return nil, nil, err
	}
	filename := filetype + ".xml.gz"
	file := storage.NewFile(filename, data)

	_, _, hashSHA256, _ := content.Sums()

	return &RepoData{
		Type: filetype,
		Checksum: RepoChecksum{
			Type:  "sha256",
			Value: hex.EncodeToString(hashSHA256),
		},
		OpenChecksum: RepoChecksum{
			Type:  "sha256",
			Value: hex.EncodeToString(h.Sum(nil)),
		},
		Location: RepoLocation{
			Href: "repodata/" + filename,
		},
		Timestamp: time.Now().Unix(),
		Size:      content.Size(),
		OpenSize:  wc.Written(),
	}, file, nil
}

// https://docs.pulpproject.org/en/2.19/plugins/pulp_rpm/tech-reference/rpm.html#other-xml
func buildOther(ctx context.Context, packages ...*Package) (*RepoData, storage.Artifact, error) { //nolint:dupl
	type Version struct {
		Epoch   string `xml:"epoch,attr"`
		Version string `xml:"ver,attr"`
		Release string `xml:"rel,attr"`
	}

	type Package struct {
		Pkgid        string       `xml:"pkgid,attr"`
		Name         string       `xml:"name,attr"`
		Architecture string       `xml:"arch,attr"`
		Version      Version      `xml:"version"`
		Changelogs   []*Changelog `xml:"changelog"`
	}

	type Otherdata struct {
		XMLName      xml.Name   `xml:"otherdata"`
		Xmlns        string     `xml:"xmlns,attr"`
		PackageCount int        `xml:"packages,attr"`
		Packages     []*Package `xml:"package"`
	}

	pkgs := make([]*Package, 0, len(packages))
	for _, pd := range packages {
		pkgs = append(pkgs, &Package{
			Pkgid:        pd.HashSHA256,
			Name:         pd.PkgName,
			Architecture: pd.FileMetadata.Architecture,
			Version: Version{
				Epoch:   pd.FileMetadata.Epoch,
				Version: pd.FileMetadata.Version,
				Release: pd.FileMetadata.Release,
			},
			Changelogs: pd.FileMetadata.Changelogs,
		})
	}

	return newRepoData("other", &Otherdata{
		Xmlns:        "http://linux.duke.edu/metadata/other",
		PackageCount: len(pkgs),
		Packages:     pkgs,
	})
}

// writtenCounter counts all written bytes
type writtenCounter struct {
	written int64
}

func (wc *writtenCounter) Write(buf []byte) (int, error) {
	n := len(buf)

	wc.written += int64(n)

	return n, nil
}

func (wc *writtenCounter) Written() int64 {
	return wc.written
}

func encode(w io.Writer, data any) error {
	e := xml.NewEncoder(w)
	e.Indent("", "  ")
	return e.Encode(data)
}
