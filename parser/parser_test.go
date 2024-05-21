package parser

import (
	"fmt"
	"interpreter-go/ast"
	"interpreter-go/lexer"
	"testing"
)

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testLetStatements(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %s", msg)
	}
	t.FailNow()
}

func testLetStatements(t *testing.T, s ast.Statement, expectedName string) bool {
	// let을 가진 토큰인지 체크
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got %s", s.TokenLiteral())
		return false
	}

	letStatement, ok := s.(*ast.LetStatement)

	// statement가 LetStatement 타입으로 래핑되도록 제대로 구현되었는지 체크
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	// letStatement의 변수명이 테스트케이스와 일치하는지 체크
	if letStatement.Name.Value != expectedName {
		t.Errorf("letStatement.Name.Value not '%s'. got %s", expectedName, letStatement.Name.Value)
		return false
	}

	// 토큰 리터럴 또한 테스트케이스와 일치하는지 체크
	if letStatement.Name.TokenLiteral() != expectedName {
		t.Errorf("letStatement.Name.TokenLiteral() not '%s'. got %s", expectedName, letStatement.Name.TokenLiteral())
		return false
	}

	return true
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return true;", true},
		{"return foobar;", "foobar"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}

		stmt := program.Statements[0]

		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("stmt not *ast.returnStatement. got=%T", stmt)
		}

		if returnStmt.TokenLiteral() != "return" {
			t.Fatalf("returnStmt.TokenLiteral not 'return', got %q",
				returnStmt.TokenLiteral())
		}

		if testLiteralExpression(t, returnStmt.ReturnValue, tt.expectedValue) {
			return
		}
	}
}

