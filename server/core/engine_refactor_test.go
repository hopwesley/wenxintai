package core

import (
	"math"
	"testing"
)

func TestRunMatchesBuildFullParam(t *testing.T) {
	combo := ComboPHY_CHE_BIO
	riasec := AllRIASECCombos[combo]
	ascAligned := AllASCCombos[combo]["aligned"]

	paramOld, resultOld, scoresOld := BuildFullParam(riasec, ascAligned, 0, 0, 0)

	out33, err := Run(Input{
		RIASECAnswers: riasec,
		ASCAnswers:    ascAligned,
	}, Mode33)
	if err != nil {
		t.Fatalf("Run Mode33 error: %v", err)
	}

	compareScores(t, scoresOld, out33.Scores)
	compareCommon(t, resultOld.Common, out33.Result.Common)
	if out33.Param.Mode33 == nil || paramOld.Mode33 == nil {
		t.Fatalf("missing mode33 data")
	}
	compareMode33(t, paramOld.Mode33, out33.Param.Mode33)

	out312, err := Run(Input{
		RIASECAnswers: riasec,
		ASCAnswers:    ascAligned,
	}, Mode312)
	if err != nil {
		t.Fatalf("Run Mode312 error: %v", err)
	}
	if out312.Param.Mode312 == nil || paramOld.Mode312 == nil {
		t.Fatalf("missing mode312 data")
	}
	compareMode312(t, paramOld.Mode312, out312.Param.Mode312)
}

func compareScores(t *testing.T, old, new []SubjectScores) {
	if len(old) != len(new) {
		t.Fatalf("scores length mismatch: %d vs %d", len(old), len(new))
	}
	for i := range old {
		if old[i].Subject != new[i].Subject {
			t.Fatalf("subject mismatch at %d: %s vs %s", i, old[i].Subject, new[i].Subject)
		}
		if !almostEqual(old[i].Fit, new[i].Fit) || !almostEqual(old[i].A, new[i].A) || !almostEqual(old[i].I, new[i].I) || !almostEqual(old[i].AZ, new[i].AZ) || !almostEqual(old[i].IZ, new[i].IZ) {
			t.Fatalf("score mismatch for %s", old[i].Subject)
		}
	}
}

func compareCommon(t *testing.T, old, new *CommonSection) {
	if old == nil || new == nil {
		t.Fatalf("common section missing")
	}
	if !almostEqual(old.GlobalCosine, new.GlobalCosine) || !almostEqual(old.QualityScore, new.QualityScore) {
		t.Fatalf("common metrics mismatch")
	}
	if len(old.Subjects) != len(new.Subjects) {
		t.Fatalf("subjects length mismatch")
	}
	for i := range old.Subjects {
		o, n := old.Subjects[i], new.Subjects[i]
		if o.Subject != n.Subject {
			t.Fatalf("common subject mismatch at %d", i)
		}
		if !almostEqual(o.InterestZ, n.InterestZ) || !almostEqual(o.AbilityZ, n.AbilityZ) || !almostEqual(o.ZGap, n.ZGap) || !almostEqual(o.AbilityShare, n.AbilityShare) || !almostEqual(o.Fit, n.Fit) {
			t.Fatalf("common subject values mismatch for %s", o.Subject)
		}
	}
}

func compareMode33(t *testing.T, old, new *Mode33Section) {
	if len(old.TopCombinations) != len(new.TopCombinations) {
		t.Fatalf("mode33 combos length mismatch")
	}
	for i := range old.TopCombinations {
		o, n := old.TopCombinations[i], new.TopCombinations[i]
		if o.Subjects != n.Subjects {
			t.Fatalf("combo subjects mismatch at %d", i)
		}
		if !almostEqual(o.AvgFit, n.AvgFit) || !almostEqual(o.MinAbility, n.MinAbility) || !almostEqual(o.Rarity, n.Rarity) || !almostEqual(o.RiskPenalty, n.RiskPenalty) || !almostEqual(o.Score, n.Score) || !almostEqual(o.ComboCosine, n.ComboCosine) {
			t.Fatalf("combo values mismatch at %d", i)
		}
	}
}

func compareMode312(t *testing.T, old, new *Mode312Section) {
	compareAnchor(t, old.AnchorPHY, new.AnchorPHY)
	compareAnchor(t, old.AnchorHIS, new.AnchorHIS)
}

func compareAnchor(t *testing.T, old, new AnchorCoreData) {
	if old.Subject != new.Subject {
		t.Fatalf("anchor subject mismatch: %s vs %s", old.Subject, new.Subject)
	}
	if !almostEqual(old.Fit, new.Fit) || !almostEqual(old.AbilityNorm, new.AbilityNorm) || !almostEqual(old.TermFit, new.TermFit) || !almostEqual(old.TermAbility, new.TermAbility) || !almostEqual(old.TermCoverage, new.TermCoverage) || !almostEqual(old.S1, new.S1) || !almostEqual(old.SFinal, new.SFinal) {
		t.Fatalf("anchor metrics mismatch for %s", old.Subject)
	}
	if len(old.Combos) != len(new.Combos) {
		t.Fatalf("anchor combos length mismatch for %s", old.Subject)
	}
	for i := range old.Combos {
		o, n := old.Combos[i], new.Combos[i]
		if o.Aux1 != n.Aux1 || o.Aux2 != n.Aux2 {
			t.Fatalf("anchor combo subjects mismatch")
		}
		if !almostEqual(o.AvgFit, n.AvgFit) || !almostEqual(o.MinFit, n.MinFit) || !almostEqual(o.ComboCos, n.ComboCos) || !almostEqual(o.AuxAbility, n.AuxAbility) || !almostEqual(o.Coverage, n.Coverage) || !almostEqual(o.MixPenalty, n.MixPenalty) || !almostEqual(o.S23, n.S23) || !almostEqual(o.SFinalCombo, n.SFinalCombo) {
			t.Fatalf("anchor combo mismatch for %s", old.Subject)
		}
	}
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-6
}
