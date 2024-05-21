package lexer

import "compiler-go/token"

type Lexer struct {
	input        string
	position     int  // 입력에서 현재 위치 (현재 문자의 주소)
	readPosition int  // 입력에서 현재 읽는 위치 (다음 문자의 주소)
	ch           byte // 현자 조사하는 문자 (position에 해당하는 문자)
}

// New 생성자
func New(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.readChar() // position, readPosition, char 초기화
	return lexer
}

func (lexer *Lexer) readChar() {
	if lexer.readPosition >= len(lexer.input) {
		lexer.ch = 0
	} else {
		lexer.ch = lexer.input[lexer.readPosition]
	}
	lexer.position = lexer.readPosition
	lexer.readPosition += 1
}

func (lexer *Lexer) NextToken() token.Token {
	var tok token.Token

	lexer.skipWhiteSpace()

	switch lexer.ch {
	case '=':
		if lexer.peekChar() == '=' {
			ch := lexer.ch
			lexer.readChar()
			literal := string(ch) + string(lexer.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, lexer.ch)
		}
	case '+':
		tok = newToken(token.PLUS, lexer.ch)
	case '-':
		tok = newToken(token.MINUS, lexer.ch)
	case '!':
		if lexer.peekChar() == '=' {
			ch := lexer.ch
			lexer.readChar()
			literal := string(ch) + string(lexer.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, lexer.ch)
		}
	case '/':
		tok = newToken(token.SLASH, lexer.ch)
	case '*':
		tok = newToken(token.ASTERISK, lexer.ch)
	case '<':
		tok = newToken(token.LT, lexer.ch)
	case '>':
		tok = newToken(token.GT, lexer.ch)
	case ';':
		tok = newToken(token.SEMICOLON, lexer.ch)
	case ',':
		tok = newToken(token.COMMA, lexer.ch)
	case '(':
		tok = newToken(token.LPAREN, lexer.ch)
	case ')':
		tok = newToken(token.RPAREN, lexer.ch)
	case '{':
		tok = newToken(token.LBRACE, lexer.ch)
	case '}':
		tok = newToken(token.RBRACE, lexer.ch)
	case '[':
		tok = newToken(token.LBRACKET, lexer.ch)
	case ']':
		tok = newToken(token.RBRACKET, lexer.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = lexer.readString()
	case ':':
		tok = newToken(token.COLON, lexer.ch)
	case 0:
		tok.Type = token.EOF
		tok.Literal = ""
	default:
		if isLetter(lexer.ch) {
			tok.Literal = lexer.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(lexer.ch) {
			tok.Type = token.INT
			tok.Literal = lexer.readNumber()
			return tok
		}
		tok = newToken(token.ILLEGAL, lexer.ch)
	}

	lexer.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// 변수명, 예약어 파싱
func (lexer *Lexer) readIdentifier() string {
	position := lexer.position
	for isLetter(lexer.ch) {
		lexer.readChar()
	}
	return lexer.input[position:lexer.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// 공백 및 줄바꿈 스킵
func (lexer *Lexer) skipWhiteSpace() {
	for lexer.ch == ' ' || lexer.ch == '\t' || lexer.ch == '\n' || lexer.ch == '\r' {
		lexer.readChar()
	}
}

// 정수 파싱 (실수는 미지원)
func (lexer *Lexer) readNumber() string {
	position := lexer.position
	for isDigit(lexer.ch) {
		lexer.readChar()
	}
	return lexer.input[position:lexer.position]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// 다음에 나올 입력을 살펴봄(=peek)
func (lexer *Lexer) peekChar() byte {
	if lexer.readPosition >= len(lexer.input) {
		return 0
	}

	return lexer.input[lexer.readPosition]
}

func (lexer *Lexer) readString() string {
	// 따옴표 다음 문자
	start := lexer.readPosition

	// 큰따옴표 혹은 입력의 끝까지 문자열을 계속 읽음
	for {
		lexer.readChar()
		if lexer.ch == '"' || lexer.ch == 0 {
			break
		}
	}

	return lexer.input[start:lexer.position]
}
