package dioid

func (t *testSuite) TestEval() {
	a := Serie{
		P: Poly{{-1, -3}, {0, 0}},
		Q: Poly{{2, 2}, {3, 3}},
		R: Gd{2, 3},
	}

	b, err := Eval("e+(g^2d^2+g^-1d^-3)x(gd+g^2d^3)*")
	if err != nil {
		t.Error(err)
	}
	t.Equal(a.String(), b.String())
}

func (t *testSuite) TestEvalSimple() {
	a := Serie{
		P: Poly{{-1, 0}, {0, 2}},
		Q: Poly{{2, 3}},
		R: Gd{0, 0},
	}
	b, err := Eval("g^-1d^-0 + e + g^2 + d^2 + g^2d^0 + g^2d^2 + g^2d^3 + g^2d^2 + g^3d^3")
	if err != nil {
		t.Error(err)
	}
	t.Equal(a.String(), b.String())
}

func (t *testSuite) TestEvalMix() {
	a := Serie{
		P: Poly{{-1, 0}, {2, 2}},
		Q: Poly{{9, 10}},
		R: Gd{0, 0},
	}
	b, err := Eval("g^-1d^-0 + e x g^2 + d^2 x g^2d^0 + g^2d^2 x g^2d^3 x g^2d^2 x g^3d^3")
	if err != nil {
		t.Error(err)
	}
	t.Equal(a.String(), b.String())
}
