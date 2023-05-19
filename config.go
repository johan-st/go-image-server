package main

import (
	"fmt"
	"io/ioutil"

	"github.com/johan-st/go-image-server/images"
	"gopkg.in/yaml.v3"
)

// CONFIG STUFF
type Config struct {
	Logging                string                     `yaml:"logging"`
	Server                 ConfServer                 `yaml:"http"`
	Handler                ConfHandler                `yaml:"files"`
	ImageParametersDefault ConfImageParametersDefault `yaml:"default_image_preset"`
	ImageParameters        []ConfImageParameters      `yaml:"image_presets"`
}

type ConfServer struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type ConfHandler struct {
	CleanStart   bool             `yaml:"clean_start"`
	PopulateFrom string           `yaml:"populate_from"`
	Paths        ConfHandlerPaths `yaml:"paths"`
	Cache        ConfHandlerCache `yaml:"cache"`
}

type ConfHandlerPaths struct {
	Originals  string `yaml:"originals"`
	Cache      string `yaml:"cache"`
	SetPerms   bool   `yaml:"set_perms"`
	CreateDirs bool   `yaml:"create_dirs"`
}

type ConfHandlerCache struct {
	MaxNum  int    `yaml:"num"`
	MaxSize string `yaml:"size"`
}

type ConfImageParametersDefault struct {
	Alias       []string `yaml:"alias"`
	Format      string   `yaml:"format"`
	QualityJpeg int      `yaml:"quality_jpeg"`
	QualityGif  int      `yaml:"quality_gif"`
	Width       int      `yaml:"width"`
	Height      int      `yaml:"height"`
	MaxSize     string   `yaml:"max_size"`
	Resize      string   `yaml:"resize"`
}

type ConfImageParameters struct {
	Name    string   `yaml:"name"`
	Alias   []string `yaml:"alias"`
	Format  string   `yaml:"format"`
	Quality int      `yaml:"quality"`
	Width   int      `yaml:"width"`
	Height  int      `yaml:"height"`
	MaxSize string   `yaml:"max_size"`
	Resize  string   `yaml:"resize"`
}

func imgConf(c *Config) images.Config {
	return images.Config{
		OriginalsDir: c.Handler.Paths.Originals,
		CacheDir:     c.Handler.Paths.Cache,
		SetPerms:     c.Handler.Paths.SetPerms,
		CreateDirs:   c.Handler.Paths.CreateDirs,
		DefaultParams: images.ImageParameters{
			Format:  images.MustFormatParse(c.ImageParametersDefault.Format),
			Width:   uint(c.ImageParametersDefault.Width),
			Height:  uint(c.ImageParametersDefault.Height),
			Quality: c.ImageParametersDefault.QualityJpeg,
			MaxSize: images.MustSizeParse(c.ImageParametersDefault.MaxSize),
		},
	}
}
func saveConfig(c Config, filename string) error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, bytes, 0644)
}

func loadConfig(filename string) (Config, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var c Config
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}

// validate enforces config rules and returns an error if any are broken
func (c *Config) validate() error {
	errs := []error{}

	// HTTP
	// port needed
	if c.Server.Port == 0 {
		errs = append(errs, fmt.Errorf("server port must be set"))
	}
	// empty host is ok
	// TODO: validate host format

	// FILES
	if c.Handler.Paths.Originals == "" {
		errs = append(errs, fmt.Errorf("path for originals must be set"))
	}
	if c.Handler.Paths.Cache == "" {
		errs = append(errs, fmt.Errorf("paths for cache must be set"))
	}
	if c.Handler.Cache.MaxNum == 0 {
		errs = append(errs, fmt.Errorf("cache num must be greater than 0"))
	}

	// DEFAULT IMAGE PARAMETERS
	if c.ImageParametersDefault.Format != "jpeg" && c.ImageParametersDefault.Format != "png" && c.ImageParametersDefault.Format != "gif" {
		errs = append(errs, fmt.Errorf("default image parameters format must be set to a valid value. Valid values are: jpeg, png, gif"))
	}
	if c.ImageParametersDefault.QualityJpeg == 0 {
		errs = append(errs, fmt.Errorf("default image parameters quality jpeg must be set to a value greater between 1 and 100 (inclusive)"))
	}
	if c.ImageParametersDefault.QualityGif == 0 {
		errs = append(errs, fmt.Errorf("default image parameters quality gif must be set to a value greater between 1 and 256 (inclusive)"))
	}
	if c.ImageParametersDefault.Width == 0 && c.ImageParametersDefault.Height == 0 {
		errs = append(errs, fmt.Errorf("default image parameters width or height (or both) must be set"))
	}
	// TODO: validate max_size format
	// 0 is ok for max_size, it means no limit
	// TODO: validate resize format

	// IMAGE PARAMETERS
	for _, p := range c.ImageParameters {
		name := p.Name
		if name == "" {
			errs = append(errs, fmt.Errorf("image parameters name must be set"))
		}
		if p.Format != "jpeg" && p.Format != "png" && p.Format != "gif" {
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
		return fmt.Errorf("config validation failed: %v", errs)
	}
	return nil
}
