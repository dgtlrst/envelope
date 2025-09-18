package edit

type CursorPointer struct {
	X int
	Y int
}

func NewCursor(x, y int) *CursorPointer {
	return &CursorPointer{x, y}
}

type Cursor interface {
	GetPosition() (x, y int)
	SetPosition(x, y int, buffer Buffer)
	Clamp(buffer Buffer)

	MoveLeft(buffer Buffer)
	MoveRight(buffer Buffer)
	MoveUp(buffer Buffer)
	MoveDown(buffer Buffer)
}

func (c *CursorPointer) GetPosition() (x, y int) {
	return c.X, c.Y
}

func (c *CursorPointer) SetPosition(x, y int, buffer Buffer) {
	c.X = x
	c.Y = y
	c.Clamp(buffer)
}

// "clamps" the cursor to a valid position
// usually called after manipulating the buffer
// such as SetPosition() or DeleteRune()
// we need to deprecate this function
// and remove the need for the user to call it altogether
func (c *CursorPointer) Clamp(buffer Buffer) {
	if c.Y < 0 {
		c.Y = 0
	} else if c.Y >= buffer.LineCount() {
		c.Y = buffer.LineCount() - 1
	}
	lineLen := len(buffer.GetLine(c.Y))
	if c.X < 0 {
		c.X = 0
	} else if c.X > lineLen {
		c.X = lineLen
	}
}

func (c *CursorPointer) MoveLeft(buffer Buffer) {
	if c.X == 0 && c.Y == 0 {
		return
	}

	if c.X > 0 {
		c.X--
	} else if c.X == 0 && c.Y > 0 {
		c.Y--
		c.X = len(buffer.GetLine(c.Y))
	}
}

func (c *CursorPointer) MoveRight(buffer Buffer) {
	lineLen := len(buffer.GetLine(c.Y))
	if c.X < lineLen {
		c.X++
	} else if c.Y < buffer.LineCount()-1 {
		c.Y++
		c.X = 0
	}
}

func (c *CursorPointer) MoveUp(buffer Buffer) {
	if c.Y > 0 {
		c.Y--
		lineLen := len(buffer.GetLine(c.Y))
		if c.X > lineLen {
			c.X = lineLen
		}
	}
}

func (c *CursorPointer) MoveDown(buffer Buffer) {
	if c.Y < buffer.LineCount()-1 {
		c.Y++
		lineLen := len(buffer.GetLine(c.Y))
		if c.X > lineLen {
			c.X = lineLen
		}
	}
}
