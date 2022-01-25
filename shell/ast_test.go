package shell

import (
	"testing"

	"github.com/musenwill/raftdemo/model"
	"github.com/stretchr/testify/require"
)

func TestIsString(t *testing.T) {
	s, ok := IsString("hello")
	require.True(t, ok)
	require.Equal(t, "hello", s)

	s, ok = IsString("hello world")
	require.True(t, ok)
	require.Equal(t, "hello world", s)

	s, ok = IsString(`"hello "`)
	require.True(t, ok)
	require.Equal(t, "hello ", s)
}

func TestIsIdentify(t *testing.T) {
	require.True(t, IsIdentify("hello"))
	require.True(t, IsIdentify("_hello"))
	require.False(t, IsIdentify("SHOW"))
	require.False(t, IsIdentify("show"))
	require.False(t, IsIdentify("0show"))
	require.False(t, IsIdentify(" hello"))
}

func TestScanner(t *testing.T) {
	s := NewScanner("	\n  Set \n node 99 state follower 'hello world' ")

	_, token, err := s.GetToken()
	require.NoError(t, err)
	require.Equal(t, SET, token)

	_, token, err = s.GetToken()
	require.NoError(t, err)
	require.Equal(t, NODE, token)

	_, num, err := s.GetInteger()
	require.NoError(t, err)
	require.Equal(t, int64(99), num)

	_, token, err = s.GetToken()
	require.NoError(t, err)
	require.Equal(t, STATE, token)

	_, follower, err := s.GetString()
	require.NoError(t, err)
	require.Equal(t, "follower", follower)

	_, helloworld, err := s.GetString()
	require.NoError(t, err)
	require.Equal(t, "hello world", helloworld)

	_, eof, err := s.GetEOF()
	require.NoError(t, err)
	require.Equal(t, EOF, eof)
}

func TestParseShowNodes(t *testing.T) {
	stmt, err := ParseStatement("show nodes")
	require.NoError(t, err)
	require.Equal(t, &ShowNodesStatement{}, stmt)

	stmt, err = ParseStatement("SHOW Nodes")
	require.NoError(t, err)
	require.Equal(t, &ShowNodesStatement{}, stmt)

	stmt, err = ParseStatement("show nodes where node = node01")
	require.NoError(t, err)
	require.Equal(t, &ShowNodesStatement{NodeID: "node01"}, stmt)

	stmt, err = ParseStatement("show nodes where node = 'node01'")
	require.NoError(t, err)
	require.Equal(t, &ShowNodesStatement{NodeID: "node01"}, stmt)

	stmt, err = ParseStatement("show nodes where index = 1 and node = 'node01'")
	require.NoError(t, err)
	require.Equal(t, &ShowNodesStatement{NodeID: "node01"}, stmt)

	stmt, err = ParseStatement("show nodes where node = 'node01' and index = 1")
	require.NoError(t, err)
	require.Equal(t, &ShowNodesStatement{NodeID: "node01"}, stmt)

	_, err = ParseStatement("show nodes where node = ")
	require.Equal(t, "expect [LEADER STRING] at 22, got EOF", err.Error())

	_, err = ParseStatement(`show nodes where node = ""`)
	require.Equal(t, "unexpected empty string at 24", err.Error())

	stmt, err = ParseStatement(`show nodes where node = "show"`)
	require.NoError(t, err)
	require.Equal(t, &ShowNodesStatement{NodeID: "show"}, stmt)
}

func TestParseSetPipe(t *testing.T) {
	stmt, err := ParseStatement("set pipe from a to b broken")
	require.NoError(t, err)
	require.Equal(t, &SetPipeStatement{[]model.Pipe{{From: "a", To: "b", State: model.StatePipe_broken}}}, stmt)

	stmt, err = ParseStatement("set pipe from a to b broken , from b to a ok")
	require.NoError(t, err)
	require.Equal(t, &SetPipeStatement{[]model.Pipe{{From: "a", To: "b", State: model.StatePipe_broken}, {From: "b", To: "a", State: model.StatePipe_ok}}}, stmt)
}
