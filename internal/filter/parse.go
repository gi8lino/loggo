package filter

import (
	"fmt"
	"strings"
)

// ParseExpression parses a human-friendly filter expression.
func ParseExpression(expr string) (Matcher, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("filter expression must not be empty")
	}

	if matcher, ok, err := parseRelativeExpression(expr); ok || err != nil {
		return matcher, err
	}

	parser, err := newExpressionParser(expr)
	if err != nil {
		return nil, err
	}

	matcher, err := parser.parse()
	if err != nil {
		return nil, err
	}

	return matcher, nil
}

// newMatcher creates a matcher from parsed expression parts.
func newMatcher(field string, op string, value string, raw string) (Matcher, error) {
	field = strings.TrimSpace(field)
	op = normalizeOp(op)
	value = strings.TrimSpace(value)

	if isTimeField(field) || op == "after" || op == "before" {
		return newTimeMatcher(field, op, value, raw)
	}

	switch op {
	case "", "contains":
		return containsMatcher{field: field, value: value, raw: raw}, nil
	case "wildcard":
		return newWildcardMatcher(field, value, raw)
	case "regex":
		return newRegexMatcher(field, value, raw)
	case "eq":
		return equalsMatcher{field: field, value: value, raw: raw}, nil
	case "neq":
		return notMatcher{matcher: equalsMatcher{field: field, value: value, raw: raw}, raw: raw}, nil
	case "gt", "gte", "lt", "lte":
		return newNumberMatcher(field, op, value, raw)
	case "exists":
		return existsMatcher{field: field, raw: raw}, nil
	case "in":
		return inMatcher{field: field, values: splitValues(value), raw: raw}, nil
	default:
		return nil, fmt.Errorf("unknown filter operator %q", op)
	}
}

// normalizeOp normalizes symbolic and word operators.
func normalizeOp(op string) string {
	switch strings.ToLower(strings.TrimSpace(op)) {
	case "=":
		return "eq"
	case "==":
		return "eq"
	case "equals":
		return "eq"
	case "eq":
		return "eq"
	case "!=":
		return "neq"
	case "not_equals":
		return "neq"
	case "neq":
		return "neq"
	case ">":
		return "gt"
	case ">=":
		return "gte"
	case "<":
		return "lt"
	case "<=":
		return "lte"
	case "contains":
		return "contains"
	case "wildcard", "glob":
		return "wildcard"
	case "regex":
		return "regex"
	case "exists":
		return "exists"
	case "in":
		return "in"
	case "after":
		return "after"
	case "before":
		return "before"
	default:
		return strings.ToLower(strings.TrimSpace(op))
	}
}

// splitValues splits a comma-separated value list.
func splitValues(value string) []string {
	values := []string{}

	for item := range strings.SplitSeq(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		values = append(values, item)
	}

	return values
}

type expressionParser struct {
	raw    string
	tokens []exprToken
	index  int
}

type exprToken struct {
	kind string
	text string
}

func newExpressionParser(raw string) (*expressionParser, error) {
	tokens, err := tokenizeExpression(raw)
	if err != nil {
		return nil, err
	}

	return &expressionParser{
		raw:    raw,
		tokens: tokens,
	}, nil
}

func (p *expressionParser) parse() (Matcher, error) {
	matcher, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	if !p.done() {
		return nil, fmt.Errorf("unexpected token %q", p.peek().text)
	}

	return matcher, nil
}

func (p *expressionParser) parseOr() (Matcher, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.matchWord("or") {
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}

		left = anyMatcher{matchers: []Matcher{left, right}, raw: p.raw}
	}

	return left, nil
}

func (p *expressionParser) parseAnd() (Matcher, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.matchWord("and") {
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}

		left = allMatcher{matchers: []Matcher{left, right}, raw: p.raw}
	}

	return left, nil
}

func (p *expressionParser) parseUnary() (Matcher, error) {
	if p.matchWord("not") {
		nested, err := p.parseUnary()
		if err != nil {
			return nil, err
		}

		return notMatcher{matcher: nested, raw: p.raw}, nil
	}

	return p.parsePrimary()
}

func (p *expressionParser) parsePrimary() (Matcher, error) {
	if p.matchSymbol("(") {
		matcher, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if !p.matchSymbol(")") {
			return nil, fmt.Errorf("missing closing parenthesis")
		}

		return matcher, nil
	}

	return p.parseSimple()
}

