package code

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCode(t *testing.T) {
	fd, err := os.CreateTemp("./", "")
	require.NoError(t, err)
	defer fd.Close()
	defer os.Remove(fd.Name())

	_, err = fd.Write([]byte(goFile1))
	require.NoError(t, err)

	c, err := NewCode(fd.Name(), "Two")
	require.NoError(t, err)

	require.Equal(t, c.CodeContent, goFile1)
	require.Equal(t, c.StartOfFuncLine, 9)
	require.Equal(t, c.EndOfFuncLine, 15)
	require.Equal(t, c.FuncName, "Two")
	require.Equal(t, c.Path, fd.Name())
}

func TestReplaceFuncBody(t *testing.T) {
	fd, err := os.CreateTemp("./", "")
	require.NoError(t, err)
	defer fd.Close()
	defer os.Remove(fd.Name())

	_, err = fd.Write([]byte(goFile1))
	require.NoError(t, err)

	c := new(Code)

	c.Path = fd.Name()
	c.CodeContent = goFile1
	c.StartOfFuncLine = 9
	c.EndOfFuncLine = 15

	require.NoError(t, c.ReplaceFuncBody("\treturn nil"), "can not replcae")

	content, err := os.ReadFile(fd.Name())
	require.NoError(t, err)

	require.Equal(t, string(content), goFile2)
}

var (
	goFile1 = `package sample

import "fmt"

func One() error {
	return nil
}

func Two() error {
	fmt.Println(1)
	fmt.Println(2)
	fmt.Println(3)
	fmt.Println(4)
	return nil
}
`
	goFile2 = `package sample

import "fmt"

func One() error {
	return nil
}

func Two() error {
	return nil
}
`
)
