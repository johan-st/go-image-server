package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// CONFIG STUFF
type Config struct {
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
	Paths ConfHandlerPaths `yaml:"paths"`
	Cache ConfHandlerCache `yaml:"cache"`
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
