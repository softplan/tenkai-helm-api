package handlers

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Environment struct {
	Name          string
	Token         string
	Filename      string
	CACertificate string
	ClusterURI    string
	Namespace     string
}

func getEnvironmentData() Environment {
	return Environment{
		Name:          "Test",
		Token:         "kubeconfig-xpto",
		Filename:      "test",
		CACertificate: "xpto",
		ClusterURI:    "mycluster.com",
		Namespace:     "test",
	}
}

func createTestFile(filename string) {
	err := ioutil.WriteFile(filename, []byte("Hello"), 0755)
	if err != nil {
		panic("Unable to create file " + filename)
	}
}

func TestCreateEnvironmentFile(t *testing.T) {
	env := getEnvironmentData()
	createTestFile(env.Filename)
	err := removeEnvironmentFile(env.Filename)
	assert.Nil(t, err, "Error deleting file")
}

func TestCreateEnvironmentFileTestOk(t *testing.T) {
	env := getEnvironmentData()
	createTestFile(env.Filename)
	createEnvironmentFile(env.Name, env.Token, env.Filename, env.CACertificate, env.ClusterURI, env.Namespace)
	err := removeEnvironmentFile(env.Filename)
	assert.Nil(t, err, "Error deleting file")
}