func (p *expressionParser) parseSimple() (Matcher, error) {
	tokens := []exprToken{}

	for !p.done() {
		token := p.peek()
		if token.text == ")" || tokenIsWord(token, "and") || tokenIsWord(token, "or") {
			break
		}

		tokens = append(tokens, token)
		p.index++
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("expected filter expression")
	}

	return parseSimpleTokens(tokens, p.raw)
}

func parseSimpleTokens(tokens []exprToken, raw string) (Matcher, error) {
	if len(tokens) == 2 && tokenIsWord(tokens[0], "after") {
		return newTimeMatcher("time", "gte", tokens[1].text, raw)
	}
	if len(tokens) == 2 && tokenIsWord(tokens[0], "before") {
		return newTimeMatcher("time", "lte", tokens[1].text, raw)
	}
	if len(tokens) == 2 && tokenIsWord(tokens[1], "exists") {
		return newMatcher(tokens[0].text, "exists", "", raw)
	}

	if len(tokens) >= 3 {
		field := tokens[0].text
		op := tokens[1].text
		value := joinTokenText(tokens[2:])

		if op == ":" {
			op = "contains"
		}

		return newMatcher(field, op, value, raw)
	}

	if len(tokens) == 1 {
		return newMatcher("raw", "contains", tokens[0].text, raw)
	}

	return nil, fmt.Errorf("invalid filter expression %q", joinTokenText(tokens))
}

func (p *expressionParser) done() bool {
	return p.index >= len(p.tokens)
}

func (p *expressionParser) peek() exprToken {
	if p.done() {
		return exprToken{}
	}

	return p.tokens[p.index]
}

func (p *expressionParser) matchWord(expected string) bool {
	if p.done() || !tokenIsWord(p.tokens[p.index], expected) {
		return false
	}

	p.index++

	return true
}

func (p *expressionParser) matchSymbol(expected string) bool {
	if p.done() || p.tokens[p.index].text != expected {
		return false
	}

	p.index++

	return true
}

func tokenizeExpression(expr string) ([]exprToken, error) {
	tokens := []exprToken{}

	for index := 0; index < len(expr); {
		switch expr[index] {
		case ' ', '\t', '\n', '\r':
			index++
		case '(', ')':
			tokens = append(tokens, exprToken{kind: "symbol", text: expr[index : index+1]})
			index++
		case ':':
			tokens = append(tokens, exprToken{kind: "op", text: ":"})
			index++
		case '>', '<', '!', '=':
			end := index + 1
			if end < len(expr) && expr[end] == '=' {
				end++
			}

			tokens = append(tokens, exprToken{kind: "op", text: expr[index:end]})
			index = end
		case '"', '\'':
			value, consumed, err := readQuotedToken(expr[index:])
			if err != nil {
				return nil, err
			}

			tokens = append(tokens, exprToken{kind: "word", text: value})
			index += consumed
		default:
			end := index
			for end < len(expr) {
				switch expr[end] {
				case ' ', '\t', '\n', '\r', '(', ')', ':', '>', '<', '!', '=':
					goto done
				}
				end++
			}

		done:
			tokens = append(tokens, exprToken{kind: "word", text: expr[index:end]})
			index = end
		}
	}

	return tokens, nil
}

func readQuotedToken(input string) (string, int, error) {
	quote := input[0]
	escaped := false

	for index := 1; index < len(input); index++ {
		switch {
		case escaped:
			escaped = false
		case input[index] == '\\':
			escaped = true
		case input[index] == quote:
			raw := input[1:index]
			raw = strings.ReplaceAll(raw, `\"`, `"`)
			raw = strings.ReplaceAll(raw, `\'`, `'`)
			raw = strings.ReplaceAll(raw, `\\`, `\`)

			return raw, index + 1, nil
		}
	}

	return "", 0, fmt.Errorf("unterminated quoted value")
}

func tokenIsWord(token exprToken, expected string) bool {
	return token.kind == "word" && strings.EqualFold(token.text, expected)
}

func joinTokenText(tokens []exprToken) string {
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		parts = append(parts, token.text)
	}

	return strings.Join(parts, " ")
}
