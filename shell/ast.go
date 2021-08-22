package shell

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	PING   = "PING"
	SHOW   = "SHOW"
	NODES  = "NODES"
	LEADER = "LEADER"
	NODE   = "NODE"
	CONFIG = "CONFIG"
	SET    = "SET"
	STATE  = "STATE"
	LEVEL  = "LEVEL"
	WHERE  = "WHERE"
	LOG    = "LOG"
	LOGS   = "LOGS"
	AND    = "AND"
	INDEX  = "INDEX"
	WRITE  = "WRITE"
	HELP   = "HELP"
	QUIT   = "QUIT"
	EXIT   = "EXIT"

	TOKEN    = "TOKEN"
	IDENTIFY = "IDENTIFY"
	STRING   = "STRING"
	INTEGER  = "INTEGER"
	EOF      = "EOF"
)

var tokenSet = map[string]bool{
	"PING":   true,
	"SHOW":   true,
	"NODES":  true,
	"LEADER": true,
	"NODE":   true,
	"CONFIG": true,
	"SET":    true,
	"STATE":  true,
	"LEVEL":  true,
	"WHERE":  true,
	"LOG":    true,
	"LOGS":   true,
	"AND":    true,
	"INDEX":  true,
	"WRITE":  true,
	"HELP":   true,
	"QUIT":   true,
	"EXIT":   true,
}

func IsToken(word string) bool {
	word = strings.ToUpper(word)
	_, ok := tokenSet[word]
	return ok
}

func IsIdentify(word string) bool {
	if IsToken(word) {
		return false
	}
	if word == EOF {
		return false
	}
	if word == "" {
		return false
	}

	rs := []rune(word)
	if unicode.IsDigit(rs[0]) {
		return false
	}

	for _, r := range rs {
		if !(unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_') {
			return false
		}
	}

	return true
}

func IsInterger(word string) (int64, bool) {
	i, err := strconv.ParseInt(word, 10, 64)
	if err != nil {
		return 0, false
	}
	return i, true
}

func IsString(word string) (string, bool) {
	if IsToken(word) {
		return "", false
	}
	if word == EOF {
		return "", false
	}

	singleQuote := 0
	doubleQuote := 0

	rs := []rune(word)
	for i, r := range rs {
		if r == '\'' {
			if i != 0 && i != len(rs)-1 {
				return "", false
			}
			singleQuote++
		}
		if r == '"' {
			if i != 0 && i != len(rs)-1 {
				return "", false
			}
			doubleQuote++
		}
	}

	if singleQuote != 0 && singleQuote != 2 {
		return "", false
	}
	if doubleQuote != 0 && doubleQuote != 2 {
		return "", false
	}
	if singleQuote > 0 && doubleQuote > 0 {
		return "", false
	}

	if singleQuote+doubleQuote > 0 {
		return string(rs[1 : len(rs)-1]), true
	} else {
		return word, true
	}
}

func IsQuoted(str string) bool {
	if len(str) <= 1 {
		return false
	}
	head := str[0]
	tail := str[len(str)-1]

	if head == tail && (head == '\'' || head == '"') {
		return true
	}
	return false
}

func UnQuote(str string) string {
	if IsQuoted(str) {
		return str[1 : len(str)-1]
	}
	return str
}

type Statement interface {
	GetName() string
	String() string
	Help() string
}

type PingStatement struct {
}

func (s *PingStatement) GetName() string {
	return PING
}

func (s *PingStatement) String() string {
	return s.GetName()
}

func (s *PingStatement) Help() string {
	return s.GetName()
}

type ShowNodesStatement struct {
	NodeID string
}

func (s *ShowNodesStatement) GetName() string {
	return "SHOW NODES"
}

func (s *ShowNodesStatement) String() string {
	if s.NodeID != "" {
		return fmt.Sprintf(`%s WHERE NODE = "%s"`, s.GetName(), s.NodeID)
	} else {
		return s.GetName()
	}
}

func (s *ShowNodesStatement) Help() string {
	return fmt.Sprintf("%s [WHERE NODE = <nodeID>]", s.GetName())
}

type ShowLeaderStatement struct {
}

func (s *ShowLeaderStatement) GetName() string {
	return "SHOW LEADER"
}

func (s *ShowLeaderStatement) String() string {
	return s.GetName()
}

func (s *ShowLeaderStatement) Help() string {
	return s.GetName()
}

type SetNodeStatement struct {
	NodeID string
	State  string
}

func (s *SetNodeStatement) GetName() string {
	return "SET NODE"
}

