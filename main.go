package main

import (
	"GoHilbert/propositional"
)

var ProofPatterns [][]*propositional.Formula;

func _Merge(order int, A *propositional.Formula, A_then_C *propositional.Formula) {
	
}

func Merge(order int, X *propositional.Formula, Y *propositional.Formula) {
	_Merge(order, X, Y)
	_Merge(order, Y, X)
}

func main()  {

	ProofPatterns = make([][]*propositional.Formula, 10)

	Axiom1, _ := propositional.NewFormula("((A)>((B)>(A)))")
	Axiom2, _ := propositional.NewFormula("(((A)>((B)>(C)))>(((A)>(B))>((A)>(C))))")

	ProofPatterns[0] = append(ProofPatterns[0], 
		Axiom1,
		Axiom2,
	)

	for order := 1; order < 3; order++ {
		for idx, P1 := range ProofPatterns[order - 1] {
			
			for i := 0; i < order - 1; i++ {
				for _, P2 := range ProofPatterns[i] {
					Merge(order, P1, P2)
				}
			}

			for i := 0; i <= idx; i++ {
				Merge(order, P1, ProofPatterns[order - 1][i])
			}
		}
	}
}