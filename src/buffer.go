package lexer

type Buffer struct {
	buf      []rune
	size     int
	position int
	isFull   bool
	isEmpty  bool
}

func NewBuffer(size int) *Buffer {
	return &Buffer{
		buf:      make([]rune, 0),
		size:     size,
		isFull:   false,
		isEmpty:  true,
		position: 0,
	}
}

func (b *Buffer) Push(r rune) {
	if b.isFull {
		b.Pop()
	}
	b.buf = append(b.buf, r)
	if len(b.buf) == b.size {
		b.isFull = true
	}
	b.isEmpty = false
}

func (b *Buffer) Pop() (result rune) {
	length := len(b.buf)
	result = b.buf[length-1]
	b.buf[0] = 0
	b.buf = b.buf[1:]
	if len(b.buf) == 0 {
		b.isEmpty = true
	}
	b.isFull = false
	return
}

func (b *Buffer) GetCurrent() rune {
	return b.buf[b.position]
}

func (b *Buffer) IsFull() bool {
	return b.isFull
}

func (b *Buffer) IsEmpty() bool {
	return b.isEmpty
}
func (b *Buffer) CurrentAtHead() bool {
	return (b.position + 1) >= len(b.buf)
}
