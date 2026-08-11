package cli

import configpkg "github.com/aaronflorey/pupdate/internal/config"

const (
	configFileName        = configpkg.FileName
	projectConfigFileName = configpkg.ProjectFileName
)

type userConfig = configpkg.User
type projectConfig = configpkg.Project
