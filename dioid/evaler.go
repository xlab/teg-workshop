package dioid

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xlab/teg-workshop/util"
)

var (
	regexSpace = regexp.MustCompile(`\s+`)
	regexGd    = regexp.MustCompile(`(g\^-?\d+d\^-?\d+|(?:g|d)\^-?\d+|gd|e)`)
	operators  = "+x*"
)

func Eval(expr string) (result Serie, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Invalid expression %v (%v)", expr, e)
		}
	}()

	tokens := tokenise(expr)
	postfix := convert2postfix(tokens)
	if serie, err := evaluatePostfix(postfix); err != nil {
		return Serie{}, errors.New("Invalid expression")
	} else {
		return serie, nil
	}
}

func prec(op string) (result int) {
	if op == "+" {
		result = 1
	} else if op == "x" {
		result = 2
	} else if op == "*" {
		result = 3
	}
	return
}

func opGTE(op1, op2 string) bool {
	return prec(op1) >= prec(op2)
}

func isOperator(token string) bool {
	return strings.Contains(operators, token)
}

func isOperand(token string) bool {
	return regexGd.MatchString(token)
}

func convert2postfix(tokens []string) []string {
	var stack util.Stack
	var result []string
	for _, token := range tokens {
		if isOperator(token) {
		OPERATOR:
			for {
				top, err := stack.Top()
				if err != nil || top == "(" {
					break OPERATOR
				}
				if opGTE(top.(string), token) {
					pop, _ := stack.Pop()
					result = append(result, pop.(string))
				} else {
					break OPERATOR
				}
				break OPERATOR
			}
			stack.Push(token)

		} else if token == "(" {
			stack.Push(token)
		} else if token == ")" {
		CLOSE_PAREN:
			for {
				top, err := stack.Top()
				if err != nil || top == "(" {
					stack.Pop() // pop off "("
					break CLOSE_PAREN
				} else {
					pop, _ := stack.Pop()
					result = append(result, pop.(string))
				}
			}
		} else if isOperand(token) {
			result = append(result, token)
		}
	}

	for !stack.IsEmpty() {
		pop, _ := stack.Pop()
		result = append(result, pop.(string))
	}

	return result
}

func evaluatePostfix(postfix []string) (Serie, error) {
	// log.Println("postifx", postfix)
	var stack util.Stack
	for _, token := range postfix {
		if isOperand(token) {
			gd, err := scanGd(token)
			if err != nil {
				return Serie{}, err
			}
			stack.Push(Serie{Q: Poly{gd}})
		} else if isOperator(token) {
			pop2 := func() (s1, s2 Serie, err error) {
				op2, err := stack.Pop()
				if err != nil {
					return Serie{}, Serie{}, err
				}
				op1, err := stack.Pop()
				if err != nil {
					return Serie{}, Serie{}, err
				}
				return op1.(Serie), op2.(Serie), nil
			}
			pop1 := func() (s Serie, err error) {
				op, err := stack.Pop()
				if err != nil {
					return Serie{}, err
				}
				return op.(Serie), nil
			}

			switch token {
			case "*":
				s, err := pop1()
				if err != nil {
					return Serie{}, err
				}
				// log.Printf("Starring %#v", s)
				stack.Push(SerieStar(s))
			case "x":
				s1, s2, err := pop2()
				if err != nil {
					return Serie{}, err
				}
				// log.Printf("OTIMES series %#v AND %#v", s2, s1)
				// not commutative! reverse order (postfix -> infix)
				stack.Push(SerieOtimes(s2, s1))
			case "+":
				s1, s2, err := pop2()
				if err != nil {
					return Serie{}, err
				}
				// log.Printf("OPLUS series %#v AND %#v", s1, s2)
				stack.Push(SerieOplus(s1, s2))
			default:
				return Serie{}, fmt.Errorf("unknown operator %v", token)
			}
		} else {
			return Serie{}, fmt.Errorf("unknown token %v", token)
		}
	}

	// log.Println("Dumping stack\n", stack.Dump())
	tmp, err := stack.Pop()
	if err != nil {
		return Serie{}, err
	}
	if result, ok := tmp.(Serie); !ok {
		return Serie{}, fmt.Errorf("result is not a valid serie")
	} else {
		return result, nil
	}
}

func tokenise(expr string) []string {
	spaced := regexGd.ReplaceAllString(expr, " ${1} ")
	symbols := []string{"(", ")"}
	for _, symbol := range symbols {
		spaced = strings.Replace(spaced, symbol, fmt.Sprintf(" %s ", symbol), -1)
	}
	stripped := regexSpace.ReplaceAllString(strings.TrimSpace(spaced), "|")
	return strings.Split(stripped, "|")
}