func (s *SetNodeStatement) String() string {
	return fmt.Sprintf(`%s "%s" %s "%s"`, s.GetName(), s.NodeID, STATE, s.State)
}

func (s *SetNodeStatement) Help() string {
	return fmt.Sprintf("%s <nodeID> STATE <state>", s.GetName())
}

type ShowConfigStatement struct {
}

func (s *ShowConfigStatement) GetName() string {
	return "SHOW CONFIG"
}

func (s *ShowConfigStatement) String() string {
	return s.GetName()
}

func (s *ShowConfigStatement) Help() string {
	return s.GetName()
}

type SetLogLevelStatement struct {
	Level string
}

func (s *SetLogLevelStatement) GetName() string {
	return "SET LOG"
}

func (s *SetLogLevelStatement) String() string {
	return fmt.Sprintf(`%s %s %s`, s.GetName(), LEVEL, s.Level)
}

func (s *SetLogLevelStatement) Help() string {
	return fmt.Sprintf("%s LEVEL <level>", s.GetName())
}

type ShowLogsStatement struct {
	NodeID string
	Index  int64
}

func (s *ShowLogsStatement) GetName() string {
	return "SHOW LOGS"
}

func (s *ShowLogsStatement) String() string {
	str := s.GetName()
	if s.NodeID == "" && s.Index == 0 {
		return str
	}
	str += (" " + WHERE + " ")

	var conditions []string
	if s.NodeID != "" {
		conditions = append(conditions, fmt.Sprintf(`NODE = "%s"`, s.NodeID))
	}
	if s.Index != 0 {
		conditions = append(conditions, fmt.Sprintf(`INDEX = %d`, s.Index))
	}
	return str + strings.Join(conditions, " AND ")
}

func (s *ShowLogsStatement) Help() string {
	return fmt.Sprintf("%s [WHERE NODE = <nodeID> AND INDEX = <index>]", s.GetName())
}

type WriteLogStatement struct {
	Name string
	Data string
}

func (s *WriteLogStatement) GetName() string {
	return "WRITE LOG"
}

func (s *WriteLogStatement) String() string {
	return fmt.Sprintf(`%s "%s"`, s.GetName(), s.Data)
}

func (s *WriteLogStatement) Help() string {
	return fmt.Sprintf("%s <data>", s.GetName())
}

type WhereStatement struct {
	Name      string
	NodeID    string
	Index     int64
	HasNodeID bool
	HasIndex  bool
}

type ExitStatement struct {
}

func (s *ExitStatement) GetName() string {
	return EXIT
}

func (s *ExitStatement) String() string {
	return s.GetName()
}

func (s *ExitStatement) Help() string {
	return "EXIT or QUIT"
}

type HelpStatement struct {
}

func (s *HelpStatement) GetName() string {
	return HELP
}

func (s *HelpStatement) String() string {
	return s.GetName()
}

func (s *HelpStatement) Help() string {
	return s.GetName()
}

type Scanner struct {
	cmd   []rune
	left  int
	right int
}

func NewScanner(cmd string) *Scanner {
	return &Scanner{
		cmd: []rune(cmd),
	}
}

func (s *Scanner) GetToken() (int, string, error) {
	word, err := s.next()
	if err != nil {
		return s.left, "", err
	}
	wordUP := strings.ToUpper(word)
	if !IsToken((wordUP)) {
		return s.left, wordUP, fmt.Errorf("expect token at %d, got %s", s.left, word)
	}
	return s.left, wordUP, nil
}

func (s *Scanner) GetIdentify() (int, string, error) {
	word, err := s.next()
	if err != nil {
		return s.left, "", err
	}
	if !IsIdentify(word) {
		return s.left, word, fmt.Errorf("expect identity at %d, got %s", s.left, word)
	}
	return s.left, word, nil
}

func (s *Scanner) GetInteger() (int, int64, error) {
	word, err := s.next()
	if err != nil {
		return s.left, 0, err
	}
	i, ok := IsInterger(word)
	if !ok {
		return s.left, 0, fmt.Errorf("expect integer at %d, got %s", s.left, word)
	}
	return s.left, i, nil
}

func (s *Scanner) GetString() (int, string, error) {
	word, err := s.next()
	if err != nil {
		return s.left, "", err
	}
	str, ok := IsString(word)
	if !ok {
		return s.left, word, fmt.Errorf("expect string at %d, got %s", s.left, word)
	}

	if len(str) == 0 {
		return s.left, word, fmt.Errorf("unexpected empty string at %d", s.left)
	}

	return s.left, str, nil
}

