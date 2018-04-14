package merkletree

import (
	"fmt"
	"strings"
)

func (m *Proof) ToString(h hashBytesFunc, f checksumToStrFunc) string {
	var lines []string

	parts := m.parts
	if len(parts) == 0 {
		return "" // checksums don't match up with receiver
	}

	lines = append(lines, fmt.Sprintf("route from %s (leaf) to root:", f(m.target)))
	lines = append(lines, "")

	var prev = m.target
	var curr []byte
	for i := 0; i < len(parts); i++ {
		if parts[i].isRight {
			curr = h(append(prev, parts[i].checksum...))
			lines = append(lines, fmt.Sprintf("%s + %s = %s", f(prev), f(parts[i].checksum), f(curr)))
		} else {
			curr = h(append(parts[i].checksum, prev...))
			lines = append(lines, fmt.Sprintf("%s + %s = %s", f(parts[i].checksum), f(prev), f(curr)))
		}
		prev = curr
	}

	return strings.Join(lines, "\n")
}

func (t *Tree) ToString(f checksumToStrFunc, n int) string {
	return t.root.ToString(f, n)
}

func (l *Leaf) ToString(f checksumToStrFunc, n int) string {
	return fmt.Sprintf("\n"+indent(n, "(L root: %s)"), f(l.checksum))
}

func (b *Branch) ToString(f checksumToStrFunc, n int) string {
	c := f(b.checksum)
	l := b.left.ToString(f, n+2)
	r := b.right.ToString(f, n+2)

	return fmt.Sprintf("\n"+indent(n, "(B root: %s %s %s)"), c, l, r)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// HELPERS
//////////

func indent(spaces int, orig string) string {
	str := ""
	for i := 0; i < spaces; i++ {
		str += " "
	}

	return str + orig
}
