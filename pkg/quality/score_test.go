package quality

import "testing"

func TestVerdict_String(t *testing.T) {
	cases := map[Verdict]string{
		Pass:            "pass",
		FailLowRes:      "fail_lowres",
		FailBlurry:      "fail_blurry",
		FailSmallBytes:  "fail_small_bytes",
		FailCorrupt:     "fail_corrupt",
		FailWrongAspect: "fail_wrong_aspect",
		FailTooLarge:    "fail_too_large",
		FailEmpty:       "fail_empty",
	}
	for v, want := range cases {
		if got := v.String(); got != want {
			t.Errorf("Verdict(%d).String() = %q, want %q", int(v), got, want)
		}
	}
	if s := Verdict(99).String(); s != "verdict_99" {
		t.Errorf("unknown verdict default = %q", s)
	}
}

func TestScore_OK(t *testing.T) {
	if !(Score{Verdict: Pass}).OK() {
		t.Fatal("Pass score should report OK")
	}
	if (Score{Verdict: FailBlurry}).OK() {
		t.Fatal("FailBlurry score should not report OK")
	}
}
