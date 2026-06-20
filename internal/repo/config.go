package repo

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/ini.v1"
)

// Remote is a configured storage remote (S3-compatible in the MVP).
type Remote struct {
	Name            string
	URL             string
	EndpointURL     string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Profile         string
}

// Config is the parsed .dvc/config.
type Config struct {
	CoreRemote string
	CacheType  string
	Remotes    map[string]Remote
}

const defaultCacheType = "reflink,copy"

// LoadConfig parses .dvc/config. A missing file yields defaults.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{CacheType: defaultCacheType, Remotes: map[string]Remote{}}
	f, err := ini.Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if s, _ := f.GetSection("core"); s != nil && s.HasKey("remote") {
		cfg.CoreRemote = s.Key("remote").String()
	}
	if s, _ := f.GetSection("cache"); s != nil && s.HasKey("type") {
		cfg.CacheType = s.Key("type").String()
	}
	for _, sec := range f.Sections() {
		name, ok := remoteName(sec.Name())
		if !ok {
			continue
		}
		cfg.Remotes[name] = Remote{
			Name:            name,
			URL:             sec.Key("url").String(),
			EndpointURL:     sec.Key("endpointurl").String(),
			Region:          sec.Key("region").String(),
			AccessKeyID:     sec.Key("access_key_id").String(),
			SecretAccessKey: sec.Key("secret_access_key").String(),
			SessionToken:    sec.Key("session_token").String(),
			Profile:         sec.Key("profile").String(),
		}
	}
	return cfg, nil
}

// remoteName extracts "name" from a section named `remote "name"`.
func remoteName(section string) (string, bool) {
	const prefix = `remote "`
	if strings.HasPrefix(section, prefix) && strings.HasSuffix(section, `"`) {
		return section[len(prefix) : len(section)-1], true
	}
	return "", false
}

// SetRemote adds or updates a remote in .dvc/config, optionally as the default.
func SetRemote(path string, r Remote, makeDefault bool) error {
	f, err := ini.Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			f = ini.Empty()
		} else {
			return err
		}
	}
	sec := f.Section(fmt.Sprintf(`remote "%s"`, r.Name))
	setOrDelete(sec, "url", r.URL)
	setOrDelete(sec, "endpointurl", r.EndpointURL)
	setOrDelete(sec, "region", r.Region)
	setOrDelete(sec, "access_key_id", r.AccessKeyID)
	setOrDelete(sec, "secret_access_key", r.SecretAccessKey)
	setOrDelete(sec, "session_token", r.SessionToken)
	setOrDelete(sec, "profile", r.Profile)

	if makeDefault {
		f.Section("core").Key("remote").SetValue(r.Name)
	}
	return f.SaveTo(path)
}

func setOrDelete(sec *ini.Section, key, val string) {
	if val == "" {
		sec.DeleteKey(key)
		return
	}
	sec.Key(key).SetValue(val)
}
