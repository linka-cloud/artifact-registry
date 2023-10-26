// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package rpm

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/opencontainers/go-digest"
	"github.com/sassoftware/go-rpmutils"

	"go.linka.cloud/artifact-registry/pkg/buffer"
	"go.linka.cloud/artifact-registry/pkg/storage"
	"go.linka.cloud/artifact-registry/pkg/timeutil"
	"go.linka.cloud/artifact-registry/pkg/validation"
)

const (
	// Can't use the syscall constants because they are not available for windows build.
	sIFMT  = 0xf000
	sIFDIR = 0x4000
	sIXUSR = 0x40
	sIXGRP = 0x8
	sIXOTH = 0x1
)

var _ storage.Artifact = (*Package)(nil)

// https://rpm-software-management.github.io/rpm/manual/spec.html
// https://refspecs.linuxbase.org/LSB_3.1.0/LSB-Core-generic/LSB-Core-generic/pkgformat.html

type Package struct {
	PkgName         string           `json:"name"`
	PkgVersion      string           `json:"version"`
	VersionMetadata *VersionMetadata `json:"versionMetadata"`
	FileMetadata    *FileMetadata    `json:"fileMetadata"`
	HashSHA256      string           `json:"hashSha256"`
	FileSize        int64            `json:"size"`
	FilePath        string           `json:"filePath"`

	reader io.ReadCloser
}

func (p *Package) Name() string {
	return p.PkgName
}

func (p *Package) Arch() string {
	return p.FileMetadata.Architecture
}

func (p *Package) Version() string {
	return p.PkgVersion
}

func (p *Package) Read(b []byte) (int, error) {
	if p.reader == nil {
		return 0, io.EOF
	}
	return p.reader.Read(b)
}

func (p *Package) Digest() digest.Digest {
	return digest.NewDigestFromEncoded(digest.SHA256, p.HashSHA256)
}

func (p *Package) Path() string {
	return p.FilePath
}

func (p *Package) Size() int64 {
	return p.FileSize
}

func (p *Package) Close() error {
	if p.reader == nil {
		return nil
	}
	return p.reader.Close()
}