func TestIdentifierStatements(t *testing.T) {
	input := "foobar"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	// input이 표현식문이 맞는지 체크
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	identifier, ok := statement.Expression.(*ast.Identifier)
	// 표현식문의 표현식이 식별자가 맞는지 체크
	if !ok {
		t.Fatalf("expression is not *ast.Identifier. got=%T", statement.Expression)
	}
	// 식별자의 value가 foobar가 맞는지 체크
	if identifier.Value != "foobar" {
		t.Errorf("identifier.Value not '%s'. got %s", "foobar", identifier.Value)
	}
	// 식별자의 토큰 리터럴이 foobar가 맞는지 체크
	if identifier.TokenLiteral() != "foobar" {
		t.Errorf("identifier.TokenLiteral not %s. got %s", "foobar", identifier.TokenLiteral())
	}

}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := statement.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expression is not *ast.IntegerLiteral. got=%T", statement.Expression)
	}
	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral not %ss. got=%s", "5", literal.TokenLiteral())
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, prefixTest := range prefixTests {
		l := lexer.New(prefixTest.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statemens. got=%d\n", 1, len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		exp, ok := statement.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("statement is not ast.PrefixExpression. got=%T", statement.Expression)
		}
		if exp.Operator != prefixTest.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", prefixTest.operator, exp.Operator)
		}
		if !testLiteralExpression(t, exp.Right, prefixTest.value) {
			return
		}
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integerLiteral, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntergerLiteral. got=%T", il)
		return false
	}

	if integerLiteral.Value != value {
		t.Errorf("integerLiteral.Value not %d. got=%d", value, integerLiteral.Value)
		return false
	}

	if integerLiteral.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integerLiteral.TokenLiteral not %d. got=%s", value, integerLiteral.TokenLiteral())
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	booleanExpression, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp is not ast.Boolean. got=%T", exp)
	}

	if booleanExpression.Value != value {
		t.Errorf("booleanExpression.Value not %t. got=%t", value, booleanExpression.Value)
		return false
	}

	if booleanExpression.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("booleanExpression.TokenLiteral not %t. got=%s", value, booleanExpression.TokenLiteral())
		return false
	}

	return true
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, infixTest := range infixTests {
		l := lexer.New(infixTest.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		if !testInfixExpression(t, statement.Expression, infixTest.leftValue, infixTest.operator, infixTest.rightValue) {
			return
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != test.expected {
			t.Errorf("expected=%q, got=%q", test.expected, actual)
		}
	}
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	identifier, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if identifier.Value != value {
		t.Errorf("identifier.Value not %s. got=%s", value, identifier.Value)
		return false
	}

	if identifier.TokenLiteral() != value {
		t.Errorf("identifier.TokenLiteral not %s. got=%s", value, identifier.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input           string
		expectedBoolean bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has not enough statements. got=%d",
				len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		boolExpression, ok := statement.Expression.(*ast.Boolean)
		if !ok {
			t.Fatalf("boolExpression is not ast.Boolean. got=%T", statement.Expression)
		}

		if boolExpression.Value != test.expectedBoolean {
			t.Errorf("boolExpression.Value is not %t. got=%t", test.expectedBoolean, boolExpression.Value)
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ifExpression, ok := statement.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("statement.Expression is not ast.IfExpression. got=%T", statement.Expression)
	}

	if !testInfixExpression(t, ifExpression.Condition, "x", "<", "y") {
		return
	}

	if len(ifExpression.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n", len(ifExpression.Consequence.Statements))
	}

	consequence, ok := ifExpression.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T", ifExpression.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if ifExpression.Alternative != nil {
		t.Errorf("ifExpression.Alternative was not nil. got=%+v", ifExpression.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n",
			len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if len(exp.Alternative.Statements) != 1 {
		t.Errorf("exp.Alternative.Statements does not contain 1 statements. got=%d\n",
			len(exp.Alternative.Statements))
	}

	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Alternative.Statements[0])
	}

	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFuncionLiteralParsing(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T",
			stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d\n",
			len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements has not 1 statements. got=%d\n",
			len(function.Body.Statements))
	}

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt is not ast.ExpressionStatement. got=%T",
			function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		statement := program.Statements[0].(*ast.ExpressionStatement)
		function := statement.Expression.(*ast.FunctionLiteral)

		if len(function.Parameters) != len(test.expectedParams) {
			t.Errorf("length parameters wrong. want %d, got%d\n", len(test.expectedParams), len(function.Parameters))
		}

		for i, identifier := range test.expectedParams {
			testLiteralExpression(t, function.Parameters[i], identifier)
		}
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
			stmt.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d",
			len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestStringLiteral(t *testing.T) {
	input := `"hello world";`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := statement.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expression not *ast.StringLiteral. got=%T", statement.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestParsingArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := statement.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expression not *ast.ArrayLiteral. got=%T", statement.Expression)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("len(array.Elements) not 3. got=%d", len(array.Elements))
	}

	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestParsingIndexExpression(t *testing.T) {
	input := "myArray[1 + 1]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	indexExp, ok := statement.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expression not *ast.IndexExpression. got=%T", statement.Expression)
	}

	if !testIdentifier(t, indexExp.Left, "myArray") {
		return
	}

	if !testInfixExpression(t, indexExp.Index, 1, "+", 1) {
		return
	}
}

func TestParsingHashLiteralsStringKeys(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := statement.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("expression is not *ast.HashLiteral. got=%T", statement.Expression)
	}

	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not ast.StringLiteral. got=%d", key)
		}

		expectedValue, ok := expected[literal.String()]

		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestParsingEmptyHashLiteral(t *testing.T) {
	input := "{}"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp is not ast.HashLiteral. got=%T", stmt.Expression)
	}
	if len(hash.Pairs) != 0 {
		t.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}
}

func TestParsingHashLiteralsWithExpressions(t *testing.T) {
	input := `{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp is not ast.HashLiteral. got=%T", stmt.Expression)
	}
	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}
	tests := map[string]func(ast.Expression){
		"one": func(e ast.Expression) {
			testInfixExpression(t, e, 0, "+", 1)
		},
		"two": func(e ast.Expression) {
			testInfixExpression(t, e, 10, "-", 8)
		},
		"three": func(e ast.Expression) {
			testInfixExpression(t, e, 15, "/", 5)
		},
	}
	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not ast.StringLiteral. got=%T", key)
			continue
		}
		testFunc, ok := tests[literal.String()]
		if !ok {
			t.Errorf("No test function for key %q found", literal.String())
			continue
		}
		testFunc(value)
	}
}
