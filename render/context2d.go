package render

var (
	TextAlignCenter = "center"
	TextAlignLeft   = "left"
	TextAlignRight  = "right"
)

type Style struct {
	LineWidth   float64
	StrokeStyle string
	FillStyle   string
	Stroke      bool
	Fill        bool
}

type Point struct {
	X, Y float64
}

type Line struct {
	*Style
	Start, End *Point
}

type Chain struct {
	*Style
	points []*Point
	Length int
}

func NewChain(pts ...*Point) *Chain {
	return &Chain{points: pts, Length: len(pts)}
}

func (c *Chain) At(i int) *Point {
	return c.points[i]
}

type Circle struct {
	*Style
	X, Y, D float64
}

type Rect struct {
	*Style
	X, Y, W, H float64
}

type RoundedRect struct {
	*Style
	X, Y, W, H, R float64
}

type Poly struct {
	*Style
	points []*Point
	Length int
}

func NewPoly(pts ...*Point) *Chain {
	return &Chain{points: pts, Length: len(pts)}
}

func (c *Poly) At(i int) *Point {
	return c.points[i]
}

type Text struct {
	*Style
	X, Y     float64
	FontSize float64
	Oblique  bool
	Vertical bool
	Align    string
	Font     string
	Label    string
}

type BezierCurve struct {
	*Style
	Start, End, C1, C2 *Point
}
