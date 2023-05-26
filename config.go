package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/johan-st/go-image-server/images"
	"gopkg.in/yaml.v3"
)

type config struct {
	LogLevel      string            `yaml:"log_level"`
	Http          confHttp          `yaml:"http"`
	Files         confFiles         `yaml:"files"`
	Cache         confCache         `yaml:"cache_rules"`
	ImageDefaults confImageDefault  `yaml:"image_defaults"`
	ImagePresets  []confImagePreset `yaml:"image_presets"`
}

type confHttp struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
	Docs bool   `yaml:"documentation"`
}

type confFiles struct {
	SetPerms     bool   `yaml:"set_perms"`
	CreateDirs   bool   `yaml:"create_dirs"`
	DirOriginals string `yaml:"originals_dir"`
	DirCache     string `yaml:"cache_dir"`

	// debug options TODO: should these be here?
	ClearOnStart bool   `yaml:"clear_on_start"`
	ClearOnExit  bool   `yaml:"clear_on_exit"`
	PopulateFrom string `yaml:"populate_from"`
}

type confCache struct {
	Cap     int    `yaml:"max_objects"`
	MaxSize string `yaml:"max_size"`
}

type confImageDefault struct {
	Format        string `yaml:"format"`
	QualityJpeg   int    `yaml:"quality_jpeg"`
	QualityGif    int    `yaml:"quality_gif"`
	Width         int    `yaml:"width"`
	Height        int    `yaml:"height"`
	MaxSize       string `yaml:"max_size"`
	Interpolation string `yaml:"interpolation"`
}

type confImagePreset struct {
	Name          string   `yaml:"name"`
	Alias         []string `yaml:"alias"`
	Format        string   `yaml:"format,omitempty"`
	Quality       int      `yaml:"quality,omitempty"`
	Width         int      `yaml:"width"`
	Height        int      `yaml:"height"`
	MaxSize       string   `yaml:"max_size,omitempty"`
	Interpolation string   `yaml:"interpolation,omitempty"`
}

func saveConfig(c config, filename string) error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, bytes, 0644)
}

func loadConfig(filename string) (config, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return config{}, err
	}

	var c config
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		return config{}, err
	}

	return c, nil
}

// validate enforces config rules and returns an error if any are broken
func (c *config) validate() error {
	errs := []error{}

	// HTTP
	// port needed
	if c.Http.Port == 0 {
		errs = append(errs, fmt.Errorf("server port must be set"))
	}
	// empty host is ok
	// TODO: validate host format

	// FILES
	if c.Files.DirOriginals == "" {
		errs = append(errs, fmt.Errorf("path for originals must be set"))
	}
	if c.Files.DirCache == "" {
		errs = append(errs, fmt.Errorf("paths for cache must be set"))
	}
	if c.Cache.Cap == 0 {
		errs = append(errs, fmt.Errorf("cache num must be greater than 0"))
	}

	// DEFAULT IMAGE PARAMETERS
	if c.ImageDefaults.Format != "jpeg" && c.ImageDefaults.Format != "png" && c.ImageDefaults.Format != "gif" {
		errs = append(errs, fmt.Errorf("default image parameters format must be set to a valid value. Valid values are: jpeg, png, gif"))
	}
	if c.ImageDefaults.QualityJpeg == 0 {
		errs = append(errs, fmt.Errorf("default image parameters quality jpeg must be set to a value greater between 1 and 100 (inclusive)"))
	}
	if c.ImageDefaults.QualityGif == 0 {
		errs = append(errs, fmt.Errorf("default image parameters quality gif must be set to a value greater between 1 and 256 (inclusive)"))
	}
	if c.ImageDefaults.Width == 0 && c.ImageDefaults.Height == 0 {
		errs = append(errs, fmt.Errorf("default image parameters width or height (or both) must be set"))
	}
	// TODO: validate max_size format
	// 0 is ok for max_size, it means no limit
	// TODO: validate resize format

	// IMAGE PARAMETERS
	for _, p := range c.ImagePresets {
		name := p.Name
		if name == "" {
			errs = append(errs, fmt.Errorf("image parameters name must be set"))
		}
		if p.Format != "" && p.Format != "jpeg" && p.Format != "png" && p.Format != "gif" {
			errs = append(errs, fmt.Errorf("image parameters (name: \"%s\") format must be set to a valid value. Valid values are: jpeg, png, gif", name))
		}
		if p.Quality == 0 && p.Format == "jpeg" {
			errs = append(errs, fmt.Errorf("image parameters (name: \"%s\") quality must be set to a value greater between 1 and 100 (inclusive)", name))
		}
		if p.Quality == 0 && p.Format == "gif" {
			errs = append(errs, fmt.Errorf("image parameters (name: \"%s\") quality must be set to a value greater between 1 and 256 (inclusive)", name))
		}
		if p.Width == 0 && p.Height == 0 {
			errs = append(errs, fmt.Errorf("image parameters (name: \"%s\") width or height (or both) must be set", name))
		}

		// TODO: validate max_size format
		// 0 is ok for max_size, it means no limit
		// TODO: validate resize format
	}

	// Return errors if any
	if len(errs) > 0 {
		errs = append(errs, fmt.Errorf("config validation failed"))
		return errors.Join(errs...)
		// return fmt.Errorf("config validation failed: %v", errs)
	}
	return nil
}

