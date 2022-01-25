package shell

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/musenwill/raftdemo/common"
	"github.com/musenwill/raftdemo/config"
	error2 "github.com/musenwill/raftdemo/http/error"
	"github.com/musenwill/raftdemo/http/types"
	"github.com/musenwill/raftdemo/model"
)

type App struct {
	httpClient *http.Client
	url        url.URL
	execMap    map[string]func(stmt Statement) error
}

func (a *App) Open() error {
	a.execMap = make(map[string]func(stmt Statement) error)
	a.execMap["PING"] = a.Ping
	a.execMap["HELP"] = a.Help
	a.execMap["SHOW NODES"] = a.ShowNodes
	a.execMap["SHOW LEADER"] = a.ShowLeader
	a.execMap["SHOW CONFIG"] = a.ShowConfig
	a.execMap["SET NODE"] = a.SetNodeState
	a.execMap["SET LOG"] = a.SetLogLevel
	a.execMap["SHOW LOGS"] = a.ShowLogs
	a.execMap["WRITE LOG"] = a.WriteLog
	a.execMap["SHOW PIPES"] = a.ShowPipes
	a.execMap["SET PIPE"] = a.SetPipe

	return nil
}

func (a *App) Close() error {
	if a.httpClient != nil {
		a.httpClient.CloseIdleConnections()
	}
	return nil
}

func (a *App) Exec(stmt Statement) error {
	f, ok := a.execMap[stmt.GetName()]
	if !ok {
		return fmt.Errorf("unknown command %s", stmt.GetName())
	}
	return f(stmt)
}

func (a *App) Ping(stmt Statement) error {
	u := a.url
	u.Path = path.Join(u.Path, "ping")

	resp, err := a.httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := a.handleError(resp); err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	ping := types.Ping{}
	if err := dec.Decode(&ping); err != nil {
		return err
	}
	form := common.NewForm()
	form.SetTags([]common.Tag{
		{K: "app name", V: ping.AppName},
		{K: "version", V: ping.Version},
		{K: "branch", V: ping.Branch},
		{K: "commit", V: ping.Commit},
		{K: "build time", V: ping.BuildTime},
		{K: "up time", V: ping.UpTime.String()}})
	fmt.Println(form)

	return nil
}

func (a *App) Help(stmt Statement) error {
	helpers := []Statement{
		&PingStatement{},
		&HelpStatement{},
		&ExitStatement{},
		&ShowNodesStatement{},
		&ShowLeaderStatement{},
		&SetNodeStatement{},
		&ShowConfigStatement{},
		&SetLogLevelStatement{},
		&ShowLogsStatement{},
		&WriteLogStatement{},
		&ShowPipesStatement{},
		&SetPipeStatement{},
	}

	for _, h := range helpers {
		fmt.Println(h.Help())
	}

	return nil
}

func (a *App) ShowNodes(stmt Statement) error {
	showNodesStmt := stmt.(*ShowNodesStatement)
	if showNodesStmt.NodeID == "" {
		return a.showNodes()
	} else {
		return a.showNode(showNodesStmt.NodeID)
	}
}

func (a *App) showNodes() error {
	u := a.url
	u.Path = path.Join(u.Path, "nodes")

	resp, err := a.httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := a.handleError(resp); err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	nodeList := types.ListNodesResponse{}
	if err := dec.Decode(&nodeList); err != nil {
		return err
	}
	form := common.NewForm()
	form.SetTags([]common.Tag{{K: "name", V: "nodes"}, {K: "total", V: fmt.Sprintf("%d", nodeList.Total)}})
	form.SetHeader((&types.Node{}).Header())
	for _, e := range nodeList.Entries {
		form.AddRow(e.Row())
	}
	fmt.Println(form)

	return nil
}

func (a *App) showNode(nodeID string) error {
	u := a.url
	u.Path = path.Join(u.Path, "nodes", nodeID)

	resp, err := a.httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := a.handleError(resp); err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	node := types.Node{}
	if err := dec.Decode(&node); err != nil {
		return err
	}
	fmt.Println(node.Form())

	return nil
}

func (a *App) ShowLeader(stmt Statement) error {
	u := a.url
	u.Path = path.Join(u.Path, "nodes", "leader")

	resp, err := a.httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := a.handleError(resp); err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	node := types.Node{}
	if err := dec.Decode(&node); err != nil {
		return err
	}
	fmt.Println(node.Form())

	return nil
}

func (a *App) SetNodeState(stmt Statement) error {
	setNodeStmt := stmt.(*SetNodeStatement)

	u := a.url
	u.Path = path.Join(u.Path, "nodes", setNodeStmt.NodeID)

	param := struct {
		State string `json:"state"`
	}{State: setNodeStmt.State}

	resp, err := a.post(http.MethodPut, u.String(), param)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return a.handleError(resp)
}

