package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	content := `database:
  address: "postgres://postgres:test@localhost:5432/todo"
server:
  port: 9000
  debug: false
metrics:
  tracingEnable: true`

	f, err := ioutil.TempFile("", "config")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	n, err := f.WriteString(content)
	require.NoError(t, err)
	require.Equal(t, len(content), n)

	c, err := ReadConfigFile(f.Name())
	require.NoError(t, err)

	// then
	require.Equal(t, "postgres://postgres:test@localhost:5432/todo", c.Database.Address)
	require.Equal(t, false, c.Server.Debug)
	require.Equal(t, 9000, c.Server.Port)
	require.Equal(t, true, c.Metrics.TracingEnable)
}

func TestNoFile(t *testing.T) {
	_, err := ReadConfigFile("somerandomfile.txt")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read config file")
}

func TestInvalidYaml(t *testing.T) {
	content := `
database:
	address: "postgres://postgres:test@localhost:15432/todo"`
	f, err := ioutil.TempFile("", "config")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	n, err := f.WriteString(content)
	require.NoError(t, err)
	require.Equal(t, len(content), n)

	_, err = ReadConfigFile(f.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "yaml parse failed")
}