// TODO: handle errors by returning them?
func toImageDefaults(c confImageDefault) (images.ImageDefaults, error) {
	errs := []error{}

	format, err := images.ParseFormat(c.Format)
	if err != nil {
		errs = append(errs, err)
	}
	size, err := images.ParseSize(c.MaxSize)
	if err != nil {
		errs = append(errs, err)
	}

	interpolation, err := images.ParseInterpolation(c.Interpolation)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		newErrs := []error{fmt.Errorf("(%d) errors while building ImageDefaults", len(errs))}
		newErrs = append(newErrs, errs...)
		return images.ImageDefaults{}, errors.Join(newErrs...)
	}

	return images.ImageDefaults{
		Format:        format,
		QualityJpeg:   c.QualityJpeg,
		QualityGif:    c.QualityGif,
		Width:         c.Width,
		Height:        c.Height,
		MaxSize:       size,
		Interpolation: interpolation,
	}, nil
}

func toImagePresets(conf []confImagePreset, def images.ImageDefaults) ([]images.ImagePreset, error) {
	presets := []images.ImagePreset{}
	var err error
	errs := []error{}

	for _, cp := range conf {
		// format
		var format images.Format
		if cp.Format != "" {
			format, err = images.ParseFormat(cp.Format)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			format = def.Format
		}

		// size
		var size images.Size
		if cp.MaxSize != "" {
			size, err = images.ParseSize(cp.MaxSize)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			size = def.MaxSize
		}

		// interpolation
		var interpolation images.Interpolation
		if cp.Interpolation != "" {
			interpolation, err = images.ParseInterpolation(cp.Interpolation)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			interpolation = def.Interpolation
		}

		// resulting preset
		p := images.ImagePreset{
			Name:          cp.Name,
			Alias:         cp.Alias,
			Format:        format,
			Quality:       cp.Quality,
			Width:         cp.Width,
			Height:        cp.Height,
			MaxSize:       size,
			Interpolation: interpolation,
		}
		presets = append(presets, p)
	}

	if len(errs) > 0 {
		newErrs := []error{fmt.Errorf("(%d) errors while building ImagePresets", len(errs))}
		newErrs = append(newErrs, errs...)
		return []images.ImagePreset{}, errors.Join(newErrs...)
	}
	return presets, nil
}

func defaultConfig() config {
	return config{
		LogLevel: "info",
		Http: confHttp{
			Port: 8080,
			Host: "",
			Docs: false,
		},
		Files: confFiles{
			ClearOnStart: false,
			PopulateFrom: "",
			SetPerms:     false,
			CreateDirs:   false,
			DirOriginals: "img/originals",
			DirCache:     "img/cached",
		},
		Cache: confCache{
			Cap:     100000,
			MaxSize: "500 GB",
		},
		ImageDefaults: confImageDefault{
			Format:        "jpeg",
			QualityJpeg:   80,
			QualityGif:    256,
			Width:         0,
			Height:        800,
			MaxSize:       "1 MB",
			Interpolation: "nearestNeighbor",
		},
		ImagePresets: []confImagePreset{
			{
				Name:          "thumbnail",
				Alias:         []string{"thumb", "th"},
				Format:        "jpeg",
				Quality:       80,
				Width:         150,
				Height:        150,
				MaxSize:       "10 KB",
				Interpolation: "lanczos3",
			},
			{
				Name:   "small",
				Alias:  []string{"small", "s"},
				Height: 400,
				Width:  0,
			},
			{
				Name:   "medium",
				Alias:  []string{"medium", "m"},
				Height: 800,
				Width:  0,
			},
			{
				Name:   "large",
				Alias:  []string{"large", "l"},
				Height: 1600,
				Width:  0,
			},
		},
	}
}
