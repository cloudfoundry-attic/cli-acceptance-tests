package integration

import (
	"fmt"
	"os"
	"strings"
)

func addOrReplaceEnvironment(newEnvName string, newEnvVal string) []string {
	var found bool
	env := os.Environ()
	for i, envPair := range env {
		splitENV := strings.Split(envPair, "=")
		if splitENV[0] == newEnvName {
			env[i] = fmt.Sprintf("%s=%s", newEnvName, newEnvVal)
			found = true
		}
	}

	if !found {
		env = append(env, fmt.Sprintf("%s=%s", newEnvName, newEnvVal))
	}
	return env
}