type VersionMetadata struct {
	ProjectURL  string `json:"projectURL,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	License     string `json:"license,omitempty"`
}

type FileMetadata struct {
	Architecture  string `json:"architecture,omitempty"`
	Epoch         string `json:"epoch,omitempty"`
	Version       string `json:"version,omitempty"`
	Release       string `json:"release,omitempty"`
	Vendor        string `json:"vendor,omitempty"`
	Group         string `json:"group,omitempty"`
	Packager      string `json:"packager,omitempty"`
	SourceRpm     string `json:"sourceRPM,omitempty"`
	BuildHost     string `json:"buildHost,omitempty"`
	BuildTime     uint64 `json:"buildTime,omitempty"`
	FileTime      uint64 `json:"fileTime,omitempty"`
	InstalledSize uint64 `json:"installedSize,omitempty"`
	ArchiveSize   uint64 `json:"archiveSize,omitempty"`

	Provides  []*Entry `json:"provide,omitempty"`
	Requires  []*Entry `json:"require,omitempty"`
	Conflicts []*Entry `json:"conflict,omitempty"`
	Obsoletes []*Entry `json:"obsolete,omitempty"`

	Files []*File `json:"files,omitempty"`

	Changelogs []*Changelog `json:"changelogs,omitempty"`
}

type Entry struct {
	Name    string `json:"name" xml:"name,attr"`
	Flags   string `json:"flags,omitempty" xml:"flags,attr,omitempty"`
	Version string `json:"version,omitempty" xml:"ver,attr,omitempty"`
	Epoch   string `json:"epoch,omitempty" xml:"epoch,attr,omitempty"`
	Release string `json:"release,omitempty" xml:"rel,attr,omitempty"`
}

type File struct {
	Path         string `json:"path" xml:",chardata"`
	Type         string `json:"type,omitempty" xml:"type,attr,omitempty"`
	IsExecutable bool   `json:"isExecutable" xml:"-"`
}

type Changelog struct {
	Author string             `json:"author,omitempty" xml:"author,attr"`
	Date   timeutil.TimeStamp `json:"date,omitempty" xml:"date,attr"`
	Text   string             `json:"text,omitempty" xml:",chardata"`
}

func NewPackage(r io.Reader, size int64, key string) (*Package, error) {
	// sign
	buf, err := buffer.CreateHashedBufferFromReader(r)
	if err != nil {
		return nil, err
	}
	defer buf.Close()
	hBuf, ssize, seek, err := SignPackage(buf, key)
	if err != nil {
		return nil, err
	}
	if _, err := buf.Seek(seek, io.SeekStart); err != nil {
		return nil, err
	}
	reader, err := buffer.CreateHashedBufferFromReader(io.MultiReader(hBuf, buf))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	pkg, err := parsePackage(reader)
	if err != nil {
		return nil, err
	}
	pkg.FileSize = size + ssize - seek

	_, _, sha256, _ := reader.Sums()
	pkg.HashSHA256 = hex.EncodeToString(sha256)
	// create image with package as config and blob as layer
	if _, err := reader.Seek(0, 0); err != nil {
		return nil, err
	}
	return pkg, nil
}

// parsePackage parses the RPM package file
func parsePackage(r io.ReadCloser) (*Package, error) {
	rpm, err := rpmutils.ReadRpm(r)
	if err != nil {
		return nil, err
	}

	nevra, err := rpm.Header.GetNEVRA()
	if err != nil {
		return nil, err
	}

	version := fmt.Sprintf("%s-%s", nevra.Version, nevra.Release)
	if nevra.Epoch != "" && nevra.Epoch != "0" {
		version = fmt.Sprintf("%s-%s", nevra.Epoch, version)
	}

	p := &Package{
		PkgName:    nevra.Name,
		PkgVersion: version,
		VersionMetadata: &VersionMetadata{
			Summary:     getString(rpm.Header, rpmutils.SUMMARY),
			Description: getString(rpm.Header, rpmutils.DESCRIPTION),
			License:     getString(rpm.Header, rpmutils.LICENSE),
			ProjectURL:  getString(rpm.Header, rpmutils.URL),
		},
		FileMetadata: &FileMetadata{
			Architecture:  nevra.Arch,
			Epoch:         nevra.Epoch,
			Version:       nevra.Version,
			Release:       nevra.Release,
			Vendor:        getString(rpm.Header, rpmutils.VENDOR),
			Group:         getString(rpm.Header, rpmutils.GROUP),
			Packager:      getString(rpm.Header, rpmutils.PACKAGER),
			SourceRpm:     getString(rpm.Header, rpmutils.SOURCERPM),
			BuildHost:     getString(rpm.Header, rpmutils.BUILDHOST),
			BuildTime:     getUInt64(rpm.Header, rpmutils.BUILDTIME),
			FileTime:      getUInt64(rpm.Header, rpmutils.FILEMTIMES),
			InstalledSize: getUInt64(rpm.Header, rpmutils.SIZE),
			ArchiveSize:   getUInt64(rpm.Header, rpmutils.SIG_PAYLOADSIZE),

			Provides:   getEntries(rpm.Header, rpmutils.PROVIDENAME, rpmutils.PROVIDEVERSION, rpmutils.PROVIDEFLAGS),
			Requires:   getEntries(rpm.Header, rpmutils.REQUIRENAME, rpmutils.REQUIREVERSION, rpmutils.REQUIREFLAGS),
			Conflicts:  getEntries(rpm.Header, 1054 /*rpmutils.CONFLICTNAME*/, 1055 /*rpmutils.CONFLICTVERSION*/, 1053 /*rpmutils.CONFLICTFLAGS*/), // https://github.com/sassoftware/go-rpmutils/pull/24
			Obsoletes:  getEntries(rpm.Header, rpmutils.OBSOLETENAME, rpmutils.OBSOLETEVERSION, rpmutils.OBSOLETEFLAGS),
			Files:      getFiles(rpm.Header),
			Changelogs: getChangelogs(rpm.Header),
		},
		reader: r,
	}
	p.FilePath = fmt.Sprintf("%s-%s.%s.rpm", p.PkgName, p.PkgVersion, p.FileMetadata.Architecture)

	if !validation.IsValidURL(p.VersionMetadata.ProjectURL) {
		p.VersionMetadata.ProjectURL = ""
	}

	return p, nil
}

func getString(h *rpmutils.RpmHeader, tag int) string {
	values, err := h.GetStrings(tag)
	if err != nil || len(values) < 1 {
		return ""
	}
	return values[0]
}

func getUInt64(h *rpmutils.RpmHeader, tag int) uint64 {
	values, err := h.GetUint64s(tag)
	if err != nil || len(values) < 1 {
		return 0
	}
	return values[0]
}

func getEntries(h *rpmutils.RpmHeader, namesTag, versionsTag, flagsTag int) []*Entry {
	names, err := h.GetStrings(namesTag)
	if err != nil || len(names) == 0 {
		return nil
	}
	flags, err := h.GetUint64s(flagsTag)
	if err != nil || len(flags) == 0 {
		return nil
	}
	versions, err := h.GetStrings(versionsTag)
	if err != nil || len(versions) == 0 {
		return nil
	}
	if len(names) != len(flags) || len(names) != len(versions) {
		return nil
	}

	entries := make([]*Entry, 0, len(names))
	for i := range names {
		e := &Entry{
			Name: names[i],
		}

		flags := flags[i]
		if (flags&rpmutils.RPMSENSE_GREATER) != 0 && (flags&rpmutils.RPMSENSE_EQUAL) != 0 {
			e.Flags = "GE"
		} else if (flags&rpmutils.RPMSENSE_LESS) != 0 && (flags&rpmutils.RPMSENSE_EQUAL) != 0 {
			e.Flags = "LE"
		} else if (flags & rpmutils.RPMSENSE_GREATER) != 0 {
			e.Flags = "GT"
		} else if (flags & rpmutils.RPMSENSE_LESS) != 0 {
			e.Flags = "LT"
		} else if (flags & rpmutils.RPMSENSE_EQUAL) != 0 {
			e.Flags = "EQ"
		}

		version := versions[i]
		if version != "" {
			parts := strings.Split(version, "-")

			versionParts := strings.Split(parts[0], ":")
			if len(versionParts) == 2 {
				e.Version = versionParts[1]
				e.Epoch = versionParts[0]
			} else {
				e.Version = versionParts[0]
				e.Epoch = "0"
			}

			if len(parts) > 1 {
				e.Release = parts[1]
			}
		}

		entries = append(entries, e)
	}
	return entries
}

func getFiles(h *rpmutils.RpmHeader) []*File {
	baseNames, _ := h.GetStrings(rpmutils.BASENAMES)
	dirNames, _ := h.GetStrings(rpmutils.DIRNAMES)
	dirIndexes, _ := h.GetUint32s(rpmutils.DIRINDEXES)
	fileFlags, _ := h.GetUint32s(rpmutils.FILEFLAGS)
	fileModes, _ := h.GetUint32s(rpmutils.FILEMODES)

	files := make([]*File, 0, len(baseNames))
	for i := range baseNames {
		if len(dirIndexes) <= i {
			continue
		}
		dirIndex := dirIndexes[i]
		if len(dirNames) <= int(dirIndex) {
			continue
		}

		var fileType string
		var isExecutable bool
		if i < len(fileFlags) && (fileFlags[i]&rpmutils.RPMFILE_GHOST) != 0 {
			fileType = "ghost"
		} else if i < len(fileModes) {
			if (fileModes[i] & sIFMT) == sIFDIR {
				fileType = "dir"
			} else {
				mode := fileModes[i] & ^uint32(sIFMT)
				isExecutable = (mode&sIXUSR) != 0 || (mode&sIXGRP) != 0 || (mode&sIXOTH) != 0
			}
		}

		files = append(files, &File{
			Path:         dirNames[dirIndex] + baseNames[i],
			Type:         fileType,
			IsExecutable: isExecutable,
		})
	}

	return files
}

func getChangelogs(h *rpmutils.RpmHeader) []*Changelog {
	texts, err := h.GetStrings(rpmutils.CHANGELOGTEXT)
	if err != nil || len(texts) == 0 {
		return nil
	}
	authors, err := h.GetStrings(rpmutils.CHANGELOGNAME)
	if err != nil || len(authors) == 0 {
		return nil
	}
	times, err := h.GetUint32s(rpmutils.CHANGELOGTIME)
	if err != nil || len(times) == 0 {
		return nil
	}
	if len(texts) != len(authors) || len(texts) != len(times) {
		return nil
	}

	changelogs := make([]*Changelog, 0, len(texts))
	for i := range texts {
		changelogs = append(changelogs, &Changelog{
			Author: authors[i],
			Date:   timeutil.TimeStamp(times[i]),
			Text:   texts[i],
		})
	}
	return changelogs
}
