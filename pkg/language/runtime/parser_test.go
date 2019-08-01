package runtime

import (
	. "github.com/dapperlabs/bamboo-node/pkg/language/runtime/ast"
	"github.com/dapperlabs/bamboo-node/pkg/language/runtime/parser"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"math/big"
	"testing"
)

func init() {
	format.TruncatedDiff = false
	format.MaxDepth = 100
}

func TestParseIncompleteConstKeyword(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    le
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(1))

	syntaxError := errors[0].(*parser.SyntaxError)

	Expect(syntaxError.Pos).
		To(Equal(&Position{Offset: 6, Line: 2, Column: 5}))

	Expect(syntaxError.Message).
		To(ContainSubstring("extraneous input"))
}

func TestParseIncompleteConstantDeclaration1(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(1))

	syntaxError1 := errors[0].(*parser.SyntaxError)

	Expect(syntaxError1.Pos).
		To(Equal(&Position{Offset: 11, Line: 3, Column: 1}))

	Expect(syntaxError1.Message).
		To(ContainSubstring("expecting Identifier"))
}

func TestParseIncompleteConstantDeclaration2(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let =
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(2))

	syntaxError1 := errors[0].(*parser.SyntaxError)

	Expect(syntaxError1.Pos).
		To(Equal(&Position{Offset: 10, Line: 2, Column: 9}))

	Expect(syntaxError1.Message).
		To(ContainSubstring("missing Identifier"))

	syntaxError2 := errors[1].(*parser.SyntaxError)

	Expect(syntaxError2.Pos).
		To(Equal(&Position{Offset: 13, Line: 3, Column: 1}))

	Expect(syntaxError2.Message).
		To(ContainSubstring("mismatched input"))
}

func TestParseBoolExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let a = true
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Value: &BoolExpression{
			Value: true,
			Pos:   &Position{Offset: 14, Line: 2, Column: 13},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 14, Line: 2, Column: 13},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseIdentifierExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let b = a
	`)

	Expect(errors).
		To(BeEmpty())

	b := &VariableDeclaration{
		IsConstant: true,
		Identifier: "b",
		Value: &IdentifierExpression{
			Identifier: "a",
			StartPos:   &Position{Offset: 14, Line: 2, Column: 13},
			EndPos:     &Position{Offset: 14, Line: 2, Column: 13},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 14, Line: 2, Column: 13},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{b},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseArrayExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let a = [1, 2]
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Value: &ArrayExpression{
			Values: []Expression{
				&IntExpression{
					Value: big.NewInt(1),
					Pos:   &Position{Offset: 15, Line: 2, Column: 14},
				},
				&IntExpression{
					Value: big.NewInt(2),
					Pos:   &Position{Offset: 18, Line: 2, Column: 17},
				},
			},
			StartPos: &Position{Offset: 14, Line: 2, Column: 13},
			EndPos:   &Position{Offset: 19, Line: 2, Column: 18},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 19, Line: 2, Column: 18},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseInvocationExpressionWithoutLabels(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let a = b(1, 2)
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Value: &InvocationExpression{
			Expression: &IdentifierExpression{
				Identifier: "b",
				StartPos:   &Position{Offset: 14, Line: 2, Column: 13},
				EndPos:     &Position{Offset: 14, Line: 2, Column: 13},
			},
			Arguments: []*Argument{
				{
					Label: "",
					Expression: &IntExpression{
						Value: big.NewInt(1),
						Pos:   &Position{Offset: 16, Line: 2, Column: 15},
					},
				},
				{
					Label: "",
					Expression: &IntExpression{
						Value: big.NewInt(2),
						Pos:   &Position{Offset: 19, Line: 2, Column: 18},
					},
				},
			},
			StartPos: &Position{Offset: 15, Line: 2, Column: 14},
			EndPos:   &Position{Offset: 20, Line: 2, Column: 19},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 20, Line: 2, Column: 19},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseInvocationExpressionWithLabels(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let a = b(x: 1, y: 2)
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Value: &InvocationExpression{
			Expression: &IdentifierExpression{
				Identifier: "b",
				StartPos:   &Position{Offset: 14, Line: 2, Column: 13},
				EndPos:     &Position{Offset: 14, Line: 2, Column: 13},
			},
			Arguments: []*Argument{
				{
					Label: "x",
					Expression: &IntExpression{
						Value: big.NewInt(1),
						Pos:   &Position{Offset: 19, Line: 2, Column: 18},
					},
				},
				{
					Label: "y",
					Expression: &IntExpression{
						Value: big.NewInt(2),
						Pos:   &Position{Offset: 25, Line: 2, Column: 24},
					},
				},
			},
			StartPos: &Position{Offset: 15, Line: 2, Column: 14},
			EndPos:   &Position{Offset: 26, Line: 2, Column: 25},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 26, Line: 2, Column: 25},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseMemberExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let a = b.c
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Value: &MemberExpression{
			Expression: &IdentifierExpression{
				Identifier: "b",
				StartPos:   &Position{Offset: 14, Line: 2, Column: 13},
				EndPos:     &Position{Offset: 14, Line: 2, Column: 13},
			},
			Identifier: "c",
			StartPos:   &Position{Offset: 15, Line: 2, Column: 14},
			EndPos:     &Position{Offset: 16, Line: 2, Column: 15},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 16, Line: 2, Column: 15},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseIndexExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let a = b[1]
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Value: &IndexExpression{
			Expression: &IdentifierExpression{
				Identifier: "b",
				StartPos:   &Position{Offset: 14, Line: 2, Column: 13},
				EndPos:     &Position{Offset: 14, Line: 2, Column: 13},
			},
			Index: &IntExpression{
				Value: big.NewInt(1),
				Pos:   &Position{Offset: 16, Line: 2, Column: 15},
			},
			StartPos: &Position{Offset: 15, Line: 2, Column: 14},
			EndPos:   &Position{Offset: 17, Line: 2, Column: 16},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 17, Line: 2, Column: 16},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseUnaryExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let a = -b
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Value: &UnaryExpression{
			Operation: OperationMinus,
			Expression: &IdentifierExpression{
				Identifier: "b",
				StartPos:   &Position{Offset: 15, Line: 2, Column: 14},
				EndPos:     &Position{Offset: 15, Line: 2, Column: 14},
			},
			StartPos: &Position{Offset: 14, Line: 2, Column: 13},
			EndPos:   &Position{Offset: 15, Line: 2, Column: 14},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 15, Line: 2, Column: 14},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseOrExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
        let a = false || true
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Type:       Type(nil),
		Value: &BinaryExpression{
			Operation: OperationOr,
			Left: &BoolExpression{
				Value: false,
				Pos:   &Position{Offset: 17, Line: 2, Column: 16},
			},
			Right: &BoolExpression{
				Value: true,
				Pos:   &Position{Offset: 26, Line: 2, Column: 25},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 26, Line: 2, Column: 25},
		},
		StartPos:      &Position{Offset: 9, Line: 2, Column: 8},
		EndPos:        &Position{Offset: 26, Line: 2, Column: 25},
		IdentifierPos: &Position{Offset: 13, Line: 2, Column: 12},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseAndExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
        let a = false && true
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Type:       Type(nil),
		Value: &BinaryExpression{
			Operation: OperationAnd,
			Left: &BoolExpression{
				Value: false,
				Pos:   &Position{Offset: 17, Line: 2, Column: 16},
			},
			Right: &BoolExpression{
				Value: true,
				Pos:   &Position{Offset: 26, Line: 2, Column: 25},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 26, Line: 2, Column: 25},
		},
		StartPos:      &Position{Offset: 9, Line: 2, Column: 8},
		EndPos:        &Position{Offset: 26, Line: 2, Column: 25},
		IdentifierPos: &Position{Offset: 13, Line: 2, Column: 12},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseEqualityExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
        let a = false == true
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Type:       Type(nil),
		Value: &BinaryExpression{
			Operation: OperationEqual,
			Left: &BoolExpression{
				Value: false,
				Pos:   &Position{Offset: 17, Line: 2, Column: 16},
			},
			Right: &BoolExpression{
				Value: true,
				Pos:   &Position{Offset: 26, Line: 2, Column: 25},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 26, Line: 2, Column: 25},
		},
		StartPos:      &Position{Offset: 9, Line: 2, Column: 8},
		EndPos:        &Position{Offset: 26, Line: 2, Column: 25},
		IdentifierPos: &Position{Offset: 13, Line: 2, Column: 12},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseRelationalExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
        let a = 1 < 2
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Type:       Type(nil),
		Value: &BinaryExpression{
			Operation: OperationLess,
			Left: &IntExpression{
				Value: big.NewInt(1),
				Pos:   &Position{Offset: 17, Line: 2, Column: 16},
			},
			Right: &IntExpression{
				Value: big.NewInt(2),
				Pos:   &Position{Offset: 21, Line: 2, Column: 20},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 21, Line: 2, Column: 20},
		},
		StartPos:      &Position{Offset: 9, Line: 2, Column: 8},
		EndPos:        &Position{Offset: 21, Line: 2, Column: 20},
		IdentifierPos: &Position{Offset: 13, Line: 2, Column: 12},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseAdditiveExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
        let a = 1 + 2
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Type:       Type(nil),
		Value: &BinaryExpression{
			Operation: OperationPlus,
			Left: &IntExpression{
				Value: big.NewInt(1),
				Pos:   &Position{Offset: 17, Line: 2, Column: 16},
			},
			Right: &IntExpression{
				Value: big.NewInt(2),
				Pos:   &Position{Offset: 21, Line: 2, Column: 20},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 21, Line: 2, Column: 20},
		},
		StartPos:      &Position{Offset: 9, Line: 2, Column: 8},
		EndPos:        &Position{Offset: 21, Line: 2, Column: 20},
		IdentifierPos: &Position{Offset: 13, Line: 2, Column: 12},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseMultiplicativeExpression(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
        let a = 1 * 2
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Type:       Type(nil),
		Value: &BinaryExpression{
			Operation: OperationMul,
			Left: &IntExpression{
				Value: big.NewInt(1),
				Pos:   &Position{Offset: 17, Line: 2, Column: 16},
			},
			Right: &IntExpression{
				Value: big.NewInt(2),
				Pos:   &Position{Offset: 21, Line: 2, Column: 20},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 21, Line: 2, Column: 20},
		},
		StartPos:      &Position{Offset: 9, Line: 2, Column: 8},
		EndPos:        &Position{Offset: 21, Line: 2, Column: 20},
		IdentifierPos: &Position{Offset: 13, Line: 2, Column: 12},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseFunctionExpressionAndReturn(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    let test = fun (): Int { return 1 }
	`)

	Expect(errors).
		To(BeEmpty())

	test := &VariableDeclaration{
		IsConstant: true,
		Identifier: "test",
		Value: &FunctionExpression{
			ReturnType: &BaseType{
				Identifier: "Int",
				Pos:        &Position{Offset: 25, Line: 2, Column: 24},
			},
			Block: &Block{
				Statements: []Statement{
					&ReturnStatement{
						Expression: &IntExpression{
							Value: big.NewInt(1),
							Pos:   &Position{Offset: 38, Line: 2, Column: 37},
						},
						StartPos: &Position{Offset: 31, Line: 2, Column: 30},
						EndPos:   &Position{Offset: 38, Line: 2, Column: 37},
					},
				},
				StartPos: &Position{Offset: 29, Line: 2, Column: 28},
				EndPos:   &Position{Offset: 40, Line: 2, Column: 39},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 40, Line: 2, Column: 39},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 40, Line: 2, Column: 39},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseFunctionAndBlock(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    fun test() { return }
	`)

	Expect(errors).
		To(BeEmpty())

	test := &FunctionDeclaration{
		IsPublic:   false,
		Identifier: "test",
		ReturnType: &BaseType{
			Pos: &Position{Offset: 15, Line: 2, Column: 14},
		},
		Block: &Block{
			Statements: []Statement{
				&ReturnStatement{
					StartPos: &Position{Offset: 19, Line: 2, Column: 18},
					EndPos:   &Position{Offset: 19, Line: 2, Column: 18},
				},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 26, Line: 2, Column: 25},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 26, Line: 2, Column: 25},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseFunctionParameterWithoutLabel(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    fun test(x: Int) { }
	`)

	Expect(errors).
		To(BeEmpty())

	test := &FunctionDeclaration{
		IsPublic:   false,
		Identifier: "test",
		Parameters: []*Parameter{
			{
				Label:      "",
				Identifier: "x",
				Type: &BaseType{
					Identifier: "Int",
					Pos:        &Position{Offset: 18, Line: 2, Column: 17},
				},
				LabelPos:      nil,
				IdentifierPos: &Position{Offset: 15, Line: 2, Column: 14},
				StartPos:      &Position{Offset: 15, Line: 2, Column: 14},
				EndPos:        &Position{Offset: 18, Line: 2, Column: 17},
			},
		},
		ReturnType: &BaseType{
			Pos: &Position{Offset: 21, Line: 2, Column: 20},
		},
		Block: &Block{
			StartPos: &Position{Offset: 23, Line: 2, Column: 22},
			EndPos:   &Position{Offset: 25, Line: 2, Column: 24},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 25, Line: 2, Column: 24},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseFunctionParameterWithLabel(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    fun test(x y: Int) { }
	`)

	Expect(errors).
		To(BeEmpty())

	test := &FunctionDeclaration{
		IsPublic:   false,
		Identifier: "test",
		Parameters: []*Parameter{
			{
				Label:      "x",
				Identifier: "y",
				Type: &BaseType{
					Identifier: "Int",
					Pos:        &Position{Offset: 20, Line: 2, Column: 19},
				},
				LabelPos:      &Position{Offset: 15, Line: 2, Column: 14},
				IdentifierPos: &Position{Offset: 17, Line: 2, Column: 16},
				StartPos:      &Position{Offset: 15, Line: 2, Column: 14},
				EndPos:        &Position{Offset: 20, Line: 2, Column: 19},
			},
		},
		ReturnType: &BaseType{
			Pos: &Position{Offset: 23, Line: 2, Column: 22},
		},
		Block: &Block{
			StartPos: &Position{Offset: 25, Line: 2, Column: 24},
			EndPos:   &Position{Offset: 27, Line: 2, Column: 26},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 27, Line: 2, Column: 26},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseIfStatement(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    fun test() {
            if true {
                return
            } else if false {
                false
                1
            } else {
                2
            }
        }
	`)

	Expect(errors).
		To(BeEmpty())

	test := &FunctionDeclaration{
		IsPublic:   false,
		Identifier: "test",
		ReturnType: &BaseType{
			Pos: &Position{Offset: 15, Line: 2, Column: 14},
		},
		Block: &Block{
			Statements: []Statement{
				&IfStatement{
					Test: &BoolExpression{
						Value: true,
						Pos:   &Position{Offset: 34, Line: 3, Column: 15},
					},
					Then: &Block{
						Statements: []Statement{
							&ReturnStatement{
								Expression: nil,
								StartPos:   &Position{Offset: 57, Line: 4, Column: 16},
								EndPos:     &Position{Offset: 57, Line: 4, Column: 16},
							},
						},
						StartPos: &Position{Offset: 39, Line: 3, Column: 20},
						EndPos:   &Position{Offset: 76, Line: 5, Column: 12},
					},
					Else: &Block{
						Statements: []Statement{
							&IfStatement{
								Test: &BoolExpression{
									Value: false,
									Pos:   &Position{Offset: 86, Line: 5, Column: 22},
								},
								Then: &Block{
									Statements: []Statement{
										&ExpressionStatement{
											Expression: &BoolExpression{
												Value: false,
												Pos:   &Position{Offset: 110, Line: 6, Column: 16},
											},
										},
										&ExpressionStatement{
											Expression: &IntExpression{
												Value: big.NewInt(1),
												Pos:   &Position{Offset: 132, Line: 7, Column: 16},
											},
										},
									},
									StartPos: &Position{Offset: 92, Line: 5, Column: 28},
									EndPos:   &Position{Offset: 146, Line: 8, Column: 12},
								},
								Else: &Block{
									Statements: []Statement{
										&ExpressionStatement{
											Expression: &IntExpression{
												Value: big.NewInt(2),
												Pos:   &Position{Offset: 171, Line: 9, Column: 16},
											},
										},
									},
									StartPos: &Position{Offset: 153, Line: 8, Column: 19},
									EndPos:   &Position{Offset: 185, Line: 10, Column: 12},
								},
								StartPos: &Position{Offset: 83, Line: 5, Column: 19},
								EndPos:   &Position{Offset: 185, Line: 10, Column: 12},
							},
						},
						StartPos: &Position{Offset: 83, Line: 5, Column: 19},
						EndPos:   &Position{Offset: 185, Line: 10, Column: 12},
					},
					StartPos: &Position{Offset: 31, Line: 3, Column: 12},
					EndPos:   &Position{Offset: 185, Line: 10, Column: 12},
				},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 195, Line: 11, Column: 8},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 195, Line: 11, Column: 8},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseIfStatementNoElse(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    fun test() {
            if true {
                return
            }
        }
	`)

	Expect(errors).
		To(BeEmpty())

	test := &FunctionDeclaration{
		IsPublic:   false,
		Identifier: "test",
		ReturnType: &BaseType{
			Pos: &Position{Offset: 15, Line: 2, Column: 14},
		},
		Block: &Block{
			Statements: []Statement{
				&IfStatement{
					Test: &BoolExpression{
						Value: true,
						Pos:   &Position{Offset: 34, Line: 3, Column: 15},
					},
					Then: &Block{
						Statements: []Statement{
							&ReturnStatement{
								Expression: nil,
								StartPos:   &Position{Offset: 57, Line: 4, Column: 16},
								EndPos:     &Position{Offset: 57, Line: 4, Column: 16},
							},
						},
						StartPos: &Position{Offset: 39, Line: 3, Column: 20},
						EndPos:   &Position{Offset: 76, Line: 5, Column: 12},
					},
					StartPos: &Position{Offset: 31, Line: 3, Column: 12},
					EndPos:   &Position{Offset: 76, Line: 5, Column: 12},
				},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 86, Line: 6, Column: 8},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 86, Line: 6, Column: 8},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseWhileStatement(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    fun test() {
            while true {
              return
            }
        }
	`)

	Expect(errors).
		To(BeEmpty())

	test := &FunctionDeclaration{
		IsPublic:   false,
		Identifier: "test",
		ReturnType: &BaseType{
			Pos: &Position{Offset: 15, Line: 2, Column: 14},
		},
		Block: &Block{
			Statements: []Statement{
				&WhileStatement{
					Test: &BoolExpression{
						Value: true,
						Pos:   &Position{Offset: 37, Line: 3, Column: 18},
					},
					Block: &Block{
						Statements: []Statement{
							&ReturnStatement{
								Expression: nil,
								StartPos:   &Position{Offset: 58, Line: 4, Column: 14},
								EndPos:     &Position{Offset: 58, Line: 4, Column: 14},
							},
						},
						StartPos: &Position{Offset: 42, Line: 3, Column: 23},
						EndPos:   &Position{Offset: 77, Line: 5, Column: 12},
					},
					StartPos: &Position{Offset: 31, Line: 3, Column: 12},
					EndPos:   &Position{Offset: 77, Line: 5, Column: 12},
				},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 87, Line: 6, Column: 8},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 87, Line: 6, Column: 8},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseAssignment(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    fun test() {
            a = 1
        }
	`)

	Expect(errors).
		To(BeEmpty())

	test := &FunctionDeclaration{
		IsPublic:   false,
		Identifier: "test",
		ReturnType: &BaseType{
			Pos: &Position{Offset: 15, Line: 2, Column: 14},
		},
		Block: &Block{
			Statements: []Statement{
				&AssignmentStatement{
					Target: &IdentifierExpression{
						Identifier: "a",
						StartPos:   &Position{Offset: 31, Line: 3, Column: 12},
						EndPos:     &Position{Offset: 31, Line: 3, Column: 12},
					},
					Value: &IntExpression{
						Value: big.NewInt(1),
						Pos:   &Position{Offset: 35, Line: 3, Column: 16},
					},
					StartPos: &Position{Offset: 31, Line: 3, Column: 12},
					EndPos:   &Position{Offset: 35, Line: 3, Column: 16},
				},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 45, Line: 4, Column: 8},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 45, Line: 4, Column: 8},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseAccessAssignment(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    fun test() {
            x.foo.bar[0][1].baz = 1
        }
	`)

	Expect(errors).
		To(BeEmpty())

	test := &FunctionDeclaration{
		IsPublic:   false,
		Identifier: "test",
		ReturnType: &BaseType{
			Pos: &Position{Offset: 15, Line: 2, Column: 14},
		},
		Block: &Block{
			Statements: []Statement{
				&AssignmentStatement{
					Target: &MemberExpression{
						Expression: &IndexExpression{
							Expression: &IndexExpression{
								Expression: &MemberExpression{
									Expression: &MemberExpression{
										Expression: &IdentifierExpression{
											Identifier: "x",
											StartPos:   &Position{Offset: 31, Line: 3, Column: 12},
											EndPos:     &Position{Offset: 31, Line: 3, Column: 12},
										},
										Identifier: "foo",
										StartPos:   &Position{Offset: 32, Line: 3, Column: 13},
										EndPos:     &Position{Offset: 33, Line: 3, Column: 14},
									},
									Identifier: "bar",
									StartPos:   &Position{Offset: 36, Line: 3, Column: 17},
									EndPos:     &Position{Offset: 37, Line: 3, Column: 18},
								},
								Index: &IntExpression{
									Value: big.NewInt(0),
									Pos:   &Position{Offset: 41, Line: 3, Column: 22},
								},
								StartPos: &Position{Offset: 40, Line: 3, Column: 21},
								EndPos:   &Position{Offset: 42, Line: 3, Column: 23},
							},
							Index: &IntExpression{
								Value: big.NewInt(1),
								Pos:   &Position{Offset: 44, Line: 3, Column: 25},
							},
							StartPos: &Position{Offset: 43, Line: 3, Column: 24},
							EndPos:   &Position{Offset: 45, Line: 3, Column: 26},
						},
						Identifier: "baz",
						StartPos:   &Position{Offset: 46, Line: 3, Column: 27},
						EndPos:     &Position{Offset: 47, Line: 3, Column: 28},
					},
					Value: &IntExpression{
						Value: big.NewInt(1),
						Pos:   &Position{Offset: 53, Line: 3, Column: 34},
					},
					StartPos: &Position{Offset: 31, Line: 3, Column: 12},
					EndPos:   &Position{Offset: 53, Line: 3, Column: 34},
				},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 63, Line: 4, Column: 8},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 63, Line: 4, Column: 8},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseExpressionStatementWithAccess(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
	    fun test() { x.foo.bar[0][1].baz }
	`)

	Expect(errors).
		To(BeEmpty())

	test := &FunctionDeclaration{
		IsPublic:   false,
		Identifier: "test",
		ReturnType: &BaseType{
			Pos: &Position{Offset: 15, Line: 2, Column: 14},
		},
		Block: &Block{
			Statements: []Statement{
				&ExpressionStatement{
					Expression: &MemberExpression{
						Expression: &IndexExpression{
							Expression: &IndexExpression{
								Expression: &MemberExpression{
									Expression: &MemberExpression{
										Expression: &IdentifierExpression{
											Identifier: "x",
											StartPos:   &Position{Offset: 19, Line: 2, Column: 18},
											EndPos:     &Position{Offset: 19, Line: 2, Column: 18},
										},
										Identifier: "foo",
										StartPos:   &Position{Offset: 20, Line: 2, Column: 19},
										EndPos:     &Position{Offset: 21, Line: 2, Column: 20},
									},
									Identifier: "bar",
									StartPos:   &Position{Offset: 24, Line: 2, Column: 23},
									EndPos:     &Position{Offset: 25, Line: 2, Column: 24},
								},
								Index: &IntExpression{
									Value: big.NewInt(0),
									Pos:   &Position{Offset: 29, Line: 2, Column: 28},
								},
								StartPos: &Position{Offset: 28, Line: 2, Column: 27},
								EndPos:   &Position{Offset: 30, Line: 2, Column: 29},
							},
							Index: &IntExpression{
								Value: big.NewInt(1),
								Pos:   &Position{Offset: 32, Line: 2, Column: 31},
							},
							StartPos: &Position{Offset: 31, Line: 2, Column: 30},
							EndPos:   &Position{Offset: 33, Line: 2, Column: 32},
						},
						Identifier: "baz",
						StartPos:   &Position{Offset: 34, Line: 2, Column: 33},
						EndPos:     &Position{Offset: 35, Line: 2, Column: 34},
					},
				},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 39, Line: 2, Column: 38},
		},
		StartPos:      &Position{Offset: 6, Line: 2, Column: 5},
		EndPos:        &Position{Offset: 39, Line: 2, Column: 38},
		IdentifierPos: &Position{Offset: 10, Line: 2, Column: 9},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseParametersAndArrayTypes(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		pub fun test(a: Int32, b: Int32[2], c: Int32[][3]): Int64[][] {}
	`)

	Expect(errors).
		To(BeEmpty())

	test := &FunctionDeclaration{
		IsPublic:   true,
		Identifier: "test",
		Parameters: []*Parameter{
			{
				Identifier: "a",
				Type: &BaseType{
					Identifier: "Int32",
					Pos:        &Position{Offset: 19, Line: 2, Column: 18},
				},
				IdentifierPos: &Position{Offset: 16, Line: 2, Column: 15},
				StartPos:      &Position{Offset: 16, Line: 2, Column: 15},
				EndPos:        &Position{Offset: 19, Line: 2, Column: 18},
			},
			{
				Identifier: "b",
				Type: &ConstantSizedType{
					Type: &BaseType{
						Identifier: "Int32",
						Pos:        &Position{Offset: 29, Line: 2, Column: 28},
					},
					Size:     2,
					StartPos: &Position{Offset: 34, Line: 2, Column: 33},
					EndPos:   &Position{Offset: 36, Line: 2, Column: 35},
				},
				IdentifierPos: &Position{Offset: 26, Line: 2, Column: 25},
				StartPos:      &Position{Offset: 26, Line: 2, Column: 25},
				EndPos:        &Position{Offset: 36, Line: 2, Column: 35},
			},
			{
				Identifier: "c",
				Type: &VariableSizedType{
					Type: &ConstantSizedType{
						Type: &BaseType{
							Identifier: "Int32",
							Pos:        &Position{Offset: 42, Line: 2, Column: 41},
						},
						Size:     3,
						StartPos: &Position{Offset: 49, Line: 2, Column: 48},
						EndPos:   &Position{Offset: 51, Line: 2, Column: 50},
					},
					StartPos: &Position{Offset: 47, Line: 2, Column: 46},
					EndPos:   &Position{Offset: 48, Line: 2, Column: 47},
				},
				IdentifierPos: &Position{Offset: 39, Line: 2, Column: 38},
				StartPos:      &Position{Offset: 39, Line: 2, Column: 38},
				EndPos:        &Position{Offset: 51, Line: 2, Column: 50},
			},
		},
		ReturnType: &VariableSizedType{
			Type: &VariableSizedType{
				Type: &BaseType{
					Identifier: "Int64",
					Pos:        &Position{Offset: 55, Line: 2, Column: 54},
				},
				StartPos: &Position{Offset: 62, Line: 2, Column: 61},
				EndPos:   &Position{Offset: 63, Line: 2, Column: 62},
			},
			StartPos: &Position{Offset: 60, Line: 2, Column: 59},
			EndPos:   &Position{Offset: 61, Line: 2, Column: 60},
		},
		Block: &Block{
			StartPos: &Position{Offset: 65, Line: 2, Column: 64},
			EndPos:   &Position{Offset: 66, Line: 2, Column: 65},
		},
		StartPos:      &Position{Offset: 3, Line: 2, Column: 2},
		EndPos:        &Position{Offset: 66, Line: 2, Column: 65},
		IdentifierPos: &Position{Offset: 11, Line: 2, Column: 10},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseIntegerLiterals(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let octal = 0o32
        let hex = 0xf2
        let binary = 0b101010
        let decimal = 1234567890
	`)

	Expect(errors).
		To(BeEmpty())

	octal := &VariableDeclaration{
		Identifier: "octal",
		IsConstant: true,
		Value: &IntExpression{
			Value: big.NewInt(26),
			Pos:   &Position{Offset: 15, Line: 2, Column: 14},
		},
		StartPos:      &Position{Offset: 3, Line: 2, Column: 2},
		EndPos:        &Position{Offset: 15, Line: 2, Column: 14},
		IdentifierPos: &Position{Offset: 7, Line: 2, Column: 6},
	}

	hex := &VariableDeclaration{
		Identifier: "hex",
		IsConstant: true,
		Value: &IntExpression{
			Value: big.NewInt(242),
			Pos:   &Position{Offset: 38, Line: 3, Column: 18},
		},
		StartPos:      &Position{Offset: 28, Line: 3, Column: 8},
		EndPos:        &Position{Offset: 38, Line: 3, Column: 18},
		IdentifierPos: &Position{Offset: 32, Line: 3, Column: 12},
	}

	binary := &VariableDeclaration{
		Identifier: "binary",
		IsConstant: true,
		Value: &IntExpression{
			Value: big.NewInt(42),
			Pos:   &Position{Offset: 64, Line: 4, Column: 21},
		},
		StartPos:      &Position{Offset: 51, Line: 4, Column: 8},
		EndPos:        &Position{Offset: 64, Line: 4, Column: 21},
		IdentifierPos: &Position{Offset: 55, Line: 4, Column: 12},
	}

	decimal := &VariableDeclaration{
		Identifier: "decimal",
		IsConstant: true,
		Value: &IntExpression{
			Value: big.NewInt(1234567890),
			Pos:   &Position{Offset: 95, Line: 5, Column: 22},
		},
		StartPos:      &Position{Offset: 81, Line: 5, Column: 8},
		EndPos:        &Position{Offset: 95, Line: 5, Column: 22},
		IdentifierPos: &Position{Offset: 85, Line: 5, Column: 12},
	}

	expected := &Program{
		Declarations: []Declaration{octal, hex, binary, decimal},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseIntegerLiteralsWithUnderscores(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let octal = 0o32_45
        let hex = 0xf2_09
        let binary = 0b101010_101010
        let decimal = 1_234_567_890
	`)

	Expect(errors).
		To(BeEmpty())

	octal := &VariableDeclaration{
		Identifier: "octal",
		IsConstant: true,
		Value: &IntExpression{
			Value: big.NewInt(1701),
			Pos:   &Position{Offset: 15, Line: 2, Column: 14},
		},
		StartPos:      &Position{Offset: 3, Line: 2, Column: 2},
		EndPos:        &Position{Offset: 15, Line: 2, Column: 14},
		IdentifierPos: &Position{Offset: 7, Line: 2, Column: 6},
	}

	hex := &VariableDeclaration{
		Identifier: "hex",
		IsConstant: true,
		Value: &IntExpression{
			Value: big.NewInt(61961),
			Pos:   &Position{Offset: 41, Line: 3, Column: 18},
		},
		StartPos:      &Position{Offset: 31, Line: 3, Column: 8},
		EndPos:        &Position{Offset: 41, Line: 3, Column: 18},
		IdentifierPos: &Position{Offset: 35, Line: 3, Column: 12},
	}

	binary := &VariableDeclaration{
		Identifier: "binary",
		IsConstant: true,
		Value: &IntExpression{
			Value: big.NewInt(2730),
			Pos:   &Position{Offset: 70, Line: 4, Column: 21},
		},
		StartPos:      &Position{Offset: 57, Line: 4, Column: 8},
		EndPos:        &Position{Offset: 70, Line: 4, Column: 21},
		IdentifierPos: &Position{Offset: 61, Line: 4, Column: 12},
	}

	decimal := &VariableDeclaration{
		Identifier: "decimal",
		IsConstant: true,
		Value: &IntExpression{
			Value: big.NewInt(1234567890),
			Pos:   &Position{Offset: 108, Line: 5, Column: 22},
		},
		StartPos:      &Position{Offset: 94, Line: 5, Column: 8},
		EndPos:        &Position{Offset: 108, Line: 5, Column: 22},
		IdentifierPos: &Position{Offset: 98, Line: 5, Column: 12},
	}

	expected := &Program{
		Declarations: []Declaration{octal, hex, binary, decimal},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseInvalidOctalIntegerLiteralWithLeadingUnderscore(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let octal = 0o_32_45
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(1))

	syntaxError := errors[0].(*parser.InvalidIntegerLiteralError)

	Expect(syntaxError.StartPos).
		To(Equal(&Position{Offset: 15, Line: 2, Column: 14}))

	Expect(syntaxError.EndPos).
		To(Equal(&Position{Offset: 22, Line: 2, Column: 21}))

	Expect(syntaxError.IntegerLiteralKind).
		To(Equal(parser.IntegerLiteralKindOctal))

	Expect(syntaxError.InvalidIntegerLiteralKind).
		To(Equal(parser.InvalidIntegerLiteralKindLeadingUnderscore))
}

func TestParseInvalidOctalIntegerLiteralWithTrailingUnderscore(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let octal = 0o32_45_
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(1))

	syntaxError := errors[0].(*parser.InvalidIntegerLiteralError)

	Expect(syntaxError.StartPos).
		To(Equal(&Position{Offset: 15, Line: 2, Column: 14}))

	Expect(syntaxError.EndPos).
		To(Equal(&Position{Offset: 22, Line: 2, Column: 21}))

	Expect(syntaxError.IntegerLiteralKind).
		To(Equal(parser.IntegerLiteralKindOctal))

	Expect(syntaxError.InvalidIntegerLiteralKind).
		To(Equal(parser.InvalidIntegerLiteralKindTrailingUnderscore))
}

func TestParseInvalidBinaryIntegerLiteralWithLeadingUnderscore(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let binary = 0b_101010_101010
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(1))

	syntaxError := errors[0].(*parser.InvalidIntegerLiteralError)

	Expect(syntaxError.StartPos).
		To(Equal(&Position{Offset: 16, Line: 2, Column: 15}))

	Expect(syntaxError.EndPos).
		To(Equal(&Position{Offset: 31, Line: 2, Column: 30}))

	Expect(syntaxError.IntegerLiteralKind).
		To(Equal(parser.IntegerLiteralKindBinary))

	Expect(syntaxError.InvalidIntegerLiteralKind).
		To(Equal(parser.InvalidIntegerLiteralKindLeadingUnderscore))
}

func TestParseInvalidBinaryIntegerLiteralWithTrailingUnderscore(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let binary = 0b101010_101010_
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(1))

	syntaxError := errors[0].(*parser.InvalidIntegerLiteralError)

	Expect(syntaxError.StartPos).
		To(Equal(&Position{Offset: 16, Line: 2, Column: 15}))

	Expect(syntaxError.EndPos).
		To(Equal(&Position{Offset: 31, Line: 2, Column: 30}))

	Expect(syntaxError.IntegerLiteralKind).
		To(Equal(parser.IntegerLiteralKindBinary))

	Expect(syntaxError.InvalidIntegerLiteralKind).
		To(Equal(parser.InvalidIntegerLiteralKindTrailingUnderscore))
}

func TestParseInvalidDecimalIntegerLiteralWithTrailingUnderscore(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let decimal = 1_234_567_890_
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(1))

	syntaxError := errors[0].(*parser.InvalidIntegerLiteralError)

	Expect(syntaxError.StartPos).
		To(Equal(&Position{Offset: 17, Line: 2, Column: 16}))

	Expect(syntaxError.EndPos).
		To(Equal(&Position{Offset: 30, Line: 2, Column: 29}))

	Expect(syntaxError.IntegerLiteralKind).
		To(Equal(parser.IntegerLiteralKindDecimal))

	Expect(syntaxError.InvalidIntegerLiteralKind).
		To(Equal(parser.InvalidIntegerLiteralKindTrailingUnderscore))
}

func TestParseInvalidHexadecimalIntegerLiteralWithLeadingUnderscore(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let hex = 0x_f2_09
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(1))

	syntaxError := errors[0].(*parser.InvalidIntegerLiteralError)

	Expect(syntaxError.StartPos).
		To(Equal(&Position{Offset: 13, Line: 2, Column: 12}))

	Expect(syntaxError.EndPos).
		To(Equal(&Position{Offset: 20, Line: 2, Column: 19}))

	Expect(syntaxError.IntegerLiteralKind).
		To(Equal(parser.IntegerLiteralKindHexadecimal))

	Expect(syntaxError.InvalidIntegerLiteralKind).
		To(Equal(parser.InvalidIntegerLiteralKindLeadingUnderscore))
}

func TestParseInvalidHexadecimalIntegerLiteralWithTrailingUnderscore(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let hex = 0xf2_09_
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(1))

	syntaxError := errors[0].(*parser.InvalidIntegerLiteralError)

	Expect(syntaxError.StartPos).
		To(Equal(&Position{Offset: 13, Line: 2, Column: 12}))

	Expect(syntaxError.EndPos).
		To(Equal(&Position{Offset: 20, Line: 2, Column: 19}))

	Expect(syntaxError.IntegerLiteralKind).
		To(Equal(parser.IntegerLiteralKindHexadecimal))

	Expect(syntaxError.InvalidIntegerLiteralKind).
		To(Equal(parser.InvalidIntegerLiteralKindTrailingUnderscore))

}

