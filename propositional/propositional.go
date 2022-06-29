package propositional

import (
	"errors"
	"fmt"
	"regexp"
)

type Atom struct {
	Identifier string
	parent *Formula
}

func (p Atom) IsComposite() bool {
	return false
}

func (p Atom) String() string {
	return fmt.Sprintf("%v", p.Identifier)
}

func (p Atom) MakeFormula() *Formula {
	if p.parent != nil {
		return p.parent
	}
	// fmt.Println("Generating parent for ", p.Identifier)
	p.parent = &Formula{
		LeftSide: nil,
		RightSide: nil,
		Literal: &p,
	}
	return p.parent
}





type Formula struct {
	LeftSide *Formula
	RightSide *Formula
	Literal *Atom
}

func (A Formula) IsComposite() bool {
	return A.Literal == nil
}

// todo : could be more optimized
func (A Formula) String() string {
	if A.IsComposite() {
		return "(" + A.LeftSide.String() + ">" + A.RightSide.String() + ")"
	} else {
		return "(" + A.Literal.String() + ")"
	}
}

func (F Formula) MakeFormula() *Formula {
	// A := F
	return &F
}





type Evaluatable interface {
	MakeFormula() *Formula
	IsComposite() bool
}

func Then(X Evaluatable, Y Evaluatable) *Formula {
	return &Formula {
		Literal: nil,
		LeftSide: X.MakeFormula(),
		RightSide: Y.MakeFormula(),
	}
}

func SyntacticalEqual(X *Formula, Y *Formula) bool {
	if X == Y {
		return true;
	}

	// (saman): why this may happen??
	if X == nil || Y == nil {
		return false;
	}

	if !X.IsComposite(){
		if Y.IsComposite() || X.Literal.Identifier != Y.Literal.Identifier {
			return false
		}
		X.Literal = Y.Literal
		return true
	}
	
	if (X.LeftSide != Y.LeftSide) && !SyntacticalEqual(X.LeftSide, Y.LeftSide) {
		return false
	}

	if (X.RightSide != Y.RightSide) && !SyntacticalEqual(X.RightSide, Y.RightSide) {
		return false
	}

	X.RightSide = Y.RightSide
	X.LeftSide = Y.LeftSide
	return true
}

func x_LeafMapTo (key string, value Evaluatable) map[string]*Formula {
	result := make(map[string]*Formula)
	result[key] = value.MakeFormula()
	// fmt.Println("Returning for ", key, " -> ", value)
	return result
}

func Destruct(X Evaluatable, Pattern Evaluatable) (map[string]*Formula, error) {
	// fmt.Println("pattern:", Pattern,"\nX:", X)
	switch Pattern.(type) {
	case Atom, Formula:
		return Destruct(X, Pattern.MakeFormula())
	case *Atom:
		return x_LeafMapTo(Pattern.(*Atom).Identifier, X), nil
	case *Formula:
		switch X.(type) {
		case Atom, Formula:
			return Destruct(X.MakeFormula(), Pattern)
		case *Atom:
			//TakeCare
			return nil, errors.New("Can't match Atom to Formula")
		case *Formula:
			if Pattern.IsComposite() {
				if !X.IsComposite() {
					return nil, errors.New("Can't match non-composite to composite pattern")
				}

				mapLeft, err := Destruct(X.(*Formula).LeftSide, Pattern.(*Formula).LeftSide)
				if err != nil {
					return nil, errors.New("L:" + err.Error())
				}
				
				mapRight, err := Destruct(X.(*Formula).RightSide, Pattern.(*Formula).RightSide)
				if err != nil {
					return nil, errors.New("R:" + err.Error())
				}

				// merge into left
				for key, valueRight := range mapRight {
					if valueLeft, prs := mapLeft[key]; prs && !SyntacticalEqual(valueLeft, valueRight) {
						return nil, errors.New("Many Expression for PlaceHolder (" + key + ")")
					} else {
						mapLeft[key] = valueRight
					}
				}

				return mapLeft, nil
			} else {
				return x_LeafMapTo(Pattern.(*Formula).Literal.Identifier, X), nil
			}
		}
	}
	return nil, errors.New("No Matching Case")
}

func DestructWithString(X Evaluatable, Pattern string) (map[string]*Formula, error) {
	F, _ := NewFormula(Pattern)
	return Destruct(X, *F)
}

func NewFormula(format string) (*Formula, error) {
	// fmt.Println("Asking New Formula for", format)

	if matched, _ := regexp.MatchString(`^\([A-Za-z0-9_>\(\)]*\)$`, format); !matched {
		return nil, errors.New("Not included in Parentheses OR Invalid Identifier")
	}

	if matched, _ := regexp.MatchString(`\>`, format); !matched {
		return Atom{Identifier: format[1:len(format) - 1]}.MakeFormula(), nil
	}
	
	// fmt.Println("Splitting Format")

	cnt := 0
	ptr := -1
	for i := 0; i < len(format); i++ {
		switch format[i] {
		case '(':
			cnt++;
		case ')':
			cnt--;
		case '>':
			if cnt == 1 {
				if ptr == -1 {
					ptr = i
				} else {
					return nil, errors.New("More than one splitter")
				}
			}
		}
	}

	if ptr == -1 {
		return nil, errors.New("Failed to Find Splitter(>)")
	}
	if cnt != 0 {
		return nil, errors.New("Invalid Parentheses")
	}

	leftSide, err := NewFormula(format[1:ptr])
	if err != nil {
		return nil, err
	}

	rightSide, err := NewFormula(format[ptr+1:len(format)-1])
	if err != nil {
		return nil, err
	}

	return Then(leftSide, rightSide), nil
}

func ChangeIdentifiers(X *Formula, IdentifiersMap *map[string]string) {
	if !X.IsComposite() {
		X.Literal.Identifier = (*IdentifiersMap)[X.Literal.Identifier]
	} else {
		ChangeIdentifiers(X.LeftSide, IdentifiersMap)
		ChangeIdentifiers(X.RightSide, IdentifiersMap)
	}
}

func ReplaceAtoms(X *Formula, ReplacementMap *map[string]*Formula) {
	if !X.IsComposite() {
		if replacement, present := (*ReplacementMap)[X.Literal.Identifier]; present {
			X.LeftSide, X.RightSide, X.Literal = replacement.LeftSide, replacement.RightSide, replacement.Literal
		}
	} else {
		ReplaceAtoms(X.LeftSide, ReplacementMap)
		ReplaceAtoms(X.RightSide, ReplacementMap)
	}
}

func NewIdentifiers(X *Formula, NewPrefix string) {
	AtomsIdentifiers, _ := Destruct(X, X)

	GeneratedIdentifiers := make(map[string]string)
	firstFreeIndex := 0

	for Identifier, _ := range AtomsIdentifiers {
		if _, present := GeneratedIdentifiers[Identifier]; !present {
			GeneratedIdentifiers[Identifier] = fmt.Sprintf("%v%v", NewPrefix, firstFreeIndex)
			firstFreeIndex++;
		}
	}
	ChangeIdentifiers(X, &GeneratedIdentifiers);
}

func CopyFormula (X *Formula) (*Formula, error) {
	return NewFormula(fmt.Sprintf("%v", X))
}