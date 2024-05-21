package parser

import (
	"fmt"
	"interpreter-go/ast"
	"interpreter-go/lexer"
	"interpreter-go/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // == 또는 !=
	LESSGREATER // > 또는 <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X 또는 !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

// 연산자들의 우선순위 지정
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	// 파라미터로 좌측 피연산자를 받음
	infixParseFn func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	currentToken token.Token
	peekToken    token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	// 렉서를 읽어서 peekToken 세팅 (curToken은 최초 peekToken이 nil이라서 비어있음)
	p.nextToken()
	// 렉서를 한번 더 읽어서 curToken, peekToken 둘 다 세팅
	p.nextToken()

	// 전위표현식 파싱 함수 추가
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	// bool 표현식 파싱 함수 추가
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)

	// 그룹표현식(소괄호) 파싱 함수 추가
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	// if문 파싱 함수 추가
	p.registerPrefix(token.IF, p.parseIfExpression)

	// function 파싱 함수 추가
	p.registerPrefix(token.FUNCTION, p.parseFunctionExpression)

	// String 파싱 함수 추가
	p.registerPrefix(token.STRING, p.parseStringLiteral)

	// Array 파싱 함수 추가
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)

	// 해시 파싱 함수 추가
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	// 중위표현식 파싱 함수 추가
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	// 함수 호출 표현식 파싱 함수 추가
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	// Array 인덱스 파싱 함수
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	return p
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.currentToken}

	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as interger", p.currentToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.currentToken, Value: p.currentTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	expression := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return expression
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.currentToken}

	// if문 다음에는 '('가 나와야함
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// '(' 다음에 조건이 나오니까 nextToken 호출
	p.nextToken()

	// 조건 파싱
	expression.Condition = p.parseExpression(LOWEST)

	// 조건이 끝나면 ')'가 나와야함
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// ')' 다음에는 '{'가 나와야함
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// '{' 이후에는 if문 true일 때의 구문이 나와야함
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.currentToken}
	block.Statements = []ast.Statement{}

	// '{'에서 다음 토큰으로 진행
	p.nextToken()

	// '{' 이후에는 '}'를 만날 때 까지 if문이 true일 때의 구문임
	for !p.currentTokenIs(token.RBRACE) && !p.currentTokenIs(token.EOF) {
		// for문을 통해 구문별로 파싱
		statement := p.parseStatement()
		if statement != nil {
			block.Statements = append(block.Statements, statement)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionExpression() ast.Expression {

	lit := &ast.FunctionLiteral{Token: p.currentToken}

	// 여는 소괄호(=파라미터 시작지점) 이 아니면 리턴
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// prefix가 FUNCTION 토큰일때 소괄호까지 토큰 진행 후 호출됨
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// 파라미터가 비어있으면 여는 소괄호 바로 뒤에 닫는 소괄호가 나올 수 있음
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken() // 닫힌 소괄호 스킵
		return identifiers
	}

	// 소괄호 내부의 파라미터부터 토큰 진행하기 위한 호출
	p.nextToken()

	// 첫 파라미터 호출 후 저장
	identifier := &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
	identifiers = append(identifiers, identifier)

	// 두번 째 파라미터가 있는지 콤마로 식별
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // 호출을 통해 토큰을 콤마로 진행
		p.nextToken() // 호출을 통해 토큰을 다음 파라미터로 진행
		identifier := &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
		identifiers = append(identifiers, identifier)
	}

	// 마지막 파라미터 수집 후 닫는 소괄호가 안나오면 잘못된 문법이므로 nil 리턴
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.currentToken,
		Left:     left,
		Operator: p.currentToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	expression := &ast.CallExpression{Token: p.currentToken, Function: function}
	expression.Arguments = p.parseExpressionList(token.RPAREN)
	return expression
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("Expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.currentToken.Type != token.EOF {
		statement := p.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		// parseStatement에 선언된 문법이 아니면 토큰 스킵
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// Let으로 변수할당하는 코드 한줄을 받으면 호출
func (p *Parser) parseLetStatement() *ast.LetStatement {
	// LetStatement 인스턴스 생성
	statement := &ast.LetStatement{Token: p.currentToken}

	// expectPeek으로 다음 토큰이 IDENT(=변수명)인지 확인
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// 변수명이 맞으면 Name을 Identifier로 초기화
	statement.Name = &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	statement.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: p.currentToken}

	p.nextToken()

	statement.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: p.currentToken}
	statement.Expression = p.parseExpression(LOWEST) // ")"를 만나면 parseExpression 루프 종료하면서 리턴

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	// 현재 토큰이 전위표현식에 해당하면 전위 파싱함수 호출
	prefixFunc := p.prefixParseFns[p.currentToken.Type]
	if prefixFunc == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}
	leftExpression := prefixFunc()

	// 문장별로 처음 호출되는 parseExpression은 precedence를 LOWEST로 받기 때문에 먼저 계산되어야 할 Expression까지 계속 루프를 돌게 됨
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infixFunc := p.infixParseFns[p.peekToken.Type]
		if infixFunc == nil {
			return leftExpression
		}

		p.nextToken()

		// 1. infixFunc으로 사용되는 parseInfixExpression에서 parseExpression 함수에 중위연산자 우선순위를 precedence로 전달하여 재귀호출하고 있음
		// 2. 우선순위가 낮은 연산이 나올때까지 계속해서 parseExpression을 재귀호출함
		// 3. parseInfixExpression와 같은 함수에서 우선순위에 의해 재귀가 끝난 데이터는 Right에 담아놓음
		// 3. 결과적으로 우선순위를 기준으로 Right 노드에 재귀적으로 Expression 노드를 담아놓는 형태가 됨
		leftExpression = infixFunc(leftExpression)
	}

	return leftExpression
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if !p.peekTokenIs(t) {
		p.peekError(t)
		return false
	}
	p.nextToken()
	return true
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.currentToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.currentToken}

	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // 호출 후 현재 토큰 : 콤마
		p.nextToken() // 호출 후 현재 토큰 : 다음 element
		list = append(list, p.parseExpression(LOWEST))
	}

	// 함수 호출식이 마지막에 end로 닫히는지 확인 ex_ ), ], ...
	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.currentToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.currentToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}
	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hash
}
