package itype

// Recti is a 2D rectangle with int coordinates.
type Recti struct {
	Left, Top     int
	Width, Height int
}

func (r Recti) MinPoint() Vec2i {
	return Vec2i{r.Left, r.Top}
}

func (r Recti) MaxPoint() Vec2i {
	return Vec2i{r.Left + r.Width, r.Top + r.Height}
}

func (r Recti) Size() Vec2i {
	return Vec2i{r.Width, r.Height}
}

// Rectf is a 2D rectangle with float32 coordinates.
type Rectf struct {
	Left, Top     float32
	Width, Height float32
}

func (r Rectf) MinPoint() Vec2f {
	return Vec2f{r.Left, r.Top}
}

func (r Rectf) MaxPoint() Vec2f {
	return Vec2f{r.Left + r.Width, r.Top + r.Height}
}

func (r Rectf) Size() Vec2f {
	return Vec2f{r.Width, r.Height}
}

// Rectd is a 2D rectangle with float64 coordinates.
type Rectd struct {
	Left, Top     float64
	Width, Height float64
}

func (r Rectd) MinPoint() Vec2d {
	return Vec2d{r.Left, r.Top}
}

func (r Rectd) MaxPoint() Vec2d {
	return Vec2d{r.Left + r.Width, r.Top + r.Height}
}

func (r Rectd) Size() Vec2d {
	return Vec2d{r.Width, r.Height}
}
