package repo

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kellegous/bungler/util"
)

var versionPattern = regexp.MustCompile("^[\\.0-9]+$")

const (
	baseURL        = "https://repo.maven.apache.org/maven2"
	versionLatest  = "latest"
	versionRelease = "release"
)

// Type ...
type Type int

const (
	// Jar ..
	Jar Type = iota

	// Src ...
	Src

	// Doc ...
	Doc
)

// Dep ...
type Dep struct {
	Org      string
	Artifact string
	Version  string
}

// BaseURL ...
func (d *Dep) BaseURL() string {
	return fmt.Sprintf("%s/%s/%s",
		baseURL,
		strings.Replace(d.Org, ".", "/", -1),
		d.Artifact)
}

func jarNameFor(d *Dep, ver string, t Type) string {
	var n string
	switch t {
	case Doc:
		n = "-javadoc"
	case Src:
		n = "-sources"
	}

	return fmt.Sprintf("%s-%s%s.jar", d.Artifact, ver, n)
}

func nameOf(d *Dep) string {
	ver := d.Version
	if ver == "" {
		ver = versionRelease
	}
	return fmt.Sprintf("%s/%s/%s", d.Org, d.Artifact, ver)
}

func download(d *Dep, dir, ver string, t Type) error {

	dst := jarNameFor(d, ver, t)

	return util.Fetch(
		filepath.Join(dir, dst),
		fmt.Sprintf("%s/%s/%s", d.BaseURL(), ver, dst))
}

// Download ...
func (d *Dep) Download(dir string, types []Type) error {
	fmt.Printf("%s\n", nameOf(d))

	ver, err := resolveVersion(d)
	if err != nil {
		return err
	}

	for _, t := range types {
		if err := download(d, dir, ver, t); err != nil {
			return err
		}
	}

	deps, err := depsOf(d, ver)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		if err := dep.Download(dir, types); err != nil {
			return err
		}
	}

	return nil
}

// Parse ...
func (d *Dep) Parse(uri string) error {
	p := strings.Split(uri, "/")

	switch len(p) {
	case 2:
		d.Org, d.Artifact, d.Version = p[0], p[1], ""
	case 3:
		d.Org, d.Artifact, d.Version = p[0], p[1], p[2]
	default:
		return fmt.Errorf("invalid dependency: %s", uri)
	}

	return nil
}

func resolveVersion(d *Dep) (string, error) {
	if d.Version != versionLatest && d.Version != versionRelease && d.Version != "" {
		return d.Version, nil
	}

	vers, err := d.Versions()
	if err != nil {
		return "", err
	}

	switch d.Version {
	case versionLatest:
		return vers.Latest, nil
	case "", versionRelease:
		return vers.Release, nil
	}

	panic("unreachable")
}

// Deps ...
func (d *Dep) Deps() ([]*Dep, error) {
	ver, err := resolveVersion(d)
	if err != nil {
		return nil, err
	}

	return depsOf(d, ver)
}

func versionFrom(ver string) string {
	if versionPattern.MatchString(ver) {
		return ver
	}
	return ""
}

func depsOf(d *Dep, ver string) ([]*Dep, error) {
	res, err := util.GetWithCheck(
		fmt.Sprintf("%s/%s/%s-%s.pom", d.BaseURL(), ver, d.Artifact, ver))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var p struct {
		Deps []*struct {
			Org      string `xml:"groupId"`
			Artifact string `xml:"artifactId"`
			Version  string `xml:"version"`
			Optional bool   `xml:"optional"`
			Scope    string `xml:"scope"`
		} `xml:"dependencies>dependency"`
	}

	if err := xml.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, err
	}

	var deps []*Dep
	for _, d := range p.Deps {
		if d.Optional {
			continue
		}

		if d.Scope != "" && d.Scope != "runtime" {
			continue
		}

		deps = append(deps, &Dep{
			Org:      d.Org,
			Artifact: d.Artifact,
			Version:  versionFrom(d.Version),
		})
	}

	return deps, nil
}

// VersionInfo ...
type VersionInfo struct {
	Versions []string `xml:"versioning>versions>version"`
	Latest   string   `xml:"versioning>latest"`
	Release  string   `xml:"versioning>release"`
}

// Versions ...
func (d *Dep) Versions() (*VersionInfo, error) {
	res, err := util.GetWithCheck(fmt.Sprintf("%s/maven-metadata.xml", d.BaseURL()))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	nfo := &VersionInfo{}
	if err := xml.NewDecoder(res.Body).Decode(nfo); err != nil {
		return nil, err
	}

	return nfo, nil
}
