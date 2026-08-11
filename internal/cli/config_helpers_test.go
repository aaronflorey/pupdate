package cli

import (
	"os"
	"os/user"
	"runtime"

	configpkg "github.com/aaronflorey/pupdate/internal/config"
	"github.com/aaronflorey/pupdate/internal/userenv"
)

var userConfigDir = os.UserConfigDir
var runtimeGOOS = runtime.GOOS
var userHomeDir = os.UserHomeDir
var currentUserHomeDir = func() (string, error) {
	current, err := user.Current()
	if err != nil {
		return "", err
	}
	return current.HomeDir, nil
}

func currentEnvironment() userenv.Environment {
	environment := userenv.System()
	environment.GOOS = runtimeGOOS
	environment.UserConfigDir = userConfigDir
	environment.UserHomeDir = userHomeDir
	environment.CurrentUserHomeDir = currentUserHomeDir
	return environment
}

func currentConfigLoader() configpkg.Loader {
	return configpkg.Loader{Environment: currentEnvironment()}
}

func resolveUserConfigPath() (string, error) {
	repositoryDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return currentConfigLoader().UserPath(repositoryDir)
}

func resolveUserConfig(value userConfig) (userConfig, error) {
	repositoryDir, err := os.Getwd()
	if err != nil {
		return userConfig{}, err
	}
	return configpkg.ResolveUser(value, repositoryDir, currentEnvironment())
}

func expandConfiguredDirectory(path string) (string, error) {
	repositoryDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return configpkg.ExpandDirectory(path, repositoryDir, currentEnvironment())
}

func resolveDirectory(path string) string {
	return userenv.ResolveDirectory(path)
}

func isTopLevelDirectoryWithinRoot(path string, root string) bool {
	return userenv.IsTopLevelDirectoryWithinRoot(path, root)
}

func sameDirectory(left string, right string) bool {
	return userenv.SameDirectory(left, right)
}
