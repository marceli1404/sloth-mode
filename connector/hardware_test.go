package main

import "testing"

func TestQuantRecommendation(t *testing.T) {
	cases := []struct {
		ram  float64
		want string
	}{{4, "Q3_K_L"}, {8, "Q4_K_M"}, {16, "Q5_K_M"}, {32, "Q6_K"}}
	for _, c := range cases {
		q, _ := recommendQuant(HardwareProfile{RAMGiB: c.ram})
		if q != c.want {
			t.Fatalf("%v GB got %s want %s", c.ram, q, c.want)
		}
	}
}
func TestQuantPinned(t *testing.T) {
	for _, q := range []string{"Q3_K_L", "Q4_K_M", "Q5_K_M", "Q6_K"} {
		s := modelSpec(q)
		if len(s.Hash) != 64 || s.Name == "" || s.URL == "" {
			t.Fatal(q)
		}
	}
}
func TestPreferenceReject(t *testing.T) {
	if setPreferredQuant(t.TempDir(), "../../malicious") == nil {
		t.Fatal("invalid quant accepted")
	}
}