func (s *Scanner) GetEOF() (int, string, error) {
	word, err := s.next()
	if err != nil {
		return s.left, "", err
	}
	wordUP := strings.ToUpper(word)
	if wordUP != EOF {
		return s.left, wordUP, fmt.Errorf("expect EOF at %d, got %s", s.left, word)
	}
	return s.left, wordUP, nil
}

func (s *Scanner) GetOperator(op string) (int, string, error) {
	word, err := s.next()
	if err != nil {
		return s.left, "", err
	}
	if word != op {
		return s.left, word, fmt.Errorf("expect '%s' at %d, got %s", op, s.left, word)
	}
	return s.left, word, nil
}

func (s *Scanner) Next() (int, string, error) {
	word, err := s.next()
	return s.left, word, err
}

func (s *Scanner) Undo() {
	s.right = s.left
}

func (s *Scanner) next() (string, error) {
	s.passSpace()
	if s.right >= len(s.cmd) {
		return EOF, nil
	}

	type quote struct {
		q   rune
		pos int
	}

	var quoteStack []quote

	s.left = s.right
	start := s.right
	for ; s.right < len(s.cmd); s.right++ {
		if len(quoteStack) == 0 && unicode.IsSpace(s.cmd[s.right]) {
			break
		}
		letter := s.cmd[s.right]
		if letter == '\'' || letter == '"' {
			if len(quoteStack) > 0 {
				if quoteStack[len(quoteStack)-1].q == letter {
					quoteStack = quoteStack[:len(quoteStack)-1]
				} else {
					quoteStack = append(quoteStack, quote{letter, s.right})
				}
			} else {
				quoteStack = append(quoteStack, quote{letter, s.right})
			}
		}
	}
	end := s.right

	if len(quoteStack) > 0 {
		return "", fmt.Errorf("incompletion quote at %d", quoteStack[len(quoteStack)-1].pos)
	}

	word := string(s.cmd[start:end])
	return word, nil
}

func (s *Scanner) passSpace() {
	for ; s.right < len(s.cmd) && unicode.IsSpace(s.cmd[s.right]); s.right++ {
	}
}

