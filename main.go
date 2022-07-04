// TOGO from here:
//  1. time spending analysis in _Merge
//  2. GoRoutins
//  3. Input Related Behavior
//  4. Rasterizations(DFS)

package main

import (
	"GoHilbert/propositional"
	"fmt"
	"time"

	"github.com/dghubble/trie"
	"github.com/sbwhitecap/tqdm"
)

var ProofPatterns [][]*propositional.Formula
var PatternsTrie *trie.PathTrie

var TimeOfBasicArrowCheck int64;
var TimeOfArrowDestruction int64;
var TimeOfHardCopy int64;
var TimeOfAtomRespacing int64;
var TimeOfDecomposition int64;
var TimeOfFinalPush int64;
var inMergeTime int64;


func _Merge(order int, A *propositional.Formula, A_then_C *propositional.Formula) {
	// fmt.Println("Checking:", A, " and ", A_then_C)
	overallCheckpoint := time.Now()
	checkpoint := time.Now()


	decomposition, err := propositional.DestructWithString(A_then_C, "((Ant)>(Con))");
	if err != nil {
		return;
	}
	antecedent := decomposition["Ant"]
	consequent := decomposition["Con"]

	TimeOfBasicArrowCheck += time.Since(checkpoint).Microseconds()
	checkpoint = time.Now()

	if _, err := propositional.Destruct(A, antecedent); err != nil {
		return
	}

	TimeOfArrowDestruction += time.Since(checkpoint).Microseconds()
	checkpoint = time.Now()

	// make sure to make a new formula tree
	A_then_C, _ = propositional.CopyFormula(A_then_C)
	propositional.NewIdentifiers(A_then_C, "Q_TMP_")
	
	TimeOfHardCopy += time.Since(checkpoint).Microseconds()
	checkpoint = time.Now()

	decomposition, _ = propositional.DestructWithString(A_then_C, "((Ant)>(Con))")
	antecedent = decomposition["Ant"]
	consequent = decomposition["Con"]
	
	replacementMap, err := propositional.Destruct(A, antecedent);
	if err != nil {
		return
	}
	propositional.ReplaceAtoms(consequent, &replacementMap)

	TimeOfAtomRespacing += time.Since(checkpoint).Microseconds()
	checkpoint = time.Now()

	consequent, err = propositional.CopyFormula(consequent)
	propositional.NewIdentifiers(consequent, "P")
	if err != nil {
		return
	}

	TimeOfDecomposition += time.Since(checkpoint).Microseconds()
	checkpoint = time.Now()

	if PatternsTrie.Get(consequent.String()) != nil {
		return
	}

	PatternsTrie.Put(consequent.String(), true)
	// fmt.Printf("New Pattern at order=%v  -->  %v\n", order, consequent)
	ProofPatterns[order] = append(ProofPatterns[order], consequent)

	TimeOfFinalPush += time.Since(checkpoint).Microseconds()
	checkpoint = time.Now()

	inMergeTime += time.Since(overallCheckpoint).Microseconds()
}

func Merge(order int, X *propositional.Formula, Y *propositional.Formula) {
	_Merge(order, X, Y)
	_Merge(order, Y, X)
}

func main()  {
	PatternsTrie = trie.NewPathTrie()
	ProofPatterns = make([][]*propositional.Formula, 10)

	Axiom1, _ := propositional.NewFormula("((A)>((B)>(A)))")
	Axiom2, _ := propositional.NewFormula("(((A)>((B)>(C)))>(((A)>(B))>((A)>(C))))")

	ProofPatterns[0] = append(ProofPatterns[0], 
		Axiom1,
		Axiom2,
	)

	PatternsTrie.Put(Axiom1.String(), true)
	PatternsTrie.Put(Axiom2.String(), true)

	for order := 1; order < 6; order++ {
		start := time.Now()
		iterations := 0

		// below line is similar to --> for idx, P1 := range ProofPatterns[order - 1] {
		tqdm.R(0, len(ProofPatterns[order - 1]), func(idx interface{}) (brk bool) {
			P1 := ProofPatterns[order - 1][idx.(int)]
			
			for i := 0; i < order - 1; i++ {
				for _, P2 := range ProofPatterns[i] {
					iterations++
					Merge(order, P1, P2)
				}
			}

			for i := 0; i <= idx.(int); i++ {
				iterations++
				Merge(order, P1, ProofPatterns[order - 1][i])
			}
			return false;
		});

		elapsed := time.Since(start).Milliseconds()
		fmt.Println("-----------------------------------------------------------------------------------------------")
		fmt.Printf("Finalizing order %v at %v members took %v for %v iterations\n", order, len(ProofPatterns[order]), elapsed, iterations)
		fmt.Printf("--Time Analysis--\nTimeOfBasicArrowCheck: %v\nTimeOfArrowDestruction: %v\nTimeOfHardCopy: %v\nTimeOfAtomRespacing: %v\nTimeOfDecomposition: %v\nTimeOfFinalPush: %v\n", TimeOfBasicArrowCheck, TimeOfArrowDestruction, TimeOfHardCopy, TimeOfAtomRespacing, TimeOfDecomposition, TimeOfFinalPush)
		fmt.Printf("OVERALL IN FUNCTION TIME: %v\n", inMergeTime)
		fmt.Println("===============================================================================================")
		fmt.Println("")
	}
}