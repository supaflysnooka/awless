package ast

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleScript
	ruleStatement
	ruleAction
	ruleEntity
	ruleDeclaration
	ruleExpr
	ruleParams
	ruleParam
	ruleIdentifier
	ruleValue
	ruleStringValue
	ruleCSVValue
	ruleCidrValue
	ruleIpValue
	ruleIntValue
	ruleIntRangeValue
	ruleRefValue
	ruleAliasValue
	ruleHoleValue
	ruleComment
	ruleSpacing
	ruleWhiteSpacing
	ruleMustWhiteSpacing
	ruleEqual
	ruleSpace
	ruleWhitespace
	ruleEndOfLine
	ruleEndOfFile
	rulePegText
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
)

var rul3s = [...]string{
	"Unknown",
	"Script",
	"Statement",
	"Action",
	"Entity",
	"Declaration",
	"Expr",
	"Params",
	"Param",
	"Identifier",
	"Value",
	"StringValue",
	"CSVValue",
	"CidrValue",
	"IpValue",
	"IntValue",
	"IntRangeValue",
	"RefValue",
	"AliasValue",
	"HoleValue",
	"Comment",
	"Spacing",
	"WhiteSpacing",
	"MustWhiteSpacing",
	"Equal",
	"Space",
	"Whitespace",
	"EndOfLine",
	"EndOfFile",
	"PegText",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Printf("%v %v\n", rule, quote)
			} else {
				fmt.Printf("\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(buffer string) {
	node.print(false, buffer)
}

func (node *node32) PrettyPrint(buffer string) {
	node.print(true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type Peg struct {
	*AST

	Buffer string
	buffer []rune
	rules  [45]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *Peg) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *Peg) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *Peg
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *Peg) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *Peg) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.addDeclarationIdentifier(text)
		case ruleAction1:
			p.addAction(text)
		case ruleAction2:
			p.addEntity(text)
		case ruleAction3:
			p.LineDone()
		case ruleAction4:
			p.addParamKey(text)
		case ruleAction5:
			p.addParamHoleValue(text)
		case ruleAction6:
			p.addParamValue(text)
		case ruleAction7:
			p.addParamRefValue(text)
		case ruleAction8:
			p.addParamCidrValue(text)
		case ruleAction9:
			p.addParamIpValue(text)
		case ruleAction10:
			p.addCsvValue(text)
		case ruleAction11:
			p.addParamValue(text)
		case ruleAction12:
			p.addParamIntValue(text)
		case ruleAction13:
			p.addParamValue(text)
		case ruleAction14:
			p.LineDone()

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *Peg) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Script <- <(Spacing Statement+ EndOfFile)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[ruleSpacing]() {
					goto l0
				}
				{
					position4 := position
					if !_rules[ruleSpacing]() {
						goto l0
					}
					{
						position5, tokenIndex5 := position, tokenIndex
						if !_rules[ruleExpr]() {
							goto l6
						}
						goto l5
					l6:
						position, tokenIndex = position5, tokenIndex5
						{
							position8 := position
							{
								position9 := position
								if !_rules[ruleIdentifier]() {
									goto l7
								}
								add(rulePegText, position9)
							}
							{
								add(ruleAction0, position)
							}
							if !_rules[ruleEqual]() {
								goto l7
							}
							if !_rules[ruleExpr]() {
								goto l7
							}
							add(ruleDeclaration, position8)
						}
						goto l5
					l7:
						position, tokenIndex = position5, tokenIndex5
						{
							position11 := position
							{
								position12, tokenIndex12 := position, tokenIndex
								if buffer[position] != rune('#') {
									goto l13
								}
								position++
							l14:
								{
									position15, tokenIndex15 := position, tokenIndex
									{
										position16, tokenIndex16 := position, tokenIndex
										if !_rules[ruleEndOfLine]() {
											goto l16
										}
										goto l15
									l16:
										position, tokenIndex = position16, tokenIndex16
									}
									if !matchDot() {
										goto l15
									}
									goto l14
								l15:
									position, tokenIndex = position15, tokenIndex15
								}
								goto l12
							l13:
								position, tokenIndex = position12, tokenIndex12
								if buffer[position] != rune('/') {
									goto l0
								}
								position++
								if buffer[position] != rune('/') {
									goto l0
								}
								position++
							l17:
								{
									position18, tokenIndex18 := position, tokenIndex
									{
										position19, tokenIndex19 := position, tokenIndex
										if !_rules[ruleEndOfLine]() {
											goto l19
										}
										goto l18
									l19:
										position, tokenIndex = position19, tokenIndex19
									}
									if !matchDot() {
										goto l18
									}
									goto l17
								l18:
									position, tokenIndex = position18, tokenIndex18
								}
								{
									add(ruleAction14, position)
								}
							}
						l12:
							add(ruleComment, position11)
						}
					}
				l5:
					if !_rules[ruleSpacing]() {
						goto l0
					}
				l21:
					{
						position22, tokenIndex22 := position, tokenIndex
						if !_rules[ruleEndOfLine]() {
							goto l22
						}
						goto l21
					l22:
						position, tokenIndex = position22, tokenIndex22
					}
					add(ruleStatement, position4)
				}
			l2:
				{
					position3, tokenIndex3 := position, tokenIndex
					{
						position23 := position
						if !_rules[ruleSpacing]() {
							goto l3
						}
						{
							position24, tokenIndex24 := position, tokenIndex
							if !_rules[ruleExpr]() {
								goto l25
							}
							goto l24
						l25:
							position, tokenIndex = position24, tokenIndex24
							{
								position27 := position
								{
									position28 := position
									if !_rules[ruleIdentifier]() {
										goto l26
									}
									add(rulePegText, position28)
								}
								{
									add(ruleAction0, position)
								}
								if !_rules[ruleEqual]() {
									goto l26
								}
								if !_rules[ruleExpr]() {
									goto l26
								}
								add(ruleDeclaration, position27)
							}
							goto l24
						l26:
							position, tokenIndex = position24, tokenIndex24
							{
								position30 := position
								{
									position31, tokenIndex31 := position, tokenIndex
									if buffer[position] != rune('#') {
										goto l32
									}
									position++
								l33:
									{
										position34, tokenIndex34 := position, tokenIndex
										{
											position35, tokenIndex35 := position, tokenIndex
											if !_rules[ruleEndOfLine]() {
												goto l35
											}
											goto l34
										l35:
											position, tokenIndex = position35, tokenIndex35
										}
										if !matchDot() {
											goto l34
										}
										goto l33
									l34:
										position, tokenIndex = position34, tokenIndex34
									}
									goto l31
								l32:
									position, tokenIndex = position31, tokenIndex31
									if buffer[position] != rune('/') {
										goto l3
									}
									position++
									if buffer[position] != rune('/') {
										goto l3
									}
									position++
								l36:
									{
										position37, tokenIndex37 := position, tokenIndex
										{
											position38, tokenIndex38 := position, tokenIndex
											if !_rules[ruleEndOfLine]() {
												goto l38
											}
											goto l37
										l38:
											position, tokenIndex = position38, tokenIndex38
										}
										if !matchDot() {
											goto l37
										}
										goto l36
									l37:
										position, tokenIndex = position37, tokenIndex37
									}
									{
										add(ruleAction14, position)
									}
								}
							l31:
								add(ruleComment, position30)
							}
						}
					l24:
						if !_rules[ruleSpacing]() {
							goto l3
						}
					l40:
						{
							position41, tokenIndex41 := position, tokenIndex
							if !_rules[ruleEndOfLine]() {
								goto l41
							}
							goto l40
						l41:
							position, tokenIndex = position41, tokenIndex41
						}
						add(ruleStatement, position23)
					}
					goto l2
				l3:
					position, tokenIndex = position3, tokenIndex3
				}
				{
					position42 := position
					{
						position43, tokenIndex43 := position, tokenIndex
						if !matchDot() {
							goto l43
						}
						goto l0
					l43:
						position, tokenIndex = position43, tokenIndex43
					}
					add(ruleEndOfFile, position42)
				}
				add(ruleScript, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 Statement <- <(Spacing (Expr / Declaration / Comment) Spacing EndOfLine*)> */
		nil,
		/* 2 Action <- <(('c' 'r' 'e' 'a' 't' 'e') / ('d' 'e' 'l' 'e' 't' 'e') / ('s' 't' 'a' 'r' 't') / ((&('d') ('d' 'e' 't' 'a' 'c' 'h')) | (&('c') ('c' 'h' 'e' 'c' 'k')) | (&('a') ('a' 't' 't' 'a' 'c' 'h')) | (&('u') ('u' 'p' 'd' 'a' 't' 'e')) | (&('s') ('s' 't' 'o' 'p')) | (&('n') ('n' 'o' 'n' 'e'))))> */
		nil,
		/* 3 Entity <- <(('v' 'p' 'c') / ('s' 'u' 'b' 'n' 'e' 't') / ('i' 'n' 's' 't' 'a' 'n' 'c' 'e') / ('t' 'a' 'g') / ('r' 'o' 'l' 'e') / ('s' 'e' 'c' 'u' 'r' 'i' 't' 'y' 'g' 'r' 'o' 'u' 'p') / ('r' 'o' 'u' 't' 'e' 't' 'a' 'b' 'l' 'e') / ('s' 't' 'o' 'r' 'a' 'g' 'e' 'o' 'b' 'j' 'e' 'c' 't') / ('l' 'o' 'a' 'd' 'b' 'a' 'l' 'a' 'n' 'c' 'e' 'r') / ((&('l') ('l' 'i' 's' 't' 'e' 'n' 'e' 'r')) | (&('q') ('q' 'u' 'e' 'u' 'e')) | (&('t') ('t' 'o' 'p' 'i' 'c')) | (&('s') ('s' 'u' 'b' 's' 'c' 'r' 'i' 'p' 't' 'i' 'o' 'n')) | (&('b') ('b' 'u' 'c' 'k' 'e' 't')) | (&('r') ('r' 'o' 'u' 't' 'e')) | (&('i') ('i' 'n' 't' 'e' 'r' 'n' 'e' 't' 'g' 'a' 't' 'e' 'w' 'a' 'y')) | (&('k') ('k' 'e' 'y' 'p' 'a' 'i' 'r')) | (&('p') ('p' 'o' 'l' 'i' 'c' 'y')) | (&('g') ('g' 'r' 'o' 'u' 'p')) | (&('u') ('u' 's' 'e' 'r')) | (&('v') ('v' 'o' 'l' 'u' 'm' 'e')) | (&('n') ('n' 'o' 'n' 'e'))))> */
		nil,
		/* 4 Declaration <- <(<Identifier> Action0 Equal Expr)> */
		nil,
		/* 5 Expr <- <(<Action> Action1 MustWhiteSpacing <Entity> Action2 (MustWhiteSpacing Params)? Action3)> */
		func() bool {
			position48, tokenIndex48 := position, tokenIndex
			{
				position49 := position
				{
					position50 := position
					{
						position51 := position
						{
							position52, tokenIndex52 := position, tokenIndex
							if buffer[position] != rune('c') {
								goto l53
							}
							position++
							if buffer[position] != rune('r') {
								goto l53
							}
							position++
							if buffer[position] != rune('e') {
								goto l53
							}
							position++
							if buffer[position] != rune('a') {
								goto l53
							}
							position++
							if buffer[position] != rune('t') {
								goto l53
							}
							position++
							if buffer[position] != rune('e') {
								goto l53
							}
							position++
							goto l52
						l53:
							position, tokenIndex = position52, tokenIndex52
							if buffer[position] != rune('d') {
								goto l54
							}
							position++
							if buffer[position] != rune('e') {
								goto l54
							}
							position++
							if buffer[position] != rune('l') {
								goto l54
							}
							position++
							if buffer[position] != rune('e') {
								goto l54
							}
							position++
							if buffer[position] != rune('t') {
								goto l54
							}
							position++
							if buffer[position] != rune('e') {
								goto l54
							}
							position++
							goto l52
						l54:
							position, tokenIndex = position52, tokenIndex52
							if buffer[position] != rune('s') {
								goto l55
							}
							position++
							if buffer[position] != rune('t') {
								goto l55
							}
							position++
							if buffer[position] != rune('a') {
								goto l55
							}
							position++
							if buffer[position] != rune('r') {
								goto l55
							}
							position++
							if buffer[position] != rune('t') {
								goto l55
							}
							position++
							goto l52
						l55:
							position, tokenIndex = position52, tokenIndex52
							{
								switch buffer[position] {
								case 'd':
									if buffer[position] != rune('d') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('a') {
										goto l48
									}
									position++
									if buffer[position] != rune('c') {
										goto l48
									}
									position++
									if buffer[position] != rune('h') {
										goto l48
									}
									position++
									break
								case 'c':
									if buffer[position] != rune('c') {
										goto l48
									}
									position++
									if buffer[position] != rune('h') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('c') {
										goto l48
									}
									position++
									if buffer[position] != rune('k') {
										goto l48
									}
									position++
									break
								case 'a':
									if buffer[position] != rune('a') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('a') {
										goto l48
									}
									position++
									if buffer[position] != rune('c') {
										goto l48
									}
									position++
									if buffer[position] != rune('h') {
										goto l48
									}
									position++
									break
								case 'u':
									if buffer[position] != rune('u') {
										goto l48
									}
									position++
									if buffer[position] != rune('p') {
										goto l48
									}
									position++
									if buffer[position] != rune('d') {
										goto l48
									}
									position++
									if buffer[position] != rune('a') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									break
								case 's':
									if buffer[position] != rune('s') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('o') {
										goto l48
									}
									position++
									if buffer[position] != rune('p') {
										goto l48
									}
									position++
									break
								default:
									if buffer[position] != rune('n') {
										goto l48
									}
									position++
									if buffer[position] != rune('o') {
										goto l48
									}
									position++
									if buffer[position] != rune('n') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									break
								}
							}

						}
					l52:
						add(ruleAction, position51)
					}
					add(rulePegText, position50)
				}
				{
					add(ruleAction1, position)
				}
				if !_rules[ruleMustWhiteSpacing]() {
					goto l48
				}
				{
					position58 := position
					{
						position59 := position
						{
							position60, tokenIndex60 := position, tokenIndex
							if buffer[position] != rune('v') {
								goto l61
							}
							position++
							if buffer[position] != rune('p') {
								goto l61
							}
							position++
							if buffer[position] != rune('c') {
								goto l61
							}
							position++
							goto l60
						l61:
							position, tokenIndex = position60, tokenIndex60
							if buffer[position] != rune('s') {
								goto l62
							}
							position++
							if buffer[position] != rune('u') {
								goto l62
							}
							position++
							if buffer[position] != rune('b') {
								goto l62
							}
							position++
							if buffer[position] != rune('n') {
								goto l62
							}
							position++
							if buffer[position] != rune('e') {
								goto l62
							}
							position++
							if buffer[position] != rune('t') {
								goto l62
							}
							position++
							goto l60
						l62:
							position, tokenIndex = position60, tokenIndex60
							if buffer[position] != rune('i') {
								goto l63
							}
							position++
							if buffer[position] != rune('n') {
								goto l63
							}
							position++
							if buffer[position] != rune('s') {
								goto l63
							}
							position++
							if buffer[position] != rune('t') {
								goto l63
							}
							position++
							if buffer[position] != rune('a') {
								goto l63
							}
							position++
							if buffer[position] != rune('n') {
								goto l63
							}
							position++
							if buffer[position] != rune('c') {
								goto l63
							}
							position++
							if buffer[position] != rune('e') {
								goto l63
							}
							position++
							goto l60
						l63:
							position, tokenIndex = position60, tokenIndex60
							if buffer[position] != rune('t') {
								goto l64
							}
							position++
							if buffer[position] != rune('a') {
								goto l64
							}
							position++
							if buffer[position] != rune('g') {
								goto l64
							}
							position++
							goto l60
						l64:
							position, tokenIndex = position60, tokenIndex60
							if buffer[position] != rune('r') {
								goto l65
							}
							position++
							if buffer[position] != rune('o') {
								goto l65
							}
							position++
							if buffer[position] != rune('l') {
								goto l65
							}
							position++
							if buffer[position] != rune('e') {
								goto l65
							}
							position++
							goto l60
						l65:
							position, tokenIndex = position60, tokenIndex60
							if buffer[position] != rune('s') {
								goto l66
							}
							position++
							if buffer[position] != rune('e') {
								goto l66
							}
							position++
							if buffer[position] != rune('c') {
								goto l66
							}
							position++
							if buffer[position] != rune('u') {
								goto l66
							}
							position++
							if buffer[position] != rune('r') {
								goto l66
							}
							position++
							if buffer[position] != rune('i') {
								goto l66
							}
							position++
							if buffer[position] != rune('t') {
								goto l66
							}
							position++
							if buffer[position] != rune('y') {
								goto l66
							}
							position++
							if buffer[position] != rune('g') {
								goto l66
							}
							position++
							if buffer[position] != rune('r') {
								goto l66
							}
							position++
							if buffer[position] != rune('o') {
								goto l66
							}
							position++
							if buffer[position] != rune('u') {
								goto l66
							}
							position++
							if buffer[position] != rune('p') {
								goto l66
							}
							position++
							goto l60
						l66:
							position, tokenIndex = position60, tokenIndex60
							if buffer[position] != rune('r') {
								goto l67
							}
							position++
							if buffer[position] != rune('o') {
								goto l67
							}
							position++
							if buffer[position] != rune('u') {
								goto l67
							}
							position++
							if buffer[position] != rune('t') {
								goto l67
							}
							position++
							if buffer[position] != rune('e') {
								goto l67
							}
							position++
							if buffer[position] != rune('t') {
								goto l67
							}
							position++
							if buffer[position] != rune('a') {
								goto l67
							}
							position++
							if buffer[position] != rune('b') {
								goto l67
							}
							position++
							if buffer[position] != rune('l') {
								goto l67
							}
							position++
							if buffer[position] != rune('e') {
								goto l67
							}
							position++
							goto l60
						l67:
							position, tokenIndex = position60, tokenIndex60
							if buffer[position] != rune('s') {
								goto l68
							}
							position++
							if buffer[position] != rune('t') {
								goto l68
							}
							position++
							if buffer[position] != rune('o') {
								goto l68
							}
							position++
							if buffer[position] != rune('r') {
								goto l68
							}
							position++
							if buffer[position] != rune('a') {
								goto l68
							}
							position++
							if buffer[position] != rune('g') {
								goto l68
							}
							position++
							if buffer[position] != rune('e') {
								goto l68
							}
							position++
							if buffer[position] != rune('o') {
								goto l68
							}
							position++
							if buffer[position] != rune('b') {
								goto l68
							}
							position++
							if buffer[position] != rune('j') {
								goto l68
							}
							position++
							if buffer[position] != rune('e') {
								goto l68
							}
							position++
							if buffer[position] != rune('c') {
								goto l68
							}
							position++
							if buffer[position] != rune('t') {
								goto l68
							}
							position++
							goto l60
						l68:
							position, tokenIndex = position60, tokenIndex60
							if buffer[position] != rune('l') {
								goto l69
							}
							position++
							if buffer[position] != rune('o') {
								goto l69
							}
							position++
							if buffer[position] != rune('a') {
								goto l69
							}
							position++
							if buffer[position] != rune('d') {
								goto l69
							}
							position++
							if buffer[position] != rune('b') {
								goto l69
							}
							position++
							if buffer[position] != rune('a') {
								goto l69
							}
							position++
							if buffer[position] != rune('l') {
								goto l69
							}
							position++
							if buffer[position] != rune('a') {
								goto l69
							}
							position++
							if buffer[position] != rune('n') {
								goto l69
							}
							position++
							if buffer[position] != rune('c') {
								goto l69
							}
							position++
							if buffer[position] != rune('e') {
								goto l69
							}
							position++
							if buffer[position] != rune('r') {
								goto l69
							}
							position++
							goto l60
						l69:
							position, tokenIndex = position60, tokenIndex60
							{
								switch buffer[position] {
								case 'l':
									if buffer[position] != rune('l') {
										goto l48
									}
									position++
									if buffer[position] != rune('i') {
										goto l48
									}
									position++
									if buffer[position] != rune('s') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('n') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('r') {
										goto l48
									}
									position++
									break
								case 'q':
									if buffer[position] != rune('q') {
										goto l48
									}
									position++
									if buffer[position] != rune('u') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('u') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									break
								case 't':
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('o') {
										goto l48
									}
									position++
									if buffer[position] != rune('p') {
										goto l48
									}
									position++
									if buffer[position] != rune('i') {
										goto l48
									}
									position++
									if buffer[position] != rune('c') {
										goto l48
									}
									position++
									break
								case 's':
									if buffer[position] != rune('s') {
										goto l48
									}
									position++
									if buffer[position] != rune('u') {
										goto l48
									}
									position++
									if buffer[position] != rune('b') {
										goto l48
									}
									position++
									if buffer[position] != rune('s') {
										goto l48
									}
									position++
									if buffer[position] != rune('c') {
										goto l48
									}
									position++
									if buffer[position] != rune('r') {
										goto l48
									}
									position++
									if buffer[position] != rune('i') {
										goto l48
									}
									position++
									if buffer[position] != rune('p') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('i') {
										goto l48
									}
									position++
									if buffer[position] != rune('o') {
										goto l48
									}
									position++
									if buffer[position] != rune('n') {
										goto l48
									}
									position++
									break
								case 'b':
									if buffer[position] != rune('b') {
										goto l48
									}
									position++
									if buffer[position] != rune('u') {
										goto l48
									}
									position++
									if buffer[position] != rune('c') {
										goto l48
									}
									position++
									if buffer[position] != rune('k') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									break
								case 'r':
									if buffer[position] != rune('r') {
										goto l48
									}
									position++
									if buffer[position] != rune('o') {
										goto l48
									}
									position++
									if buffer[position] != rune('u') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									break
								case 'i':
									if buffer[position] != rune('i') {
										goto l48
									}
									position++
									if buffer[position] != rune('n') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('r') {
										goto l48
									}
									position++
									if buffer[position] != rune('n') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('g') {
										goto l48
									}
									position++
									if buffer[position] != rune('a') {
										goto l48
									}
									position++
									if buffer[position] != rune('t') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('w') {
										goto l48
									}
									position++
									if buffer[position] != rune('a') {
										goto l48
									}
									position++
									if buffer[position] != rune('y') {
										goto l48
									}
									position++
									break
								case 'k':
									if buffer[position] != rune('k') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('y') {
										goto l48
									}
									position++
									if buffer[position] != rune('p') {
										goto l48
									}
									position++
									if buffer[position] != rune('a') {
										goto l48
									}
									position++
									if buffer[position] != rune('i') {
										goto l48
									}
									position++
									if buffer[position] != rune('r') {
										goto l48
									}
									position++
									break
								case 'p':
									if buffer[position] != rune('p') {
										goto l48
									}
									position++
									if buffer[position] != rune('o') {
										goto l48
									}
									position++
									if buffer[position] != rune('l') {
										goto l48
									}
									position++
									if buffer[position] != rune('i') {
										goto l48
									}
									position++
									if buffer[position] != rune('c') {
										goto l48
									}
									position++
									if buffer[position] != rune('y') {
										goto l48
									}
									position++
									break
								case 'g':
									if buffer[position] != rune('g') {
										goto l48
									}
									position++
									if buffer[position] != rune('r') {
										goto l48
									}
									position++
									if buffer[position] != rune('o') {
										goto l48
									}
									position++
									if buffer[position] != rune('u') {
										goto l48
									}
									position++
									if buffer[position] != rune('p') {
										goto l48
									}
									position++
									break
								case 'u':
									if buffer[position] != rune('u') {
										goto l48
									}
									position++
									if buffer[position] != rune('s') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									if buffer[position] != rune('r') {
										goto l48
									}
									position++
									break
								case 'v':
									if buffer[position] != rune('v') {
										goto l48
									}
									position++
									if buffer[position] != rune('o') {
										goto l48
									}
									position++
									if buffer[position] != rune('l') {
										goto l48
									}
									position++
									if buffer[position] != rune('u') {
										goto l48
									}
									position++
									if buffer[position] != rune('m') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									break
								default:
									if buffer[position] != rune('n') {
										goto l48
									}
									position++
									if buffer[position] != rune('o') {
										goto l48
									}
									position++
									if buffer[position] != rune('n') {
										goto l48
									}
									position++
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									break
								}
							}

						}
					l60:
						add(ruleEntity, position59)
					}
					add(rulePegText, position58)
				}
				{
					add(ruleAction2, position)
				}
				{
					position72, tokenIndex72 := position, tokenIndex
					if !_rules[ruleMustWhiteSpacing]() {
						goto l72
					}
					{
						position74 := position
						{
							position77 := position
							{
								position78 := position
								if !_rules[ruleIdentifier]() {
									goto l72
								}
								add(rulePegText, position78)
							}
							{
								add(ruleAction4, position)
							}
							if !_rules[ruleEqual]() {
								goto l72
							}
							{
								position80 := position
								{
									position81, tokenIndex81 := position, tokenIndex
									{
										position83 := position
										{
											position84 := position
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l82
											}
											position++
										l85:
											{
												position86, tokenIndex86 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l86
												}
												position++
												goto l85
											l86:
												position, tokenIndex = position86, tokenIndex86
											}
											if !matchDot() {
												goto l82
											}
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l82
											}
											position++
										l87:
											{
												position88, tokenIndex88 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l88
												}
												position++
												goto l87
											l88:
												position, tokenIndex = position88, tokenIndex88
											}
											if !matchDot() {
												goto l82
											}
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l82
											}
											position++
										l89:
											{
												position90, tokenIndex90 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l90
												}
												position++
												goto l89
											l90:
												position, tokenIndex = position90, tokenIndex90
											}
											if !matchDot() {
												goto l82
											}
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l82
											}
											position++
										l91:
											{
												position92, tokenIndex92 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l92
												}
												position++
												goto l91
											l92:
												position, tokenIndex = position92, tokenIndex92
											}
											if buffer[position] != rune('/') {
												goto l82
											}
											position++
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l82
											}
											position++
										l93:
											{
												position94, tokenIndex94 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l94
												}
												position++
												goto l93
											l94:
												position, tokenIndex = position94, tokenIndex94
											}
											add(ruleCidrValue, position84)
										}
										add(rulePegText, position83)
									}
									{
										add(ruleAction8, position)
									}
									goto l81
								l82:
									position, tokenIndex = position81, tokenIndex81
									{
										position97 := position
										{
											position98 := position
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l96
											}
											position++
										l99:
											{
												position100, tokenIndex100 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l100
												}
												position++
												goto l99
											l100:
												position, tokenIndex = position100, tokenIndex100
											}
											if !matchDot() {
												goto l96
											}
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l96
											}
											position++
										l101:
											{
												position102, tokenIndex102 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l102
												}
												position++
												goto l101
											l102:
												position, tokenIndex = position102, tokenIndex102
											}
											if !matchDot() {
												goto l96
											}
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l96
											}
											position++
										l103:
											{
												position104, tokenIndex104 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l104
												}
												position++
												goto l103
											l104:
												position, tokenIndex = position104, tokenIndex104
											}
											if !matchDot() {
												goto l96
											}
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l96
											}
											position++
										l105:
											{
												position106, tokenIndex106 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l106
												}
												position++
												goto l105
											l106:
												position, tokenIndex = position106, tokenIndex106
											}
											add(ruleIpValue, position98)
										}
										add(rulePegText, position97)
									}
									{
										add(ruleAction9, position)
									}
									goto l81
								l96:
									position, tokenIndex = position81, tokenIndex81
									{
										position109 := position
										{
											position110 := position
											if !_rules[ruleStringValue]() {
												goto l108
											}
											if !_rules[ruleWhiteSpacing]() {
												goto l108
											}
											if buffer[position] != rune(',') {
												goto l108
											}
											position++
											if !_rules[ruleWhiteSpacing]() {
												goto l108
											}
										l111:
											{
												position112, tokenIndex112 := position, tokenIndex
												if !_rules[ruleStringValue]() {
													goto l112
												}
												if !_rules[ruleWhiteSpacing]() {
													goto l112
												}
												if buffer[position] != rune(',') {
													goto l112
												}
												position++
												if !_rules[ruleWhiteSpacing]() {
													goto l112
												}
												goto l111
											l112:
												position, tokenIndex = position112, tokenIndex112
											}
											if !_rules[ruleStringValue]() {
												goto l108
											}
											add(ruleCSVValue, position110)
										}
										add(rulePegText, position109)
									}
									{
										add(ruleAction10, position)
									}
									goto l81
								l108:
									position, tokenIndex = position81, tokenIndex81
									{
										position115 := position
										{
											position116 := position
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l114
											}
											position++
										l117:
											{
												position118, tokenIndex118 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l118
												}
												position++
												goto l117
											l118:
												position, tokenIndex = position118, tokenIndex118
											}
											if buffer[position] != rune('-') {
												goto l114
											}
											position++
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l114
											}
											position++
										l119:
											{
												position120, tokenIndex120 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l120
												}
												position++
												goto l119
											l120:
												position, tokenIndex = position120, tokenIndex120
											}
											add(ruleIntRangeValue, position116)
										}
										add(rulePegText, position115)
									}
									{
										add(ruleAction11, position)
									}
									goto l81
								l114:
									position, tokenIndex = position81, tokenIndex81
									{
										position123 := position
										{
											position124 := position
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l122
											}
											position++
										l125:
											{
												position126, tokenIndex126 := position, tokenIndex
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l126
												}
												position++
												goto l125
											l126:
												position, tokenIndex = position126, tokenIndex126
											}
											add(ruleIntValue, position124)
										}
										add(rulePegText, position123)
									}
									{
										add(ruleAction12, position)
									}
									goto l81
								l122:
									position, tokenIndex = position81, tokenIndex81
									{
										switch buffer[position] {
										case '$':
											{
												position129 := position
												if buffer[position] != rune('$') {
													goto l72
												}
												position++
												{
													position130 := position
													if !_rules[ruleIdentifier]() {
														goto l72
													}
													add(rulePegText, position130)
												}
												add(ruleRefValue, position129)
											}
											{
												add(ruleAction7, position)
											}
											break
										case '@':
											{
												position132 := position
												{
													position133 := position
													if buffer[position] != rune('@') {
														goto l72
													}
													position++
													if !_rules[ruleStringValue]() {
														goto l72
													}
													add(rulePegText, position133)
												}
												add(ruleAliasValue, position132)
											}
											{
												add(ruleAction6, position)
											}
											break
										case '{':
											{
												position135 := position
												if buffer[position] != rune('{') {
													goto l72
												}
												position++
												if !_rules[ruleWhiteSpacing]() {
													goto l72
												}
												{
													position136 := position
													if !_rules[ruleIdentifier]() {
														goto l72
													}
													add(rulePegText, position136)
												}
												if !_rules[ruleWhiteSpacing]() {
													goto l72
												}
												if buffer[position] != rune('}') {
													goto l72
												}
												position++
												add(ruleHoleValue, position135)
											}
											{
												add(ruleAction5, position)
											}
											break
										default:
											{
												position138 := position
												if !_rules[ruleStringValue]() {
													goto l72
												}
												add(rulePegText, position138)
											}
											{
												add(ruleAction13, position)
											}
											break
										}
									}

								}
							l81:
								add(ruleValue, position80)
							}
							if !_rules[ruleWhiteSpacing]() {
								goto l72
							}
							add(ruleParam, position77)
						}
					l75:
						{
							position76, tokenIndex76 := position, tokenIndex
							{
								position140 := position
								{
									position141 := position
									if !_rules[ruleIdentifier]() {
										goto l76
									}
									add(rulePegText, position141)
								}
								{
									add(ruleAction4, position)
								}
								if !_rules[ruleEqual]() {
									goto l76
								}
								{
									position143 := position
									{
										position144, tokenIndex144 := position, tokenIndex
										{
											position146 := position
											{
												position147 := position
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l145
												}
												position++
											l148:
												{
													position149, tokenIndex149 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l149
													}
													position++
													goto l148
												l149:
													position, tokenIndex = position149, tokenIndex149
												}
												if !matchDot() {
													goto l145
												}
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l145
												}
												position++
											l150:
												{
													position151, tokenIndex151 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l151
													}
													position++
													goto l150
												l151:
													position, tokenIndex = position151, tokenIndex151
												}
												if !matchDot() {
													goto l145
												}
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l145
												}
												position++
											l152:
												{
													position153, tokenIndex153 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l153
													}
													position++
													goto l152
												l153:
													position, tokenIndex = position153, tokenIndex153
												}
												if !matchDot() {
													goto l145
												}
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l145
												}
												position++
											l154:
												{
													position155, tokenIndex155 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l155
													}
													position++
													goto l154
												l155:
													position, tokenIndex = position155, tokenIndex155
												}
												if buffer[position] != rune('/') {
													goto l145
												}
												position++
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l145
												}
												position++
											l156:
												{
													position157, tokenIndex157 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l157
													}
													position++
													goto l156
												l157:
													position, tokenIndex = position157, tokenIndex157
												}
												add(ruleCidrValue, position147)
											}
											add(rulePegText, position146)
										}
										{
											add(ruleAction8, position)
										}
										goto l144
									l145:
										position, tokenIndex = position144, tokenIndex144
										{
											position160 := position
											{
												position161 := position
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l159
												}
												position++
											l162:
												{
													position163, tokenIndex163 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l163
													}
													position++
													goto l162
												l163:
													position, tokenIndex = position163, tokenIndex163
												}
												if !matchDot() {
													goto l159
												}
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l159
												}
												position++
											l164:
												{
													position165, tokenIndex165 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l165
													}
													position++
													goto l164
												l165:
													position, tokenIndex = position165, tokenIndex165
												}
												if !matchDot() {
													goto l159
												}
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l159
												}
												position++
											l166:
												{
													position167, tokenIndex167 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l167
													}
													position++
													goto l166
												l167:
													position, tokenIndex = position167, tokenIndex167
												}
												if !matchDot() {
													goto l159
												}
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l159
												}
												position++
											l168:
												{
													position169, tokenIndex169 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l169
													}
													position++
													goto l168
												l169:
													position, tokenIndex = position169, tokenIndex169
												}
												add(ruleIpValue, position161)
											}
											add(rulePegText, position160)
										}
										{
											add(ruleAction9, position)
										}
										goto l144
									l159:
										position, tokenIndex = position144, tokenIndex144
										{
											position172 := position
											{
												position173 := position
												if !_rules[ruleStringValue]() {
													goto l171
												}
												if !_rules[ruleWhiteSpacing]() {
													goto l171
												}
												if buffer[position] != rune(',') {
													goto l171
												}
												position++
												if !_rules[ruleWhiteSpacing]() {
													goto l171
												}
											l174:
												{
													position175, tokenIndex175 := position, tokenIndex
													if !_rules[ruleStringValue]() {
														goto l175
													}
													if !_rules[ruleWhiteSpacing]() {
														goto l175
													}
													if buffer[position] != rune(',') {
														goto l175
													}
													position++
													if !_rules[ruleWhiteSpacing]() {
														goto l175
													}
													goto l174
												l175:
													position, tokenIndex = position175, tokenIndex175
												}
												if !_rules[ruleStringValue]() {
													goto l171
												}
												add(ruleCSVValue, position173)
											}
											add(rulePegText, position172)
										}
										{
											add(ruleAction10, position)
										}
										goto l144
									l171:
										position, tokenIndex = position144, tokenIndex144
										{
											position178 := position
											{
												position179 := position
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l177
												}
												position++
											l180:
												{
													position181, tokenIndex181 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l181
													}
													position++
													goto l180
												l181:
													position, tokenIndex = position181, tokenIndex181
												}
												if buffer[position] != rune('-') {
													goto l177
												}
												position++
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l177
												}
												position++
											l182:
												{
													position183, tokenIndex183 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l183
													}
													position++
													goto l182
												l183:
													position, tokenIndex = position183, tokenIndex183
												}
												add(ruleIntRangeValue, position179)
											}
											add(rulePegText, position178)
										}
										{
											add(ruleAction11, position)
										}
										goto l144
									l177:
										position, tokenIndex = position144, tokenIndex144
										{
											position186 := position
											{
												position187 := position
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l185
												}
												position++
											l188:
												{
													position189, tokenIndex189 := position, tokenIndex
													if c := buffer[position]; c < rune('0') || c > rune('9') {
														goto l189
													}
													position++
													goto l188
												l189:
													position, tokenIndex = position189, tokenIndex189
												}
												add(ruleIntValue, position187)
											}
											add(rulePegText, position186)
										}
										{
											add(ruleAction12, position)
										}
										goto l144
									l185:
										position, tokenIndex = position144, tokenIndex144
										{
											switch buffer[position] {
											case '$':
												{
													position192 := position
													if buffer[position] != rune('$') {
														goto l76
													}
													position++
													{
														position193 := position
														if !_rules[ruleIdentifier]() {
															goto l76
														}
														add(rulePegText, position193)
													}
													add(ruleRefValue, position192)
												}
												{
													add(ruleAction7, position)
												}
												break
											case '@':
												{
													position195 := position
													{
														position196 := position
														if buffer[position] != rune('@') {
															goto l76
														}
														position++
														if !_rules[ruleStringValue]() {
															goto l76
														}
														add(rulePegText, position196)
													}
													add(ruleAliasValue, position195)
												}
												{
													add(ruleAction6, position)
												}
												break
											case '{':
												{
													position198 := position
													if buffer[position] != rune('{') {
														goto l76
													}
													position++
													if !_rules[ruleWhiteSpacing]() {
														goto l76
													}
													{
														position199 := position
														if !_rules[ruleIdentifier]() {
															goto l76
														}
														add(rulePegText, position199)
													}
													if !_rules[ruleWhiteSpacing]() {
														goto l76
													}
													if buffer[position] != rune('}') {
														goto l76
													}
													position++
													add(ruleHoleValue, position198)
												}
												{
													add(ruleAction5, position)
												}
												break
											default:
												{
													position201 := position
													if !_rules[ruleStringValue]() {
														goto l76
													}
													add(rulePegText, position201)
												}
												{
													add(ruleAction13, position)
												}
												break
											}
										}

									}
								l144:
									add(ruleValue, position143)
								}
								if !_rules[ruleWhiteSpacing]() {
									goto l76
								}
								add(ruleParam, position140)
							}
							goto l75
						l76:
							position, tokenIndex = position76, tokenIndex76
						}
						add(ruleParams, position74)
					}
					goto l73
				l72:
					position, tokenIndex = position72, tokenIndex72
				}
			l73:
				{
					add(ruleAction3, position)
				}
				add(ruleExpr, position49)
			}
			return true
		l48:
			position, tokenIndex = position48, tokenIndex48
			return false
		},
		/* 6 Params <- <Param+> */
		nil,
		/* 7 Param <- <(<Identifier> Action4 Equal Value WhiteSpacing)> */
		nil,
		/* 8 Identifier <- <((&('.') '.') | (&('_') '_') | (&('-') '-') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> */
		func() bool {
			position206, tokenIndex206 := position, tokenIndex
			{
				position207 := position
				{
					switch buffer[position] {
					case '.':
						if buffer[position] != rune('.') {
							goto l206
						}
						position++
						break
					case '_':
						if buffer[position] != rune('_') {
							goto l206
						}
						position++
						break
					case '-':
						if buffer[position] != rune('-') {
							goto l206
						}
						position++
						break
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l206
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l206
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l206
						}
						position++
						break
					}
				}

			l208:
				{
					position209, tokenIndex209 := position, tokenIndex
					{
						switch buffer[position] {
						case '.':
							if buffer[position] != rune('.') {
								goto l209
							}
							position++
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l209
							}
							position++
							break
						case '-':
							if buffer[position] != rune('-') {
								goto l209
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l209
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l209
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l209
							}
							position++
							break
						}
					}

					goto l208
				l209:
					position, tokenIndex = position209, tokenIndex209
				}
				add(ruleIdentifier, position207)
			}
			return true
		l206:
			position, tokenIndex = position206, tokenIndex206
			return false
		},
		/* 9 Value <- <((<CidrValue> Action8) / (<IpValue> Action9) / (<CSVValue> Action10) / (<IntRangeValue> Action11) / (<IntValue> Action12) / ((&('$') (RefValue Action7)) | (&('@') (AliasValue Action6)) | (&('{') (HoleValue Action5)) | (&('-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') (<StringValue> Action13))))> */
		nil,
		/* 10 StringValue <- <((&('/') '/') | (&(':') ':') | (&('_') '_') | (&('.') '.') | (&('-') '-') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> */
		func() bool {
			position213, tokenIndex213 := position, tokenIndex
			{
				position214 := position
				{
					switch buffer[position] {
					case '/':
						if buffer[position] != rune('/') {
							goto l213
						}
						position++
						break
					case ':':
						if buffer[position] != rune(':') {
							goto l213
						}
						position++
						break
					case '_':
						if buffer[position] != rune('_') {
							goto l213
						}
						position++
						break
					case '.':
						if buffer[position] != rune('.') {
							goto l213
						}
						position++
						break
					case '-':
						if buffer[position] != rune('-') {
							goto l213
						}
						position++
						break
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l213
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l213
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l213
						}
						position++
						break
					}
				}

			l215:
				{
					position216, tokenIndex216 := position, tokenIndex
					{
						switch buffer[position] {
						case '/':
							if buffer[position] != rune('/') {
								goto l216
							}
							position++
							break
						case ':':
							if buffer[position] != rune(':') {
								goto l216
							}
							position++
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l216
							}
							position++
							break
						case '.':
							if buffer[position] != rune('.') {
								goto l216
							}
							position++
							break
						case '-':
							if buffer[position] != rune('-') {
								goto l216
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l216
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l216
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l216
							}
							position++
							break
						}
					}

					goto l215
				l216:
					position, tokenIndex = position216, tokenIndex216
				}
				add(ruleStringValue, position214)
			}
			return true
		l213:
			position, tokenIndex = position213, tokenIndex213
			return false
		},
		/* 11 CSVValue <- <((StringValue WhiteSpacing ',' WhiteSpacing)+ StringValue)> */
		nil,
		/* 12 CidrValue <- <([0-9]+ . [0-9]+ . [0-9]+ . [0-9]+ '/' [0-9]+)> */
		nil,
		/* 13 IpValue <- <([0-9]+ . [0-9]+ . [0-9]+ . [0-9]+)> */
		nil,
		/* 14 IntValue <- <[0-9]+> */
		nil,
		/* 15 IntRangeValue <- <([0-9]+ '-' [0-9]+)> */
		nil,
		/* 16 RefValue <- <('$' <Identifier>)> */
		nil,
		/* 17 AliasValue <- <<('@' StringValue)>> */
		nil,
		/* 18 HoleValue <- <('{' WhiteSpacing <Identifier> WhiteSpacing '}')> */
		nil,
		/* 19 Comment <- <(('#' (!EndOfLine .)*) / ('/' '/' (!EndOfLine .)* Action14))> */
		nil,
		/* 20 Spacing <- <Space*> */
		func() bool {
			{
				position229 := position
			l230:
				{
					position231, tokenIndex231 := position, tokenIndex
					{
						position232 := position
						{
							position233, tokenIndex233 := position, tokenIndex
							if !_rules[ruleWhitespace]() {
								goto l234
							}
							goto l233
						l234:
							position, tokenIndex = position233, tokenIndex233
							if !_rules[ruleEndOfLine]() {
								goto l231
							}
						}
					l233:
						add(ruleSpace, position232)
					}
					goto l230
				l231:
					position, tokenIndex = position231, tokenIndex231
				}
				add(ruleSpacing, position229)
			}
			return true
		},
		/* 21 WhiteSpacing <- <Whitespace*> */
		func() bool {
			{
				position236 := position
			l237:
				{
					position238, tokenIndex238 := position, tokenIndex
					if !_rules[ruleWhitespace]() {
						goto l238
					}
					goto l237
				l238:
					position, tokenIndex = position238, tokenIndex238
				}
				add(ruleWhiteSpacing, position236)
			}
			return true
		},
		/* 22 MustWhiteSpacing <- <Whitespace+> */
		func() bool {
			position239, tokenIndex239 := position, tokenIndex
			{
				position240 := position
				if !_rules[ruleWhitespace]() {
					goto l239
				}
			l241:
				{
					position242, tokenIndex242 := position, tokenIndex
					if !_rules[ruleWhitespace]() {
						goto l242
					}
					goto l241
				l242:
					position, tokenIndex = position242, tokenIndex242
				}
				add(ruleMustWhiteSpacing, position240)
			}
			return true
		l239:
			position, tokenIndex = position239, tokenIndex239
			return false
		},
		/* 23 Equal <- <(Spacing '=' Spacing)> */
		func() bool {
			position243, tokenIndex243 := position, tokenIndex
			{
				position244 := position
				if !_rules[ruleSpacing]() {
					goto l243
				}
				if buffer[position] != rune('=') {
					goto l243
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l243
				}
				add(ruleEqual, position244)
			}
			return true
		l243:
			position, tokenIndex = position243, tokenIndex243
			return false
		},
		/* 24 Space <- <(Whitespace / EndOfLine)> */
		nil,
		/* 25 Whitespace <- <(' ' / '\t')> */
		func() bool {
			position246, tokenIndex246 := position, tokenIndex
			{
				position247 := position
				{
					position248, tokenIndex248 := position, tokenIndex
					if buffer[position] != rune(' ') {
						goto l249
					}
					position++
					goto l248
				l249:
					position, tokenIndex = position248, tokenIndex248
					if buffer[position] != rune('\t') {
						goto l246
					}
					position++
				}
			l248:
				add(ruleWhitespace, position247)
			}
			return true
		l246:
			position, tokenIndex = position246, tokenIndex246
			return false
		},
		/* 26 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position250, tokenIndex250 := position, tokenIndex
			{
				position251 := position
				{
					position252, tokenIndex252 := position, tokenIndex
					if buffer[position] != rune('\r') {
						goto l253
					}
					position++
					if buffer[position] != rune('\n') {
						goto l253
					}
					position++
					goto l252
				l253:
					position, tokenIndex = position252, tokenIndex252
					if buffer[position] != rune('\n') {
						goto l254
					}
					position++
					goto l252
				l254:
					position, tokenIndex = position252, tokenIndex252
					if buffer[position] != rune('\r') {
						goto l250
					}
					position++
				}
			l252:
				add(ruleEndOfLine, position251)
			}
			return true
		l250:
			position, tokenIndex = position250, tokenIndex250
			return false
		},
		/* 27 EndOfFile <- <!.> */
		nil,
		nil,
		/* 30 Action0 <- <{ p.addDeclarationIdentifier(text) }> */
		nil,
		/* 31 Action1 <- <{ p.addAction(text) }> */
		nil,
		/* 32 Action2 <- <{ p.addEntity(text) }> */
		nil,
		/* 33 Action3 <- <{ p.LineDone() }> */
		nil,
		/* 34 Action4 <- <{ p.addParamKey(text) }> */
		nil,
		/* 35 Action5 <- <{  p.addParamHoleValue(text) }> */
		nil,
		/* 36 Action6 <- <{  p.addParamValue(text) }> */
		nil,
		/* 37 Action7 <- <{  p.addParamRefValue(text) }> */
		nil,
		/* 38 Action8 <- <{ p.addParamCidrValue(text) }> */
		nil,
		/* 39 Action9 <- <{ p.addParamIpValue(text) }> */
		nil,
		/* 40 Action10 <- <{p.addCsvValue(text)}> */
		nil,
		/* 41 Action11 <- <{ p.addParamValue(text) }> */
		nil,
		/* 42 Action12 <- <{ p.addParamIntValue(text) }> */
		nil,
		/* 43 Action13 <- <{ p.addParamValue(text) }> */
		nil,
		/* 44 Action14 <- <{ p.LineDone() }> */
		nil,
	}
	p.rules = _rules
}