func ParseStatements(cmds string) ([]Statement, error) {
	var stmts []Statement

	commands := strings.Split(cmds, "\n")
	for _, cmd := range commands {
		stmt, err := ParseStatement(cmd)
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	return stmts, nil
}

func ParseStatement(cmd string) (Statement, error) {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) > 0 {
		parser := &Parser{scanner: NewScanner(cmd)}
		stmt, err := parser.Parse()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	return nil, nil
}

type Parser struct {
	scanner *Scanner
}

func (p *Parser) Parse() (Statement, error) {
	pos, token, err := p.scanner.GetToken()
	if err != nil {
		return nil, err
	}

	switch token {
	case QUIT:
		return &ExitStatement{}, nil
	case EXIT:
		return &ExitStatement{}, nil
	case PING:
		return &PingStatement{}, nil
	case HELP:
		return &HelpStatement{}, nil
	case SHOW:
		pos, token, err := p.scanner.GetToken()
		if err != nil {
			return nil, err
		}
		switch token {
		case NODES:
			return p.parseShowNodes()
		case LEADER:
			return p.parseShowLeader()
		case CONFIG:
			return p.parseShowConfig()
		case LOGS:
			return p.parseShowLogs()
		default:
			return nil, p.error(pos, []string{NODES, LEADER, CONFIG, LOGS}, token)
		}
	case SET:
		pos, token, err := p.scanner.GetToken()
		if err != nil {
			return nil, err
		}
		switch token {
		case NODE:
			return p.parseSetNode()
		case LOG:
			return p.parseSetLogLevel()
		default:
			return nil, p.error(pos, []string{NODE, LOG}, token)
		}
	case WRITE:
		pos, token, err := p.scanner.GetToken()
		if err != nil {
			return nil, err
		}
		switch token {
		case LOG:
			return p.parseWriteLog()
		default:
			return nil, p.error(pos, []string{LOG}, token)
		}
	default:
		return nil, p.error(pos, []string{SHOW, SET, WRITE, QUIT, EXIT}, token)
	}
}

func (p *Parser) parseShowNodes() (Statement, error) {
	stmt := &ShowNodesStatement{}

	pos, word, err := p.scanner.Next()
	if err != nil {
		return stmt, err
	}
	word = strings.ToUpper(word)

	if word == WHERE {
		s, err := p.parseWhereClause()
		if err != nil {
			return stmt, err
		}
		stmt.NodeID = s.NodeID
		return stmt, nil
	} else if word == EOF {
		return stmt, nil
	} else {
		return stmt, p.error(pos, []string{WHERE, EOF}, word)
	}
}

func (p *Parser) parseShowLeader() (Statement, error) {
	stmt := &ShowLeaderStatement{}

	_, _, err := p.scanner.GetEOF()
	return stmt, err
}

func (p *Parser) parseShowConfig() (Statement, error) {
	stmt := &ShowConfigStatement{}

	_, _, err := p.scanner.GetEOF()
	return stmt, err
}

func (p *Parser) parseShowLogs() (Statement, error) {
	stmt := &ShowLogsStatement{}

	pos, word, err := p.scanner.Next()
	if err != nil {
		return stmt, err
	}
	word = strings.ToUpper(word)

	if word == WHERE {
		s, err := p.parseWhereClause()
		if err != nil {
			return stmt, err
		}
		stmt.NodeID = s.NodeID
		stmt.Index = s.Index
		return stmt, nil
	} else if word == EOF {
		return stmt, nil
	} else {
		return stmt, p.error(pos, []string{WHERE, EOF}, word)
	}
}

func (p *Parser) parseSetNode() (Statement, error) {
	stmt := &SetNodeStatement{}

	_, nodeID, err := p.scanner.GetString()
	if err != nil {
		return stmt, err
	}
	stmt.NodeID = nodeID

	pos, token, err := p.scanner.GetToken()
	if token != STATE {
		return stmt, p.error(pos, []string{STATE}, token)
	}

	_, state, err := p.scanner.GetString()
	if err != nil {
		return stmt, err
	}
	stmt.State = state

	_, _, err = p.scanner.GetEOF()
	return stmt, err
}

func (p *Parser) parseSetLogLevel() (Statement, error) {
	stmt := &SetLogLevelStatement{}

	pos, token, err := p.scanner.GetToken()
	if token != LEVEL {
		return stmt, p.error(pos, []string{LEVEL}, token)
	}

	_, level, err := p.scanner.GetString()
	if err != nil {
		return stmt, err
	}
	stmt.Level = level

	_, _, err = p.scanner.GetEOF()
	return stmt, err
}

func (p *Parser) parseWriteLog() (Statement, error) {
	stmt := &WriteLogStatement{Name: "WRITE LOG"}

	_, data, err := p.scanner.GetString()
	if err != nil {
		return stmt, err
	}
	stmt.Data = data

	_, _, err = p.scanner.GetEOF()
	return stmt, err
}

func (p *Parser) error(pos int, expect []string, got string) error {
	return fmt.Errorf("expect %+v at %d, got %s", expect, pos, got)
}

func (p *Parser) parseWhereClause() (WhereStatement, error) {
	stmt := WhereStatement{Name: "WHERE"}

	pos, token, err := p.scanner.GetToken()
	if err != nil {
		return stmt, err
	}
	{
		_, _, err := p.scanner.GetOperator("=")
		if err != nil {
			return stmt, err
		}
	}
	switch token {
	case NODE:
		pos, word, err := p.scanner.Next()
		if err != nil {
			return stmt, nil
		}
		if strings.ToUpper(word) == LEADER {
			stmt.NodeID = "leader"
			stmt.HasNodeID = true
		} else if nodeID, ok := IsString(word); ok {
			if len(nodeID) == 0 {
				return stmt, fmt.Errorf("unexpected empty string at %d", pos)
			}

			stmt.NodeID = nodeID
			stmt.HasNodeID = true
		} else {
			return stmt, p.error(pos, []string{LEADER, STRING}, word)
		}
	case INDEX:
		_, index, err := p.scanner.GetInteger()
		if err != nil {
			return stmt, err
		}
		stmt.Index = index
		stmt.HasIndex = true
	default:
		return stmt, p.error(pos, []string{NODE, INDEX}, token)
	}

	pos, word, err := p.scanner.Next()
	if err != nil {
		return stmt, err
	}
	word = strings.ToUpper(word)

	if IsToken(word) && word == AND {
		s, err := p.parseWhereClause()
		if err != nil {
			return stmt, err
		}
		if stmt.HasNodeID && s.HasNodeID {
			return stmt, fmt.Errorf("multi %s in where clause", NODE)
		}
		if stmt.HasIndex && s.HasIndex {
			return stmt, fmt.Errorf("multi %s in where clause", INDEX)
		}
		if !stmt.HasNodeID {
			stmt.NodeID = s.NodeID
		}
		if !stmt.HasIndex {
			stmt.Index = s.Index
		}
		return stmt, nil
	} else if word == EOF {
		return stmt, nil
	} else {
		return stmt, p.error(pos, []string{AND, EOF}, word)
	}
}
