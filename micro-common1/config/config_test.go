package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestConfig(t *testing.T) {
	assert.NotNil(t, Data.Project.Name)
}

func TestNormal(t *testing.T) {
	a:=4000000/30/12
	fmt.Println(a)

	fmt.Println(path.Ext("asdf"))
}