func TestParseInvalidIntegerLiteral(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let hex = 0z123
	`)

	Expect(actual).
		To(BeNil())

	Expect(errors).
		To(HaveLen(1))

	syntaxError := errors[0].(*parser.InvalidIntegerLiteralError)

	Expect(syntaxError.StartPos).
		To(Equal(&Position{Offset: 13, Line: 2, Column: 12}))

	Expect(syntaxError.EndPos).
		To(Equal(&Position{Offset: 17, Line: 2, Column: 16}))

	Expect(syntaxError.IntegerLiteralKind).
		To(Equal(parser.IntegerLiteralKindUnknown))

	Expect(syntaxError.InvalidIntegerLiteralKind).
		To(Equal(parser.InvalidIntegerLiteralKindUnknownPrefix))
}

func TestParseIntegerTypes(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let a: Int8 = 1
		let b: Int16 = 2
		let c: Int32 = 3
		let d: Int64 = 4
		let e: UInt8 = 5
		let f: UInt16 = 6
		let g: UInt32 = 7
		let h: UInt64 = 8
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		Identifier: "a",
		IsConstant: true,
		Type: &BaseType{
			Identifier: "Int8",
			Pos:        &Position{Offset: 10, Line: 2, Column: 9},
		},
		Value: &IntExpression{
			Value: big.NewInt(1),
			Pos:   &Position{Offset: 17, Line: 2, Column: 16},
		},
		StartPos:      &Position{Offset: 3, Line: 2, Column: 2},
		EndPos:        &Position{Offset: 17, Line: 2, Column: 16},
		IdentifierPos: &Position{Offset: 7, Line: 2, Column: 6},
	}
	b := &VariableDeclaration{
		Identifier: "b",
		IsConstant: true,
		Type: &BaseType{
			Identifier: "Int16",
			Pos:        &Position{Offset: 28, Line: 3, Column: 9},
		},
		Value: &IntExpression{
			Value: big.NewInt(2),
			Pos:   &Position{Offset: 36, Line: 3, Column: 17},
		},
		StartPos:      &Position{Offset: 21, Line: 3, Column: 2},
		EndPos:        &Position{Offset: 36, Line: 3, Column: 17},
		IdentifierPos: &Position{Offset: 25, Line: 3, Column: 6},
	}
	c := &VariableDeclaration{
		Identifier: "c",
		IsConstant: true,
		Type: &BaseType{
			Identifier: "Int32",
			Pos:        &Position{Offset: 47, Line: 4, Column: 9},
		},
		Value: &IntExpression{
			Value: big.NewInt(3),
			Pos:   &Position{Offset: 55, Line: 4, Column: 17},
		},
		StartPos:      &Position{Offset: 40, Line: 4, Column: 2},
		EndPos:        &Position{Offset: 55, Line: 4, Column: 17},
		IdentifierPos: &Position{Offset: 44, Line: 4, Column: 6},
	}
	d := &VariableDeclaration{
		Identifier: "d",
		IsConstant: true,
		Type: &BaseType{
			Identifier: "Int64",
			Pos:        &Position{Offset: 66, Line: 5, Column: 9},
		},
		Value: &IntExpression{
			Value: big.NewInt(4),
			Pos:   &Position{Offset: 74, Line: 5, Column: 17},
		},
		StartPos:      &Position{Offset: 59, Line: 5, Column: 2},
		EndPos:        &Position{Offset: 74, Line: 5, Column: 17},
		IdentifierPos: &Position{Offset: 63, Line: 5, Column: 6},
	}
	e := &VariableDeclaration{
		Identifier: "e",
		IsConstant: true,
		Type: &BaseType{
			Identifier: "UInt8",
			Pos:        &Position{Offset: 85, Line: 6, Column: 9},
		},
		Value: &IntExpression{
			Value: big.NewInt(5),
			Pos:   &Position{Offset: 93, Line: 6, Column: 17},
		},
		StartPos:      &Position{Offset: 78, Line: 6, Column: 2},
		EndPos:        &Position{Offset: 93, Line: 6, Column: 17},
		IdentifierPos: &Position{Offset: 82, Line: 6, Column: 6},
	}
	f := &VariableDeclaration{
		Identifier: "f",
		IsConstant: true,
		Type: &BaseType{
			Identifier: "UInt16",
			Pos:        &Position{Offset: 104, Line: 7, Column: 9},
		},
		Value: &IntExpression{
			Value: big.NewInt(6),
			Pos:   &Position{Offset: 113, Line: 7, Column: 18},
		},
		StartPos:      &Position{Offset: 97, Line: 7, Column: 2},
		EndPos:        &Position{Offset: 113, Line: 7, Column: 18},
		IdentifierPos: &Position{Offset: 101, Line: 7, Column: 6},
	}
	g := &VariableDeclaration{
		Identifier: "g",
		IsConstant: true,
		Type: &BaseType{
			Identifier: "UInt32",
			Pos:        &Position{Offset: 124, Line: 8, Column: 9},
		},
		Value: &IntExpression{
			Value: big.NewInt(7),
			Pos:   &Position{Offset: 133, Line: 8, Column: 18},
		},
		StartPos:      &Position{Offset: 117, Line: 8, Column: 2},
		EndPos:        &Position{Offset: 133, Line: 8, Column: 18},
		IdentifierPos: &Position{Offset: 121, Line: 8, Column: 6},
	}
	h := &VariableDeclaration{
		Identifier: "h",
		IsConstant: true,
		Type: &BaseType{
			Identifier: "UInt64",
			Pos:        &Position{Offset: 144, Line: 9, Column: 9},
		},
		Value: &IntExpression{
			Value: big.NewInt(8),
			Pos:   &Position{Offset: 153, Line: 9, Column: 18},
		},
		StartPos:      &Position{Offset: 137, Line: 9, Column: 2},
		EndPos:        &Position{Offset: 153, Line: 9, Column: 18},
		IdentifierPos: &Position{Offset: 141, Line: 9, Column: 6},
	}

	expected := &Program{
		Declarations: []Declaration{a, b, c, d, e, f, g, h},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseFunctionType(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let add: ((Int8, Int16): Int32) = nothing
	`)

	Expect(errors).
		To(BeEmpty())

	add := &VariableDeclaration{
		Identifier: "add",
		IsConstant: true,
		Type: &FunctionType{
			ParameterTypes: []Type{
				&BaseType{
					Identifier: "Int8",
					Pos:        &Position{Offset: 14, Line: 2, Column: 13},
				},
				&BaseType{
					Identifier: "Int16",
					Pos:        &Position{Offset: 20, Line: 2, Column: 19},
				},
			},
			ReturnType: &BaseType{
				Identifier: "Int32",
				Pos:        &Position{Offset: 28, Line: 2, Column: 27},
			},
			StartPos: &Position{Offset: 12, Line: 2, Column: 11},
			EndPos:   &Position{Offset: 28, Line: 2, Column: 27},
		},
		Value: &IdentifierExpression{
			Identifier: "nothing",
			StartPos:   &Position{Offset: 37, Line: 2, Column: 36},
			EndPos:     &Position{Offset: 43, Line: 2, Column: 42},
		},
		StartPos:      &Position{Offset: 3, Line: 2, Column: 2},
		EndPos:        &Position{Offset: 37, Line: 2, Column: 36},
		IdentifierPos: &Position{Offset: 7, Line: 2, Column: 6},
	}

	expected := &Program{
		Declarations: []Declaration{add},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseFunctionArrayType(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let test: ((Int8): Int16)[2] = []
	`)

	Expect(errors).
		To(BeEmpty())

	test := &VariableDeclaration{
		Identifier: "test",
		IsConstant: true,
		Type: &ConstantSizedType{
			Type: &FunctionType{
				ParameterTypes: []Type{
					&BaseType{
						Identifier: "Int8",
						Pos:        &Position{Offset: 15, Line: 2, Column: 14},
					},
				},
				ReturnType: &BaseType{
					Identifier: "Int16",
					Pos:        &Position{Offset: 22, Line: 2, Column: 21},
				},
				StartPos: &Position{Offset: 13, Line: 2, Column: 12},
				EndPos:   &Position{Offset: 22, Line: 2, Column: 21},
			},
			Size:     2,
			StartPos: &Position{Offset: 28, Line: 2, Column: 27},
			EndPos:   &Position{Offset: 30, Line: 2, Column: 29},
		},
		Value: &ArrayExpression{
			StartPos: &Position{Offset: 34, Line: 2, Column: 33},
			EndPos:   &Position{Offset: 35, Line: 2, Column: 34},
		},
		StartPos:      &Position{Offset: 3, Line: 2, Column: 2},
		EndPos:        &Position{Offset: 35, Line: 2, Column: 34},
		IdentifierPos: &Position{Offset: 7, Line: 2, Column: 6},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseFunctionTypeWithArrayReturnType(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let test: ((Int8): Int16[2]) = nothing
	`)

	Expect(errors).
		To(BeEmpty())

	test := &VariableDeclaration{
		Identifier: "test",
		IsConstant: true,
		Type: &FunctionType{
			ParameterTypes: []Type{
				&BaseType{
					Identifier: "Int8",
					Pos:        &Position{Offset: 15, Line: 2, Column: 14},
				},
			},
			ReturnType: &ConstantSizedType{
				Type: &BaseType{
					Identifier: "Int16",
					Pos:        &Position{Offset: 22, Line: 2, Column: 21},
				},
				Size:     2,
				StartPos: &Position{Offset: 27, Line: 2, Column: 26},
				EndPos:   &Position{Offset: 29, Line: 2, Column: 28},
			},
			StartPos: &Position{Offset: 13, Line: 2, Column: 12},
			EndPos:   &Position{Offset: 29, Line: 2, Column: 28},
		},
		Value: &IdentifierExpression{
			Identifier: "nothing",
			StartPos:   &Position{Offset: 34, Line: 2, Column: 33},
			EndPos:     &Position{Offset: 40, Line: 2, Column: 39},
		},
		StartPos:      &Position{Offset: 3, Line: 2, Column: 2},
		EndPos:        &Position{Offset: 34, Line: 2, Column: 33},
		IdentifierPos: &Position{Offset: 7, Line: 2, Column: 6},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseFunctionTypeWithFunctionReturnTypeInParentheses(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let test: ((Int8): ((Int16): Int32)) = nothing
	`)

	Expect(errors).
		To(BeEmpty())

	test := &VariableDeclaration{
		Identifier: "test",
		IsConstant: true,
		Type: &FunctionType{
			ParameterTypes: []Type{
				&BaseType{
					Identifier: "Int8",
					Pos:        &Position{Offset: 15, Line: 2, Column: 14},
				},
			},
			ReturnType: &FunctionType{
				ParameterTypes: []Type{
					&BaseType{
						Identifier: "Int16",
						Pos:        &Position{Offset: 24, Line: 2, Column: 23},
					},
				},
				ReturnType: &BaseType{
					Identifier: "Int32",
					Pos:        &Position{Offset: 32, Line: 2, Column: 31},
				},
				StartPos: &Position{Offset: 22, Line: 2, Column: 21},
				EndPos:   &Position{Offset: 32, Line: 2, Column: 31},
			},
			StartPos: &Position{Offset: 13, Line: 2, Column: 12},
			EndPos:   &Position{Offset: 32, Line: 2, Column: 31},
		},
		Value: &IdentifierExpression{
			Identifier: "nothing",
			StartPos:   &Position{Offset: 42, Line: 2, Column: 41},
			EndPos:     &Position{Offset: 48, Line: 2, Column: 47},
		},
		StartPos:      &Position{Offset: 3, Line: 2, Column: 2},
		EndPos:        &Position{Offset: 42, Line: 2, Column: 41},
		IdentifierPos: &Position{Offset: 7, Line: 2, Column: 6},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseFunctionTypeWithFunctionReturnType(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let test: ((Int8): ((Int16): Int32)) = nothing
	`)

	Expect(errors).
		To(BeEmpty())

	test := &VariableDeclaration{
		Identifier: "test",
		IsConstant: true,
		Type: &FunctionType{
			ParameterTypes: []Type{
				&BaseType{
					Identifier: "Int8",
					Pos:        &Position{Offset: 15, Line: 2, Column: 14},
				},
			},
			ReturnType: &FunctionType{
				ParameterTypes: []Type{
					&BaseType{
						Identifier: "Int16",
						Pos:        &Position{Offset: 24, Line: 2, Column: 23},
					},
				},
				ReturnType: &BaseType{
					Identifier: "Int32",
					Pos:        &Position{Offset: 32, Line: 2, Column: 31},
				},
				StartPos: &Position{Offset: 22, Line: 2, Column: 21},
				EndPos:   &Position{Offset: 32, Line: 2, Column: 31},
			},
			StartPos: &Position{Offset: 13, Line: 2, Column: 12},
			EndPos:   &Position{Offset: 32, Line: 2, Column: 31},
		},
		Value: &IdentifierExpression{
			Identifier: "nothing",
			StartPos:   &Position{Offset: 42, Line: 2, Column: 41},
			EndPos:     &Position{Offset: 48, Line: 2, Column: 47},
		},
		StartPos:      &Position{Offset: 3, Line: 2, Column: 2},
		EndPos:        &Position{Offset: 42, Line: 2, Column: 41},
		IdentifierPos: &Position{Offset: 7, Line: 2, Column: 6},
	}

	expected := &Program{
		Declarations: []Declaration{test},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseMissingReturnType(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
		let noop: ((): Void) =
            fun () { return }
	`)

	Expect(errors).
		To(BeEmpty())

	noop := &VariableDeclaration{
		Identifier: "noop",
		IsConstant: true,
		Type: &FunctionType{
			ReturnType: &BaseType{
				Identifier: "Void",
				Pos:        &Position{Offset: 18, Line: 2, Column: 17},
			},
			StartPos: &Position{Offset: 13, Line: 2, Column: 12},
			EndPos:   &Position{Offset: 18, Line: 2, Column: 17},
		},
		Value: &FunctionExpression{
			ReturnType: &BaseType{
				Pos: &Position{Offset: 43, Line: 3, Column: 17},
			},
			Block: &Block{
				Statements: []Statement{
					&ReturnStatement{
						StartPos: &Position{Offset: 47, Line: 3, Column: 21},
						EndPos:   &Position{Offset: 47, Line: 3, Column: 21},
					},
				},
				StartPos: &Position{Offset: 45, Line: 3, Column: 19},
				EndPos:   &Position{Offset: 54, Line: 3, Column: 28},
			},
			StartPos: &Position{Offset: 38, Line: 3, Column: 12},
			EndPos:   &Position{Offset: 54, Line: 3, Column: 28},
		},
		StartPos:      &Position{Offset: 3, Line: 2, Column: 2},
		EndPos:        &Position{Offset: 54, Line: 3, Column: 28},
		IdentifierPos: &Position{Offset: 7, Line: 2, Column: 6},
	}

	expected := &Program{
		Declarations: []Declaration{noop},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseLeftAssociativity(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
        let a = 1 + 2 + 3
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Type:       Type(nil),
		Value: &BinaryExpression{
			Operation: OperationPlus,
			Left: &BinaryExpression{
				Operation: OperationPlus,
				Left: &IntExpression{
					Value: big.NewInt(1),
					Pos:   &Position{Offset: 17, Line: 2, Column: 16},
				},
				Right: &IntExpression{
					Value: big.NewInt(2),
					Pos:   &Position{Offset: 21, Line: 2, Column: 20},
				},
				StartPos: &Position{Offset: 17, Line: 2, Column: 16},
				EndPos:   &Position{Offset: 21, Line: 2, Column: 20},
			},
			Right: &IntExpression{
				Value: big.NewInt(3),
				Pos:   &Position{Offset: 25, Line: 2, Column: 24},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 25, Line: 2, Column: 24},
		},
		StartPos:      &Position{Offset: 9, Line: 2, Column: 8},
		EndPos:        &Position{Offset: 25, Line: 2, Column: 24},
		IdentifierPos: &Position{Offset: 13, Line: 2, Column: 12},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}

func TestParseInvalidDoubleIntegerUnary(t *testing.T) {
	RegisterTestingT(t)

	program, errors := parser.Parse(`
	   var a = 1
	   let b = --a
	`)

	Expect(program).
		To(BeNil())

	Expect(errors).
		To(Equal([]error{
			&parser.JuxtaposedUnaryOperatorsError{
				Pos: &Position{Offset: 27, Line: 3, Column: 12},
			},
		}))
}

func TestParseInvalidDoubleBooleanUnary(t *testing.T) {
	RegisterTestingT(t)

	program, errors := parser.Parse(`
	   let b = !!true
	`)

	Expect(program).
		To(BeNil())

	Expect(errors).
		To(Equal([]error{
			&parser.JuxtaposedUnaryOperatorsError{
				Pos: &Position{Offset: 13, Line: 2, Column: 12},
			},
		}))
}

func TestParseTernaryRightAssociativity(t *testing.T) {
	RegisterTestingT(t)

	actual, errors := parser.Parse(`
        let a = 2 > 1
          ? 0
          : 3 > 2 ? 1 : 2
	`)

	Expect(errors).
		To(BeEmpty())

	a := &VariableDeclaration{
		IsConstant: true,
		Identifier: "a",
		Type:       Type(nil),
		Value: &ConditionalExpression{
			Test: &BinaryExpression{
				Operation: OperationGreater,
				Left: &IntExpression{
					Value: big.NewInt(2),
					Pos:   &Position{Offset: 17, Line: 2, Column: 16},
				},
				Right: &IntExpression{
					Value: big.NewInt(1),
					Pos:   &Position{Offset: 21, Line: 2, Column: 20},
				},
				StartPos: &Position{Offset: 17, Line: 2, Column: 16},
				EndPos:   &Position{Offset: 21, Line: 2, Column: 20},
			},
			Then: &IntExpression{
				Value: big.NewInt(0),
				Pos:   &Position{Offset: 35, Line: 3, Column: 12},
			},
			Else: &ConditionalExpression{
				Test: &BinaryExpression{
					Operation: OperationGreater,
					Left: &IntExpression{
						Value: big.NewInt(3),
						Pos:   &Position{Offset: 49, Line: 4, Column: 12},
					},
					Right: &IntExpression{
						Value: big.NewInt(2),
						Pos:   &Position{Offset: 53, Line: 4, Column: 16},
					},
					StartPos: &Position{Offset: 49, Line: 4, Column: 12},
					EndPos:   &Position{Offset: 53, Line: 4, Column: 16},
				},
				Then: &IntExpression{
					Value: big.NewInt(1),
					Pos:   &Position{Offset: 57, Line: 4, Column: 20},
				},
				Else: &IntExpression{
					Value: big.NewInt(2),
					Pos:   &Position{Offset: 61, Line: 4, Column: 24},
				},
				StartPos: &Position{Offset: 49, Line: 4, Column: 12},
				EndPos:   &Position{Offset: 61, Line: 4, Column: 24},
			},
			StartPos: &Position{Offset: 17, Line: 2, Column: 16},
			EndPos:   &Position{Offset: 61, Line: 4, Column: 24},
		},
		StartPos:      &Position{Offset: 9, Line: 2, Column: 8},
		EndPos:        &Position{Offset: 61, Line: 4, Column: 24},
		IdentifierPos: &Position{Offset: 13, Line: 2, Column: 12},
	}

	expected := &Program{
		Declarations: []Declaration{a},
	}

	Expect(actual).
		To(Equal(expected))
}
