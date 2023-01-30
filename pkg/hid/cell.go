package hid

type Cell struct {
	char     rune
	prevChar rune
	isDirty  bool
}

func (cell *Cell) SetIsDirty(dirty bool) {
	cell.isDirty = dirty
}

func (cell *Cell) SetChar(char rune) {
	cell.prevChar = cell.char
	cell.char = char

	if cell.GetChar() != cell.GetPrevChar() {
		cell.SetIsDirty(true)
	} else {
		cell.SetIsDirty(false)
	}
}

func (cell *Cell) GetChar() rune {
	return cell.char
}
func (cell *Cell) GetPrevChar() rune {
	return cell.prevChar
}
func (cell *Cell) IsDirty() bool {
	return cell.isDirty
}
