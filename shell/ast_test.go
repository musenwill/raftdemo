package shell_test

import (
	"testing"

	"github.com/musenwill/raftdemo/shell"
	"github.com/stretchr/testify/require"
)

func TestIsString(t *testing.T) {
	s, ok := shell.IsString("hello")
	require.True(t, ok)
	require.Equal(t, "hello", s)

	s, ok = shell.IsString("hello world")
	require.True(t, ok)
	require.Equal(t, "hello world", s)

	s, ok = shell.IsString(`"hello "`)
	require.True(t, ok)
	require.Equal(t, "hello ", s)
}

func TestIsIdentify(t *testing.T) {
	require.True(t, shell.IsIdentify("hello"))
	require.True(t, shell.IsIdentify("_hello"))
	require.False(t, shell.IsIdentify("SHOW"))
	require.False(t, shell.IsIdentify("show"))
	require.False(t, shell.IsIdentify("0show"))
	require.False(t, shell.IsIdentify(" hello"))
}

func TestScanner(t *testing.T) {
	s := shell.NewScanner("	\n  Set \n node 99 state follower 'hello world' ")

	_, token, err := s.GetToken()
	require.NoError(t, err)
	require.Equal(t, shell.SET, token)

	_, token, err = s.GetToken()
	require.NoError(t, err)
	require.Equal(t, shell.NODE, token)

	_, num, err := s.GetInteger()
	require.NoError(t, err)
	require.Equal(t, int64(99), num)

	_, token, err = s.GetToken()
	require.NoError(t, err)
	require.Equal(t, shell.STATE, token)

	_, follower, err := s.GetString()
	require.NoError(t, err)
	require.Equal(t, "follower", follower)

	_, helloworld, err := s.GetString()
	require.NoError(t, err)
	require.Equal(t, "hello world", helloworld)

	_, eof, err := s.GetEOF()
	require.NoError(t, err)
	require.Equal(t, shell.EOF, eof)
}

func TestParseShowNodes(t *testing.T) {
	stmt, err := shell.ParseStatement("show nodes")
	require.NoError(t, err)
	require.Equal(t, &shell.ShowNodesStatement{}, stmt)

	stmt, err = shell.ParseStatement("SHOW Nodes")
	require.NoError(t, err)
	require.Equal(t, &shell.ShowNodesStatement{}, stmt)

	stmt, err = shell.ParseStatement("show nodes where node = node01")
	require.NoError(t, err)
	require.Equal(t, &shell.ShowNodesStatement{NodeID: "node01"}, stmt)

	stmt, err = shell.ParseStatement("show nodes where node = 'node01'")
	require.NoError(t, err)
	require.Equal(t, &shell.ShowNodesStatement{NodeID: "node01"}, stmt)

	stmt, err = shell.ParseStatement("show nodes where index = 1 and node = 'node01'")
	require.NoError(t, err)
	require.Equal(t, &shell.ShowNodesStatement{NodeID: "node01"}, stmt)

	stmt, err = shell.ParseStatement("show nodes where node = 'node01' and index = 1")
	require.NoError(t, err)
	require.Equal(t, &shell.ShowNodesStatement{NodeID: "node01"}, stmt)

	stmt, err = shell.ParseStatement("show nodes where node = ")
	require.Equal(t, "expect [LEADER STRING] at 22, got EOF", err.Error())

	stmt, err = shell.ParseStatement(`show nodes where node = ""`)
	require.Equal(t, "unexpected empty string at 24", err.Error())

	stmt, err = shell.ParseStatement(`show nodes where node = "show"`)
	require.NoError(t, err)
	require.Equal(t, &shell.ShowNodesStatement{NodeID: "show"}, stmt)
}
