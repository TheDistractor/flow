package fbpparse

import (
	/*"bytes"*/
	"fmt"
	"math"
	"sort"
	"strconv"
)

const END_SYMBOL rune = 4

/* The rule types inferred from the grammar are below. */
type Rule uint8

const (
	RuleUnknown Rule = iota
	Rulestart
	Ruleline
	RuleLineTerminator
	Rulecomment
	Ruleconnection
	Rulebridge
	Ruleleftlet
	Ruleiip
	Rulerightlet
	Rulenode
	Rulecomponent
	RulecompMeta
	Ruleport
	Ruleanychar
	Ruleiipchar
	Rule_
	Rule__

	RulePre_
	Rule_In_
	Rule_Suf
)

var Rul3s = [...]string{
	"Unknown",
	"start",
	"line",
	"LineTerminator",
	"comment",
	"connection",
	"bridge",
	"leftlet",
	"iip",
	"rightlet",
	"node",
	"component",
	"compMeta",
	"port",
	"anychar",
	"iipchar",
	"_",
	"__",

	"Pre_",
	"_In_",
	"_Suf",
}

type TokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule Rule, begin, end, next, depth int)
	Expand(index int) TokenTree
	Tokens() <-chan token32
	Error() []token32
	trim(length int)
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	Rule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) GetToken32() token32 {
	return token32{Rule: t.Rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", Rul3s[t.Rule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.Rule == RuleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type State16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) PreOrder() (<-chan State16, [][]token16) {
	s, ordered := make(chan State16, 6), t.Order()
	go func() {
		var states [8]State16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.Rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.Rule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{Rule: Rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{Rule: RulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.Rule != RuleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.Rule != RuleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{Rule: Rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", Rul3s[token.Rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", Rul3s[token.Rule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", Rul3s[token.Rule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule Rule, begin, end, depth, index int) {
	t.tree[index] = token16{Rule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	Rule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) GetToken32() token32 {
	return token32{Rule: t.Rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", Rul3s[t.Rule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.Rule == RuleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type State32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) PreOrder() (<-chan State32, [][]token32) {
	s, ordered := make(chan State32, 6), t.Order()
	go func() {
		var states [8]State32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.Rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.Rule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{Rule: Rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{Rule: RulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.Rule != RuleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.Rule != RuleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{Rule: Rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", Rul3s[token.Rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", Rul3s[token.Rule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", Rul3s[token.Rule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule Rule, begin, end, depth, index int) {
	t.tree[index] = token32{Rule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.GetToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type Fbp struct {
	Buffer string
	buffer []rune
	rules  [18]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	TokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
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
	p *Fbp
}

func (e *parseError) Error() string {
	tokens, error := e.p.TokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			Rul3s[token.Rule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *Fbp) PrintSyntaxTree() {
	p.TokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *Fbp) Highlighter() {
	p.TokenTree.PrintSyntax()
}

func (p *Fbp) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != END_SYMBOL {
		p.buffer = append(p.buffer, END_SYMBOL)
	}

	var tree TokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.TokenTree = tree
		if matches {
			p.TokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule Rule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != END_SYMBOL {
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

	rules = [...]func() bool{
		nil,
		/* 0 start <- <(line* _ !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					if !rules[Ruleline]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position3, tokenIndex3, depth3
				}
				if !rules[Rule_]() {
					goto l0
				}
				{
					position4, tokenIndex4, depth4 := position, tokenIndex, depth
					if !matchDot() {
						goto l4
					}
					goto l0
				l4:
					position, tokenIndex, depth = position4, tokenIndex4, depth4
				}
				depth--
				add(Rulestart, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 line <- <((_ (('e' / 'E') ('x' / 'X') ('p' / 'P') ('o' / 'O') ('r' / 'R') ('t' / 'T') '=') ([A-Z] / [a-z] / '.' / [0-9] / '_')+ ':' ([A-Z] / [0-9] / '_')+ _ LineTerminator?) / (_ (('i' / 'I') ('n' / 'N') ('p' / 'P') ('o' / 'O') ('r' / 'R') ('t' / 'T') '=') ([A-Z] / [a-z] / [0-9] / '_')+ '.' ([A-Z] / [0-9] / '_')+ ':' ([A-Z] / [0-9] / '_')+ _ LineTerminator?) / (_ (('o' / 'O') ('u' / 'U') ('t' / 'T') ('p' / 'P') ('o' / 'O') ('r' / 'R') ('t' / 'T') '=') ([A-Z] / [a-z] / [0-9] / '_')+ '.' ([A-Z] / [0-9] / '_')+ ':' ([A-Z] / [0-9] / '_')+ _ LineTerminator?) / (comment ('\n' / '\r')?) / (_ ('\n' / '\r')) / (_ connection _ LineTerminator?))> */
		func() bool {
			position5, tokenIndex5, depth5 := position, tokenIndex, depth
			{
				position6 := position
				depth++
				{
					position7, tokenIndex7, depth7 := position, tokenIndex, depth
					if !rules[Rule_]() {
						goto l8
					}
					{
						position9, tokenIndex9, depth9 := position, tokenIndex, depth
						if buffer[position] != rune('e') {
							goto l10
						}
						position++
						goto l9
					l10:
						position, tokenIndex, depth = position9, tokenIndex9, depth9
						if buffer[position] != rune('E') {
							goto l8
						}
						position++
					}
				l9:
					{
						position11, tokenIndex11, depth11 := position, tokenIndex, depth
						if buffer[position] != rune('x') {
							goto l12
						}
						position++
						goto l11
					l12:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if buffer[position] != rune('X') {
							goto l8
						}
						position++
					}
				l11:
					{
						position13, tokenIndex13, depth13 := position, tokenIndex, depth
						if buffer[position] != rune('p') {
							goto l14
						}
						position++
						goto l13
					l14:
						position, tokenIndex, depth = position13, tokenIndex13, depth13
						if buffer[position] != rune('P') {
							goto l8
						}
						position++
					}
				l13:
					{
						position15, tokenIndex15, depth15 := position, tokenIndex, depth
						if buffer[position] != rune('o') {
							goto l16
						}
						position++
						goto l15
					l16:
						position, tokenIndex, depth = position15, tokenIndex15, depth15
						if buffer[position] != rune('O') {
							goto l8
						}
						position++
					}
				l15:
					{
						position17, tokenIndex17, depth17 := position, tokenIndex, depth
						if buffer[position] != rune('r') {
							goto l18
						}
						position++
						goto l17
					l18:
						position, tokenIndex, depth = position17, tokenIndex17, depth17
						if buffer[position] != rune('R') {
							goto l8
						}
						position++
					}
				l17:
					{
						position19, tokenIndex19, depth19 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l20
						}
						position++
						goto l19
					l20:
						position, tokenIndex, depth = position19, tokenIndex19, depth19
						if buffer[position] != rune('T') {
							goto l8
						}
						position++
					}
				l19:
					if buffer[position] != rune('=') {
						goto l8
					}
					position++
					{
						position23, tokenIndex23, depth23 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l24
						}
						position++
						goto l23
					l24:
						position, tokenIndex, depth = position23, tokenIndex23, depth23
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l25
						}
						position++
						goto l23
					l25:
						position, tokenIndex, depth = position23, tokenIndex23, depth23
						if buffer[position] != rune('.') {
							goto l26
						}
						position++
						goto l23
					l26:
						position, tokenIndex, depth = position23, tokenIndex23, depth23
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l27
						}
						position++
						goto l23
					l27:
						position, tokenIndex, depth = position23, tokenIndex23, depth23
						if buffer[position] != rune('_') {
							goto l8
						}
						position++
					}
				l23:
				l21:
					{
						position22, tokenIndex22, depth22 := position, tokenIndex, depth
						{
							position28, tokenIndex28, depth28 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l29
							}
							position++
							goto l28
						l29:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l30
							}
							position++
							goto l28
						l30:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
							if buffer[position] != rune('.') {
								goto l31
							}
							position++
							goto l28
						l31:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l32
							}
							position++
							goto l28
						l32:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
							if buffer[position] != rune('_') {
								goto l22
							}
							position++
						}
					l28:
						goto l21
					l22:
						position, tokenIndex, depth = position22, tokenIndex22, depth22
					}
					if buffer[position] != rune(':') {
						goto l8
					}
					position++
					{
						position35, tokenIndex35, depth35 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l36
						}
						position++
						goto l35
					l36:
						position, tokenIndex, depth = position35, tokenIndex35, depth35
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l37
						}
						position++
						goto l35
					l37:
						position, tokenIndex, depth = position35, tokenIndex35, depth35
						if buffer[position] != rune('_') {
							goto l8
						}
						position++
					}
				l35:
				l33:
					{
						position34, tokenIndex34, depth34 := position, tokenIndex, depth
						{
							position38, tokenIndex38, depth38 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l39
							}
							position++
							goto l38
						l39:
							position, tokenIndex, depth = position38, tokenIndex38, depth38
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l40
							}
							position++
							goto l38
						l40:
							position, tokenIndex, depth = position38, tokenIndex38, depth38
							if buffer[position] != rune('_') {
								goto l34
							}
							position++
						}
					l38:
						goto l33
					l34:
						position, tokenIndex, depth = position34, tokenIndex34, depth34
					}
					if !rules[Rule_]() {
						goto l8
					}
					{
						position41, tokenIndex41, depth41 := position, tokenIndex, depth
						if !rules[RuleLineTerminator]() {
							goto l41
						}
						goto l42
					l41:
						position, tokenIndex, depth = position41, tokenIndex41, depth41
					}
				l42:
					goto l7
				l8:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
					if !rules[Rule_]() {
						goto l43
					}
					{
						position44, tokenIndex44, depth44 := position, tokenIndex, depth
						if buffer[position] != rune('i') {
							goto l45
						}
						position++
						goto l44
					l45:
						position, tokenIndex, depth = position44, tokenIndex44, depth44
						if buffer[position] != rune('I') {
							goto l43
						}
						position++
					}
				l44:
					{
						position46, tokenIndex46, depth46 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l47
						}
						position++
						goto l46
					l47:
						position, tokenIndex, depth = position46, tokenIndex46, depth46
						if buffer[position] != rune('N') {
							goto l43
						}
						position++
					}
				l46:
					{
						position48, tokenIndex48, depth48 := position, tokenIndex, depth
						if buffer[position] != rune('p') {
							goto l49
						}
						position++
						goto l48
					l49:
						position, tokenIndex, depth = position48, tokenIndex48, depth48
						if buffer[position] != rune('P') {
							goto l43
						}
						position++
					}
				l48:
					{
						position50, tokenIndex50, depth50 := position, tokenIndex, depth
						if buffer[position] != rune('o') {
							goto l51
						}
						position++
						goto l50
					l51:
						position, tokenIndex, depth = position50, tokenIndex50, depth50
						if buffer[position] != rune('O') {
							goto l43
						}
						position++
					}
				l50:
					{
						position52, tokenIndex52, depth52 := position, tokenIndex, depth
						if buffer[position] != rune('r') {
							goto l53
						}
						position++
						goto l52
					l53:
						position, tokenIndex, depth = position52, tokenIndex52, depth52
						if buffer[position] != rune('R') {
							goto l43
						}
						position++
					}
				l52:
					{
						position54, tokenIndex54, depth54 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l55
						}
						position++
						goto l54
					l55:
						position, tokenIndex, depth = position54, tokenIndex54, depth54
						if buffer[position] != rune('T') {
							goto l43
						}
						position++
					}
				l54:
					if buffer[position] != rune('=') {
						goto l43
					}
					position++
					{
						position58, tokenIndex58, depth58 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l59
						}
						position++
						goto l58
					l59:
						position, tokenIndex, depth = position58, tokenIndex58, depth58
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l60
						}
						position++
						goto l58
					l60:
						position, tokenIndex, depth = position58, tokenIndex58, depth58
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l61
						}
						position++
						goto l58
					l61:
						position, tokenIndex, depth = position58, tokenIndex58, depth58
						if buffer[position] != rune('_') {
							goto l43
						}
						position++
					}
				l58:
				l56:
					{
						position57, tokenIndex57, depth57 := position, tokenIndex, depth
						{
							position62, tokenIndex62, depth62 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l63
							}
							position++
							goto l62
						l63:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l64
							}
							position++
							goto l62
						l64:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l65
							}
							position++
							goto l62
						l65:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if buffer[position] != rune('_') {
								goto l57
							}
							position++
						}
					l62:
						goto l56
					l57:
						position, tokenIndex, depth = position57, tokenIndex57, depth57
					}
					if buffer[position] != rune('.') {
						goto l43
					}
					position++
					{
						position68, tokenIndex68, depth68 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l69
						}
						position++
						goto l68
					l69:
						position, tokenIndex, depth = position68, tokenIndex68, depth68
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l70
						}
						position++
						goto l68
					l70:
						position, tokenIndex, depth = position68, tokenIndex68, depth68
						if buffer[position] != rune('_') {
							goto l43
						}
						position++
					}
				l68:
				l66:
					{
						position67, tokenIndex67, depth67 := position, tokenIndex, depth
						{
							position71, tokenIndex71, depth71 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l72
							}
							position++
							goto l71
						l72:
							position, tokenIndex, depth = position71, tokenIndex71, depth71
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l73
							}
							position++
							goto l71
						l73:
							position, tokenIndex, depth = position71, tokenIndex71, depth71
							if buffer[position] != rune('_') {
								goto l67
							}
							position++
						}
					l71:
						goto l66
					l67:
						position, tokenIndex, depth = position67, tokenIndex67, depth67
					}
					if buffer[position] != rune(':') {
						goto l43
					}
					position++
					{
						position76, tokenIndex76, depth76 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l77
						}
						position++
						goto l76
					l77:
						position, tokenIndex, depth = position76, tokenIndex76, depth76
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l78
						}
						position++
						goto l76
					l78:
						position, tokenIndex, depth = position76, tokenIndex76, depth76
						if buffer[position] != rune('_') {
							goto l43
						}
						position++
					}
				l76:
				l74:
					{
						position75, tokenIndex75, depth75 := position, tokenIndex, depth
						{
							position79, tokenIndex79, depth79 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l80
							}
							position++
							goto l79
						l80:
							position, tokenIndex, depth = position79, tokenIndex79, depth79
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l81
							}
							position++
							goto l79
						l81:
							position, tokenIndex, depth = position79, tokenIndex79, depth79
							if buffer[position] != rune('_') {
								goto l75
							}
							position++
						}
					l79:
						goto l74
					l75:
						position, tokenIndex, depth = position75, tokenIndex75, depth75
					}
					if !rules[Rule_]() {
						goto l43
					}
					{
						position82, tokenIndex82, depth82 := position, tokenIndex, depth
						if !rules[RuleLineTerminator]() {
							goto l82
						}
						goto l83
					l82:
						position, tokenIndex, depth = position82, tokenIndex82, depth82
					}
				l83:
					goto l7
				l43:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
					if !rules[Rule_]() {
						goto l84
					}
					{
						position85, tokenIndex85, depth85 := position, tokenIndex, depth
						if buffer[position] != rune('o') {
							goto l86
						}
						position++
						goto l85
					l86:
						position, tokenIndex, depth = position85, tokenIndex85, depth85
						if buffer[position] != rune('O') {
							goto l84
						}
						position++
					}
				l85:
					{
						position87, tokenIndex87, depth87 := position, tokenIndex, depth
						if buffer[position] != rune('u') {
							goto l88
						}
						position++
						goto l87
					l88:
						position, tokenIndex, depth = position87, tokenIndex87, depth87
						if buffer[position] != rune('U') {
							goto l84
						}
						position++
					}
				l87:
					{
						position89, tokenIndex89, depth89 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l90
						}
						position++
						goto l89
					l90:
						position, tokenIndex, depth = position89, tokenIndex89, depth89
						if buffer[position] != rune('T') {
							goto l84
						}
						position++
					}
				l89:
					{
						position91, tokenIndex91, depth91 := position, tokenIndex, depth
						if buffer[position] != rune('p') {
							goto l92
						}
						position++
						goto l91
					l92:
						position, tokenIndex, depth = position91, tokenIndex91, depth91
						if buffer[position] != rune('P') {
							goto l84
						}
						position++
					}
				l91:
					{
						position93, tokenIndex93, depth93 := position, tokenIndex, depth
						if buffer[position] != rune('o') {
							goto l94
						}
						position++
						goto l93
					l94:
						position, tokenIndex, depth = position93, tokenIndex93, depth93
						if buffer[position] != rune('O') {
							goto l84
						}
						position++
					}
				l93:
					{
						position95, tokenIndex95, depth95 := position, tokenIndex, depth
						if buffer[position] != rune('r') {
							goto l96
						}
						position++
						goto l95
					l96:
						position, tokenIndex, depth = position95, tokenIndex95, depth95
						if buffer[position] != rune('R') {
							goto l84
						}
						position++
					}
				l95:
					{
						position97, tokenIndex97, depth97 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l98
						}
						position++
						goto l97
					l98:
						position, tokenIndex, depth = position97, tokenIndex97, depth97
						if buffer[position] != rune('T') {
							goto l84
						}
						position++
					}
				l97:
					if buffer[position] != rune('=') {
						goto l84
					}
					position++
					{
						position101, tokenIndex101, depth101 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l102
						}
						position++
						goto l101
					l102:
						position, tokenIndex, depth = position101, tokenIndex101, depth101
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l103
						}
						position++
						goto l101
					l103:
						position, tokenIndex, depth = position101, tokenIndex101, depth101
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l104
						}
						position++
						goto l101
					l104:
						position, tokenIndex, depth = position101, tokenIndex101, depth101
						if buffer[position] != rune('_') {
							goto l84
						}
						position++
					}
				l101:
				l99:
					{
						position100, tokenIndex100, depth100 := position, tokenIndex, depth
						{
							position105, tokenIndex105, depth105 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l106
							}
							position++
							goto l105
						l106:
							position, tokenIndex, depth = position105, tokenIndex105, depth105
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l107
							}
							position++
							goto l105
						l107:
							position, tokenIndex, depth = position105, tokenIndex105, depth105
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l108
							}
							position++
							goto l105
						l108:
							position, tokenIndex, depth = position105, tokenIndex105, depth105
							if buffer[position] != rune('_') {
								goto l100
							}
							position++
						}
					l105:
						goto l99
					l100:
						position, tokenIndex, depth = position100, tokenIndex100, depth100
					}
					if buffer[position] != rune('.') {
						goto l84
					}
					position++
					{
						position111, tokenIndex111, depth111 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l112
						}
						position++
						goto l111
					l112:
						position, tokenIndex, depth = position111, tokenIndex111, depth111
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l113
						}
						position++
						goto l111
					l113:
						position, tokenIndex, depth = position111, tokenIndex111, depth111
						if buffer[position] != rune('_') {
							goto l84
						}
						position++
					}
				l111:
				l109:
					{
						position110, tokenIndex110, depth110 := position, tokenIndex, depth
						{
							position114, tokenIndex114, depth114 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l115
							}
							position++
							goto l114
						l115:
							position, tokenIndex, depth = position114, tokenIndex114, depth114
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l116
							}
							position++
							goto l114
						l116:
							position, tokenIndex, depth = position114, tokenIndex114, depth114
							if buffer[position] != rune('_') {
								goto l110
							}
							position++
						}
					l114:
						goto l109
					l110:
						position, tokenIndex, depth = position110, tokenIndex110, depth110
					}
					if buffer[position] != rune(':') {
						goto l84
					}
					position++
					{
						position119, tokenIndex119, depth119 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l120
						}
						position++
						goto l119
					l120:
						position, tokenIndex, depth = position119, tokenIndex119, depth119
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l121
						}
						position++
						goto l119
					l121:
						position, tokenIndex, depth = position119, tokenIndex119, depth119
						if buffer[position] != rune('_') {
							goto l84
						}
						position++
					}
				l119:
				l117:
					{
						position118, tokenIndex118, depth118 := position, tokenIndex, depth
						{
							position122, tokenIndex122, depth122 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l123
							}
							position++
							goto l122
						l123:
							position, tokenIndex, depth = position122, tokenIndex122, depth122
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l124
							}
							position++
							goto l122
						l124:
							position, tokenIndex, depth = position122, tokenIndex122, depth122
							if buffer[position] != rune('_') {
								goto l118
							}
							position++
						}
					l122:
						goto l117
					l118:
						position, tokenIndex, depth = position118, tokenIndex118, depth118
					}
					if !rules[Rule_]() {
						goto l84
					}
					{
						position125, tokenIndex125, depth125 := position, tokenIndex, depth
						if !rules[RuleLineTerminator]() {
							goto l125
						}
						goto l126
					l125:
						position, tokenIndex, depth = position125, tokenIndex125, depth125
					}
				l126:
					goto l7
				l84:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
					if !rules[Rulecomment]() {
						goto l127
					}
					{
						position128, tokenIndex128, depth128 := position, tokenIndex, depth
						{
							position130, tokenIndex130, depth130 := position, tokenIndex, depth
							if buffer[position] != rune('\n') {
								goto l131
							}
							position++
							goto l130
						l131:
							position, tokenIndex, depth = position130, tokenIndex130, depth130
							if buffer[position] != rune('\r') {
								goto l128
							}
							position++
						}
					l130:
						goto l129
					l128:
						position, tokenIndex, depth = position128, tokenIndex128, depth128
					}
				l129:
					goto l7
				l127:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
					if !rules[Rule_]() {
						goto l132
					}
					{
						position133, tokenIndex133, depth133 := position, tokenIndex, depth
						if buffer[position] != rune('\n') {
							goto l134
						}
						position++
						goto l133
					l134:
						position, tokenIndex, depth = position133, tokenIndex133, depth133
						if buffer[position] != rune('\r') {
							goto l132
						}
						position++
					}
				l133:
					goto l7
				l132:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
					if !rules[Rule_]() {
						goto l5
					}
					if !rules[Ruleconnection]() {
						goto l5
					}
					if !rules[Rule_]() {
						goto l5
					}
					{
						position135, tokenIndex135, depth135 := position, tokenIndex, depth
						if !rules[RuleLineTerminator]() {
							goto l135
						}
						goto l136
					l135:
						position, tokenIndex, depth = position135, tokenIndex135, depth135
					}
				l136:
				}
			l7:
				depth--
				add(Ruleline, position6)
			}
			return true
		l5:
			position, tokenIndex, depth = position5, tokenIndex5, depth5
			return false
		},
		/* 2 LineTerminator <- <(_ ','? comment? ('\n' / '\r')?)> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				if !rules[Rule_]() {
					goto l137
				}
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l139
					}
					position++
					goto l140
				l139:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
				}
			l140:
				{
					position141, tokenIndex141, depth141 := position, tokenIndex, depth
					if !rules[Rulecomment]() {
						goto l141
					}
					goto l142
				l141:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
				}
			l142:
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					{
						position145, tokenIndex145, depth145 := position, tokenIndex, depth
						if buffer[position] != rune('\n') {
							goto l146
						}
						position++
						goto l145
					l146:
						position, tokenIndex, depth = position145, tokenIndex145, depth145
						if buffer[position] != rune('\r') {
							goto l143
						}
						position++
					}
				l145:
					goto l144
				l143:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
				}
			l144:
				depth--
				add(RuleLineTerminator, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 3 comment <- <(_ '#' anychar*)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				if !rules[Rule_]() {
					goto l147
				}
				if buffer[position] != rune('#') {
					goto l147
				}
				position++
			l149:
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if !rules[Ruleanychar]() {
						goto l150
					}
					goto l149
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
				depth--
				add(Rulecomment, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 4 connection <- <((bridge _ ('-' '>') _ connection) / bridge)> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if !rules[Rulebridge]() {
						goto l154
					}
					if !rules[Rule_]() {
						goto l154
					}
					if buffer[position] != rune('-') {
						goto l154
					}
					position++
					if buffer[position] != rune('>') {
						goto l154
					}
					position++
					if !rules[Rule_]() {
						goto l154
					}
					if !rules[Ruleconnection]() {
						goto l154
					}
					goto l153
				l154:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					if !rules[Rulebridge]() {
						goto l151
					}
				}
			l153:
				depth--
				add(Ruleconnection, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 5 bridge <- <((port _ node _ port) / iip / rightlet / leftlet)> */
		func() bool {
			position155, tokenIndex155, depth155 := position, tokenIndex, depth
			{
				position156 := position
				depth++
				{
					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					if !rules[Ruleport]() {
						goto l158
					}
					if !rules[Rule_]() {
						goto l158
					}
					if !rules[Rulenode]() {
						goto l158
					}
					if !rules[Rule_]() {
						goto l158
					}
					if !rules[Ruleport]() {
						goto l158
					}
					goto l157
				l158:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
					if !rules[Ruleiip]() {
						goto l159
					}
					goto l157
				l159:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
					if !rules[Rulerightlet]() {
						goto l160
					}
					goto l157
				l160:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
					if !rules[Ruleleftlet]() {
						goto l155
					}
				}
			l157:
				depth--
				add(Rulebridge, position156)
			}
			return true
		l155:
			position, tokenIndex, depth = position155, tokenIndex155, depth155
			return false
		},
		/* 6 leftlet <- <(node _ port)> */
		func() bool {
			position161, tokenIndex161, depth161 := position, tokenIndex, depth
			{
				position162 := position
				depth++
				if !rules[Rulenode]() {
					goto l161
				}
				if !rules[Rule_]() {
					goto l161
				}
				if !rules[Ruleport]() {
					goto l161
				}
				depth--
				add(Ruleleftlet, position162)
			}
			return true
		l161:
			position, tokenIndex, depth = position161, tokenIndex161, depth161
			return false
		},
		/* 7 iip <- <('\'' iipchar* '\'')> */
		func() bool {
			position163, tokenIndex163, depth163 := position, tokenIndex, depth
			{
				position164 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l163
				}
				position++
			l165:
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if !rules[Ruleiipchar]() {
						goto l166
					}
					goto l165
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
				if buffer[position] != rune('\'') {
					goto l163
				}
				position++
				depth--
				add(Ruleiip, position164)
			}
			return true
		l163:
			position, tokenIndex, depth = position163, tokenIndex163, depth163
			return false
		},
		/* 8 rightlet <- <(port _ node)> */
		func() bool {
			position167, tokenIndex167, depth167 := position, tokenIndex, depth
			{
				position168 := position
				depth++
				if !rules[Ruleport]() {
					goto l167
				}
				if !rules[Rule_]() {
					goto l167
				}
				if !rules[Rulenode]() {
					goto l167
				}
				depth--
				add(Rulerightlet, position168)
			}
			return true
		l167:
			position, tokenIndex, depth = position167, tokenIndex167, depth167
			return false
		},
		/* 9 node <- <(([a-z] / [A-Z] / [0-9] / '_')+ component?)> */
		func() bool {
			position169, tokenIndex169, depth169 := position, tokenIndex, depth
			{
				position170 := position
				depth++
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l174
					}
					position++
					goto l173
				l174:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l175
					}
					position++
					goto l173
				l175:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l176
					}
					position++
					goto l173
				l176:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('_') {
						goto l169
					}
					position++
				}
			l173:
			l171:
				{
					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					{
						position177, tokenIndex177, depth177 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l178
						}
						position++
						goto l177
					l178:
						position, tokenIndex, depth = position177, tokenIndex177, depth177
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l179
						}
						position++
						goto l177
					l179:
						position, tokenIndex, depth = position177, tokenIndex177, depth177
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l180
						}
						position++
						goto l177
					l180:
						position, tokenIndex, depth = position177, tokenIndex177, depth177
						if buffer[position] != rune('_') {
							goto l172
						}
						position++
					}
				l177:
					goto l171
				l172:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
				}
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					if !rules[Rulecomponent]() {
						goto l181
					}
					goto l182
				l181:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
				}
			l182:
				depth--
				add(Rulenode, position170)
			}
			return true
		l169:
			position, tokenIndex, depth = position169, tokenIndex169, depth169
			return false
		},
		/* 10 component <- <('(' ([a-z] / [A-Z] / '/' / '-' / [0-9] / '_')* compMeta? ')')> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				if buffer[position] != rune('(') {
					goto l183
				}
				position++
			l185:
				{
					position186, tokenIndex186, depth186 := position, tokenIndex, depth
					{
						position187, tokenIndex187, depth187 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l188
						}
						position++
						goto l187
					l188:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l189
						}
						position++
						goto l187
					l189:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
						if buffer[position] != rune('/') {
							goto l190
						}
						position++
						goto l187
					l190:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
						if buffer[position] != rune('-') {
							goto l191
						}
						position++
						goto l187
					l191:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l192
						}
						position++
						goto l187
					l192:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
						if buffer[position] != rune('_') {
							goto l186
						}
						position++
					}
				l187:
					goto l185
				l186:
					position, tokenIndex, depth = position186, tokenIndex186, depth186
				}
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if !rules[RulecompMeta]() {
						goto l193
					}
					goto l194
				l193:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
				}
			l194:
				if buffer[position] != rune(')') {
					goto l183
				}
				position++
				depth--
				add(Rulecomponent, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 11 compMeta <- <(':' ([a-z] / [A-Z] / '/')+)> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				if buffer[position] != rune(':') {
					goto l195
				}
				position++
				{
					position199, tokenIndex199, depth199 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l200
					}
					position++
					goto l199
				l200:
					position, tokenIndex, depth = position199, tokenIndex199, depth199
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l201
					}
					position++
					goto l199
				l201:
					position, tokenIndex, depth = position199, tokenIndex199, depth199
					if buffer[position] != rune('/') {
						goto l195
					}
					position++
				}
			l199:
			l197:
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					{
						position202, tokenIndex202, depth202 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l203
						}
						position++
						goto l202
					l203:
						position, tokenIndex, depth = position202, tokenIndex202, depth202
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l204
						}
						position++
						goto l202
					l204:
						position, tokenIndex, depth = position202, tokenIndex202, depth202
						if buffer[position] != rune('/') {
							goto l198
						}
						position++
					}
				l202:
					goto l197
				l198:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
				}
				depth--
				add(RulecompMeta, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 12 port <- <(([A-Z] / '.' / [0-9] / '_')+ __)> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				{
					position209, tokenIndex209, depth209 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l210
					}
					position++
					goto l209
				l210:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
					if buffer[position] != rune('.') {
						goto l211
					}
					position++
					goto l209
				l211:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l212
					}
					position++
					goto l209
				l212:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
					if buffer[position] != rune('_') {
						goto l205
					}
					position++
				}
			l209:
			l207:
				{
					position208, tokenIndex208, depth208 := position, tokenIndex, depth
					{
						position213, tokenIndex213, depth213 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l214
						}
						position++
						goto l213
					l214:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune('.') {
							goto l215
						}
						position++
						goto l213
					l215:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l216
						}
						position++
						goto l213
					l216:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune('_') {
							goto l208
						}
						position++
					}
				l213:
					goto l207
				l208:
					position, tokenIndex, depth = position208, tokenIndex208, depth208
				}
				if !rules[Rule__]() {
					goto l205
				}
				depth--
				add(Ruleport, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 13 anychar <- <(!('\n' / '\r') .)> */
		func() bool {
			position217, tokenIndex217, depth217 := position, tokenIndex, depth
			{
				position218 := position
				depth++
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					{
						position220, tokenIndex220, depth220 := position, tokenIndex, depth
						if buffer[position] != rune('\n') {
							goto l221
						}
						position++
						goto l220
					l221:
						position, tokenIndex, depth = position220, tokenIndex220, depth220
						if buffer[position] != rune('\r') {
							goto l219
						}
						position++
					}
				l220:
					goto l217
				l219:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
				}
				if !matchDot() {
					goto l217
				}
				depth--
				add(Ruleanychar, position218)
			}
			return true
		l217:
			position, tokenIndex, depth = position217, tokenIndex217, depth217
			return false
		},
		/* 14 iipchar <- <(('\\' '\'') / (!'\'' .))> */
		func() bool {
			position222, tokenIndex222, depth222 := position, tokenIndex, depth
			{
				position223 := position
				depth++
				{
					position224, tokenIndex224, depth224 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l225
					}
					position++
					if buffer[position] != rune('\'') {
						goto l225
					}
					position++
					goto l224
				l225:
					position, tokenIndex, depth = position224, tokenIndex224, depth224
					{
						position226, tokenIndex226, depth226 := position, tokenIndex, depth
						if buffer[position] != rune('\'') {
							goto l226
						}
						position++
						goto l222
					l226:
						position, tokenIndex, depth = position226, tokenIndex226, depth226
					}
					if !matchDot() {
						goto l222
					}
				}
			l224:
				depth--
				add(Ruleiipchar, position223)
			}
			return true
		l222:
			position, tokenIndex, depth = position222, tokenIndex222, depth222
			return false
		},
		/* 15 _ <- <(' ' / '\t')*> */
		func() bool {
			{
				position228 := position
				depth++
			l229:
				{
					position230, tokenIndex230, depth230 := position, tokenIndex, depth
					{
						position231, tokenIndex231, depth231 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l232
						}
						position++
						goto l231
					l232:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
						if buffer[position] != rune('\t') {
							goto l230
						}
						position++
					}
				l231:
					goto l229
				l230:
					position, tokenIndex, depth = position230, tokenIndex230, depth230
				}
				depth--
				add(Rule_, position228)
			}
			return true
		},
		/* 16 __ <- <(' ' / '\t')+> */
		func() bool {
			position233, tokenIndex233, depth233 := position, tokenIndex, depth
			{
				position234 := position
				depth++
				{
					position237, tokenIndex237, depth237 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l238
					}
					position++
					goto l237
				l238:
					position, tokenIndex, depth = position237, tokenIndex237, depth237
					if buffer[position] != rune('\t') {
						goto l233
					}
					position++
				}
			l237:
			l235:
				{
					position236, tokenIndex236, depth236 := position, tokenIndex, depth
					{
						position239, tokenIndex239, depth239 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l240
						}
						position++
						goto l239
					l240:
						position, tokenIndex, depth = position239, tokenIndex239, depth239
						if buffer[position] != rune('\t') {
							goto l236
						}
						position++
					}
				l239:
					goto l235
				l236:
					position, tokenIndex, depth = position236, tokenIndex236, depth236
				}
				depth--
				add(Rule__, position234)
			}
			return true
		l233:
			position, tokenIndex, depth = position233, tokenIndex233, depth233
			return false
		},
	}
	p.rules = rules
}