func (a *App) ShowConfig(stmt Statement) error {
	u := a.url
	u.Path = path.Join(u.Path, "config")

	resp, err := a.httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := a.handleError(resp); err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	cfg := config.Config{}
	if err := dec.Decode(&cfg); err != nil {
		return err
	}

	return a.printStruct(cfg)
}

func (a *App) SetLogLevel(stmt Statement) error {
	setLogStmt := stmt.(*SetLogLevelStatement)

	u := a.url
	u.Path = path.Join(u.Path, "config")

	param := struct {
		Level string `json:"level"`
	}{Level: setLogStmt.Level}

	resp, err := a.post(http.MethodPut, u.String(), param)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return a.handleError(resp)
}

func (a *App) ShowLogs(stmt Statement) error {
	showLogsStmt := stmt.(*ShowLogsStatement)
	if showLogsStmt.NodeID == "" {
		showLogsStmt.NodeID = strings.ToLower(LEADER)
	}

	if showLogsStmt.Index == 0 {
		return a.showLogs(showLogsStmt.NodeID)
	} else {
		return a.showLog(showLogsStmt.NodeID, showLogsStmt.Index)
	}
}

func (a *App) showLogs(nodeID string) error {
	u := a.url
	u.Path = path.Join(u.Path, "nodes", nodeID, "logs")

	resp, err := a.httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := a.handleError(resp); err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	entryList := types.ListEntriesResponse{}
	if err := dec.Decode(&entryList); err != nil {
		return err
	}
	form := common.NewForm()
	form.SetTags([]common.Tag{{K: "name", V: "entries"}, {K: "total", V: fmt.Sprintf("%d", entryList.Total)}})
	form.SetHeader((&model.Entry{}).Header())
	for _, e := range entryList.Entries {
		form.AddRow(e.Row())
	}
	fmt.Println(form)

	return nil
}

func (a *App) showLog(nodeID string, index int64) error {
	u := a.url
	u.Path = path.Join(u.Path, "nodes", nodeID, "logs", fmt.Sprintf("%d", index))

	resp, err := a.httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := a.handleError(resp); err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	entry := model.Entry{}
	if err := dec.Decode(&entry); err != nil {
		return err
	}
	fmt.Println(entry.Form())

	return nil
}

func (a *App) WriteLog(stmt Statement) error {
	writeLogStmt := stmt.(*WriteLogStatement)

	u := a.url
	u.Path = path.Join(u.Path, "logs")

	param := struct {
		Data string `json:"data"`
	}{Data: writeLogStmt.Data}

	resp, err := a.post(http.MethodPost, u.String(), param)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return a.handleError(resp)
}

func (a *App) ShowPipes(stmt Statement) error {
	u := a.url
	u.Path = path.Join(u.Path, "proxy")

	resp, err := a.httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := a.handleError(resp); err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()

	pipeList := types.ListPipesResponse{}
	if err := dec.Decode(&pipeList); err != nil {
		return err
	}

	form := common.NewForm()
	form.SetTags([]common.Tag{{K: "name", V: "entries"}, {K: "total", V: fmt.Sprintf("%d", pipeList.Total)}})
	form.SetHeader((&model.Pipe{}).Header())
	for _, e := range pipeList.Entries {
		form.AddRow(e.Row())
	}
	fmt.Println(form)

	return nil
}

func (a *App) SetPipe(stmt Statement) error {
	setPipeStmt := stmt.(*SetPipeStatement)

	u := a.url
	u.Path = path.Join(u.Path, "proxy")

	param := types.UpdatePipesRequest{
		Entries: setPipeStmt.pipes,
	}

	resp, err := a.post(http.MethodPut, u.String(), param)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return a.handleError(resp)

}

func (a *App) handleError(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()

	httpError := error2.HttpError{}
	if err := dec.Decode(&httpError); err != nil {
		return err
	}

	form := common.NewForm()
	form.SetTags([]common.Tag{
		{K: "url", V: resp.Request.URL.String()},
		{K: "status", V: fmt.Sprintf("%d", resp.StatusCode)},
		{K: "code", V: fmt.Sprintf("%d", httpError.Code)},
		{K: "error", V: httpError.Err},
	})

	return errors.New(form.String())
}

func (a *App) post(method, url string, param interface{}) (*http.Response, error) {
	json, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return a.httpClient.Do(req)
}

func (a *App) printStruct(s interface{}) error {
	js, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(js))
	return nil
}
