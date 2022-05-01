/*
 * The MIT License (MIT)
 *
 * Copyright © 2022 Hao Luo <haozzzzzzzz@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

// Package scanner implements a scanner for go-idl src text.
// It takes a []byte as src which can then be tokenized
// through repeated calls to the Scan method.
//
// Go-idl scanner imitated go/scanner.
package scanner

import (
	"fmt"
	"github.com/hacksomecn/go-idl/gopkg"
	"github.com/hacksomecn/go-idl/parser/ast"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"unicode"
	"unicode/utf8"
)

// ScanFiles scan .gidl files in dir
func ScanFiles(
	path string, // file or dir path
	modulePackagePath string, // package path with module name, if empty will find package for path
) (
	files []*ast.TokenFile,
	fileMap map[string]*ast.TokenFile, // file abs path -> *ast.TokenFile
	err error,
) {
	// get package name
	fileNames, err := FindIdlFiles(path)
	if err != nil {
		logrus.Errorf("get idl files failed. error: %s", err)
		return
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		logrus.Errorf("get abs path failed. path: %s, error: %s", path, err)
		return
	}

	isAbsPathExists, isAbsPathADir := PathExists(absPath)
	if !isAbsPathExists {
		err = fmt.Errorf("file path not exists. %s", absPath)
		return
	}

	absDirPath := absPath
	if !isAbsPathADir {
		absDirPath = filepath.Dir(absPath)
	}

	if modulePackagePath == "" {
		modulePackagePath, err = gopkg.GetModulePackagePath(absDirPath)
		if err != nil {
			logrus.Errorf(" error: %s", err)
			return
		}
	}

	files = make([]*ast.TokenFile, 0)
	fileMap = make(map[string]*ast.TokenFile, 0)
	if len(fileNames) == 0 {
		return
	}

	for _, fileName := range fileNames {
		absFilePath := fmt.Sprintf("%s/%s", absDirPath, fileName)
		pos := &ast.FilePos{
			Package:  modulePackagePath,
			FileName: fileName,
			Name:     fileName,
			FilePath: absFilePath,
		}
		tokenFile := ast.NewTokenFile(pos)

		fileMap[absFilePath] = tokenFile
		files = append(files, tokenFile)
	}

	return
}

type Scanner struct {
	file *ast.TokenFile
	src  []byte // file src

	// scanning state
	// refer to go/scanner
	ch         rune // current character
	offset     int  // current character offset
	lineOffset int  // current line offset

	lineNo int // current line number

	rdOffset int // reading offset (position after current character)

	ErrorList ErrorList
}

func NewScanner(
	astFile *ast.TokenFile,
	src []byte,
) (scanner *Scanner, err error) {
	scanner = &Scanner{
		file: astFile,
		src:  src,
	}

	err = scanner.init()
	if err != nil {
		logrus.Errorf("init scanner failed. error: %s", err)
		return
	}
	return
}

const (
	bom = 0xFEFF // byte order mark, only permitted as very first character
	eof = -1     // end of line char replacer
)

func (m *Scanner) init() (err error) {
	fileInfo, err := os.Stat(m.file.Pos.FilePath)
	if err != nil {
		logrus.Errorf("get file stat failed. path: %s, error: %s", m.file.Pos.FilePath, err)
		return
	}

	lenSrc := int64(len(m.src))

	if fileInfo.Size() != lenSrc {
		err = fmt.Errorf("file size (%d) does not match src len (%d)", fileInfo.Size(), lenSrc)
		return
	}

	m.lineNo = 1
	// read first character
	m.next()
	if m.ch == bom { // byte order mark
		m.next()
	}

	return
}

func (m *Scanner) next() {
	if m.rdOffset >= len(m.src) { // end of file
		m.ch = eof
		m.offset = len(m.src)
		if m.ch == '\n' {
			m.lineNo++
			m.lineOffset = m.offset
			m.file.AddLineOffset(m.offset)
		}
		return
	}

	// read a char
	m.offset = m.rdOffset
	if m.ch == '\n' { // new line position
		m.lineNo++
		m.lineOffset = m.offset
		m.file.AddLineOffset(m.offset)
	}

	chRune := rune(m.src[m.rdOffset])
	chWidth := 1
	switch {
	case chRune == 0:
		m.error(m.offset, "illegal character NUL")
	case chRune >= utf8.RuneSelf: // not ASCII
		chRune, chWidth = utf8.DecodeRune(m.src[m.rdOffset:]) // decode an rune
		if chRune == utf8.RuneError && chWidth == 1 {
			m.error(m.offset, "illegal UTF-8 encoding")
		} else if chRune == bom && m.offset > 0 {
			m.error(m.offset, "illegal type order mark")
		}
	}

	m.rdOffset += chWidth
	m.ch = chRune
}

func (m *Scanner) error(offset int, msg string) {
	tokenPos := &ast.TokenPos{
		FilePos: m.file.Pos,
		Offset:  offset,
	}

	m.ErrorList.Add(tokenPos, msg)
}

func (m *Scanner) errorf(offset int, msg string, args ...any) {
	m.error(offset, fmt.Sprintf(msg, args...))
}

// peek returns the byte following the most recently read character without
// advancing the scanner. If the scanner is at EOF, peek returns 0.
func (m *Scanner) peek() byte {
	if m.rdOffset < len(m.src) {
		return m.src[m.rdOffset]
	}
	return 0
}

// Scan scans the next token and returns the token position, the token,
// and its literal string if applicable. The source end is indicated by
// token.EOF.
//
// If the returned token is a literal (token.IDENT, token.INT, token.FLOAT,
// token.IMAG, token.CHAR, token.STRING) or token.COMMENT, the literal string
// has the corresponding value.
//
// If the returned token is a keyword, the literal string is the keyword.
//
// If the returned token is token.SEMICOLON, the corresponding
// literal string is ";" if the semicolon was present in the source,
// and "\n" if the semicolon was inserted because of a newline or
// at EOF.
//
// If the returned token is token.ILLEGAL, the literal string is the
// offending character.
//
// In all other cases, Scan returns an empty literal string.
//
// For more tolerant parsing, Scan will return a valid token if
// possible even if a syntax error was encountered. Thus, even
// if the resulting token sequence contains no illegal tokens,
// a client may not assume that no error occurred. Instead it
// must check the scanner's ErrorCount or the number of calls
// of the error handler, if there was one installed.
//
// Scan adds line information to the file added to the file
// set with Init. Token positions are relative to that file
// and thus relative to the file set.
//
func (m *Scanner) Scan() (pos *ast.TokenPos, tok ast.Token, lit string) {
	m.skipWhitespace()

	// current token start
	pos = m.pos(m.offset, m.lineNo, m.lineOffset)

	// determine token value
	switch ch := m.ch; {
	case isLetter(ch) || ch == '@': // word or decorator
		lit = m.scanIdentifier()
		tok = ast.LookupKeywordIdent(lit)

	case isDecimal(ch) || ch == '.' && isDecimal(rune(m.peek())): // number
		tok, lit = m.scanNumber()

	default:
		m.next() // always make progress
		switch ch {
		case eof:
			tok = ast.EOF

		case '\n':
			tok = ast.NEWLINE

		case '"':
			tok = ast.STRING
			lit = m.scanString()

		case '\'':
			tok = ast.CHAR
			lit = m.scanRune()

		case '`':
			tok = ast.STRING
			lit = m.scanRawString()

		case '/':
			if m.ch == '/' || m.ch == '*' {
				// comment
				tok = ast.COMMENT
				lit = m.scanComment()
			} else {
				tok = ast.IDENT
			}

		case '=':
			tok = ast.ASSIGN

		default:
			var isOperator bool
			tok, isOperator = ast.LookupOperatorToken(ch)
			if !isOperator {
				tok = ast.IDENT
			}
		}

	}
	return
}

func (m *Scanner) pos(
	offset int,
	line int,
	lineOffset int,
) (astPos *ast.TokenPos) {
	astPos = &ast.TokenPos{
		FilePos:    m.file.Pos,
		LineNo:     line,
		LineOffset: lineOffset,
		Offset:     offset,
	}
	return
}

func (m *Scanner) scanComment() string {
	// initial '/' already consumed; s.ch == '/' || s.ch == '*'
	offs := m.offset - 1 // position of initial '/'
	next := -1           // position immediately following the comment; < 0 means invalid comment
	numCR := 0

	if m.ch == '/' {
		//-style comment
		// (the final '\n' is not considered part of the comment)
		m.next()
		for m.ch != '\n' && m.ch >= 0 {
			if m.ch == '\r' {
				numCR++
			}
			m.next()
		}
		// if we are at '\n', the position following the comment is afterwards
		next = m.offset
		if m.ch == '\n' {
			next++
		}
		goto exit
	}

	/*-style comment */
	m.next()
	for m.ch >= 0 {
		ch := m.ch
		if ch == '\r' {
			numCR++
		}
		m.next()
		if ch == '*' && m.ch == '/' {
			m.next()
			next = m.offset
			goto exit
		}
	}

	m.error(offs, "comment not terminated")

exit:
	lit := m.src[offs:m.offset]

	// On Windows, a (//-comment) line may end in "\r\n".
	// Remove the final '\r' before analyzing the text for
	// line directives (matching the compiler). Remove any
	// other '\r' afterwards (matching the pre-existing be-
	// havior of the scanner).
	if numCR > 0 && len(lit) >= 2 && lit[1] == '/' && lit[len(lit)-1] == '\r' {
		lit = lit[:len(lit)-1]
		numCR--
	}

	if numCR > 0 {
		lit = stripCR(lit, lit[1] == '*')
	}

	return string(lit)
}

// find a new line to start parsing
func (m *Scanner) findLineEnd() bool {
	// initial '/' already consumed

	defer func(offs int) {
		// reset scanner state to where it was upon calling findLineEnd
		m.ch = '/'
		m.offset = offs
		m.rdOffset = offs + 1
		m.next() // consume initial '/' again
	}(m.offset - 1)

	// read ahead until a newline, EOF, or non-comment token is found
	for m.ch == '/' || m.ch == '*' {
		if m.ch == '/' {
			//-style comment always contains a newline
			return true
		}
		/*-style comment: look for newline */
		m.next()
		for m.ch >= 0 {
			ch := m.ch
			if ch == '\n' {
				return true
			}
			m.next()
			if ch == '*' && m.ch == '/' {
				m.next()
				break
			}
		}
		m.skipWhitespace() // s.insertSemi is set
		if m.ch < 0 || m.ch == '\n' {
			return true
		}
		if m.ch != '/' {
			// non-comment token
			return false
		}
		m.next() // consume '/'
	}

	return false
}

// scan single rune
func (m *Scanner) scanRune() (text string) {
	// '\'' opening already consumed
	offs := m.offset - 1

	valid := true
	n := 0
	for {
		ch := m.ch
		if ch == '\n' || ch < 0 {
			// only report error if we don't have one already
			if valid {
				m.error(offs, "rune literal not terminated")
				valid = false
			}
			break
		}
		m.next()
		if ch == '\'' {
			break
		}
		n++
		if ch == '\\' {
			if !m.scanEscape('\'') {
				valid = false
			}
			// continue to read to closing quote
		}
	}

	if valid && n != 1 {
		m.error(offs, "illegal rune literal")
	}

	return string(m.src[offs:m.offset])
}

func isLetter(ch rune) bool {
	return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return isDecimal(ch) || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

// scanIdentifier reads the string of valid identifier characters at s.offset.
// It must only be called when s.ch is known to be a valid letter.
//
// Be careful when making changes to this function: it is optimized and affects
// scanning performance significantly.
func (m *Scanner) scanIdentifier() string {
	offs := m.offset

	// Optimize for the common case of an ASCII identifier.
	//
	// Ranging over s.src[s.rdOffset:] lets us avoid some bounds checks, and
	// avoids conversions to runes.
	//
	// In case we encounter a non-ASCII character, fall back on the slower path
	// of calling into s.next().
	for rdOffset, b := range m.src[m.rdOffset:] {
		if 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z' || b == '_' || '0' <= b && b <= '9' {
			// Avoid assigning a rune for the common case of an ascii character.
			continue
		}
		m.rdOffset += rdOffset
		if 0 < b && b < utf8.RuneSelf {
			// Optimization: we've encountered an ASCII character that's not a letter
			// or number. Avoid the call into s.next() and corresponding set up.
			//
			// Note that s.next() does some line accounting if s.ch is '\n', so this
			// shortcut is only possible because we know that the preceding character
			// is not '\n'.
			m.ch = rune(b)
			m.offset = m.rdOffset
			m.rdOffset++
			goto exit
		}
		// We know that the preceding character is valid for an identifier because
		// scanIdentifier is only called when s.ch is a letter, so calling s.next()
		// at s.rdOffset resets the scanner state.
		m.next()
		for isLetter(m.ch) || isDigit(m.ch) {
			m.next()
		}
		goto exit
	}
	m.offset = len(m.src)
	m.rdOffset = len(m.src)
	m.ch = eof

exit:
	return string(m.src[offs:m.offset])
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= lower(ch) && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16 // larger than any legal digit val
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
func isHex(ch rune) bool     { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }

// digits accepts the sequence { digit | '_' }.
// If base <= 10, digits accepts any decimal digit but records
// the offset (relative to the source start) of a digit >= base
// in *invalid, if *invalid < 0.
// digits returns a bitset describing whether the sequence contained
// digits (bit 0 is set), or separators '_' (bit 1 is set).
func (m *Scanner) digits(base int, invalid *int) (digsep int) {
	if base <= 10 {
		max := rune('0' + base)
		for isDecimal(m.ch) || m.ch == '_' {
			ds := 1
			if m.ch == '_' {
				ds = 2
			} else if m.ch >= max && *invalid < 0 {
				*invalid = m.offset // record invalid rune offset
			}
			digsep |= ds
			m.next()
		}
	} else {
		for isHex(m.ch) || m.ch == '_' {
			ds := 1
			if m.ch == '_' {
				ds = 2
			}
			digsep |= ds
			m.next()
		}
	}
	return
}

func (m *Scanner) scanNumber() (ast.Token, string) {
	offs := m.offset
	tok := ast.ILLEGAL

	base := 10        // number base
	prefix := rune(0) // one of 0 (decimal), '0' (0-octal), 'x', 'o', or 'b'
	digsep := 0       // bit 0: digit present, bit 1: '_' present
	invalid := -1     // index of invalid digit in literal, or < 0

	// integer part
	if m.ch != '.' {
		tok = ast.INT
		if m.ch == '0' {
			m.next()
			switch lower(m.ch) {
			case 'x':
				m.next()
				base, prefix = 16, 'x'
			case 'o':
				m.next()
				base, prefix = 8, 'o'
			case 'b':
				m.next()
				base, prefix = 2, 'b'
			default:
				base, prefix = 8, '0'
				digsep = 1 // leading 0
			}
		}
		digsep |= m.digits(base, &invalid)
	}

	// fractional part
	if m.ch == '.' {
		tok = ast.FLOAT
		if prefix == 'o' || prefix == 'b' {
			m.error(m.offset, "invalid radix point in "+literalName(prefix))
		}
		m.next()
		digsep |= m.digits(base, &invalid)
	}

	if digsep&1 == 0 {
		m.error(m.offset, literalName(prefix)+" has no digits")
	}

	// exponent
	if e := lower(m.ch); e == 'e' || e == 'p' {
		switch {
		case e == 'e' && prefix != 0 && prefix != '0':
			m.errorf(m.offset, "%q exponent requires decimal mantissa", m.ch)
		case e == 'p' && prefix != 'x':
			m.errorf(m.offset, "%q exponent requires hexadecimal mantissa", m.ch)
		}
		m.next()
		tok = ast.FLOAT
		if m.ch == '+' || m.ch == '-' {
			m.next()
		}
		ds := m.digits(10, nil)
		digsep |= ds
		if ds&1 == 0 {
			m.error(m.offset, "exponent has no digits")
		}
	} else if prefix == 'x' && tok == ast.FLOAT {
		m.error(m.offset, "hexadecimal mantissa requires a 'p' exponent")
	}

	// suffix 'i'
	if m.ch == 'i' {
		tok = ast.IMAG
		m.next()
	}

	lit := string(m.src[offs:m.offset])
	if tok == ast.INT && invalid >= 0 {
		m.errorf(invalid, "invalid digit %q in %s", lit[invalid-offs], literalName(prefix))
	}
	if digsep&2 != 0 {
		if i := invalidSep(lit); i >= 0 {
			m.error(offs+i, "'_' must separate successive digits")
		}
	}

	return tok, lit
}

func literalName(prefix rune) string {
	switch prefix {
	case 'x':
		return "hexadecimal literal" // 16
	case 'o', '0':
		return "octal literal" // 8
	case 'b':
		return "binary literal" // 2
	}
	return "decimal literal" //  10
}

// invalidSep returns the index of the first invalid separator in x, or -1.
func invalidSep(x string) int {
	x1 := ' ' // prefix char, we only care if it's 'x'
	d := '.'  // digit, one of '_', '0' (a digit), or '.' (anything else)
	i := 0

	// a prefix counts as a digit
	if len(x) >= 2 && x[0] == '0' {
		x1 = lower(rune(x[1]))
		if x1 == 'x' || x1 == 'o' || x1 == 'b' {
			d = '0'
			i = 2
		}
	}

	// mantissa and exponent
	for ; i < len(x); i++ {
		p := d // previous digit
		d = rune(x[i])
		switch {
		case d == '_':
			if p != '0' {
				return i
			}
		case isDecimal(d) || x1 == 'x' && isHex(d):
			d = '0'
		default:
			if p == '_' {
				return i - 1
			}
			d = '.'
		}
	}
	if d == '_' {
		return len(x) - 1
	}

	return -1
}

// scanEscape parses an escape sequence where rune is the accepted
// escaped quote. In case of a syntax error, it stops at the offending
// character (without consuming it) and returns false. Otherwise
// it returns true.
func (m *Scanner) scanEscape(quote rune) bool {
	offs := m.offset

	var n int
	var base, max uint32
	switch m.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		m.next()
		return true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		m.next()
		n, base, max = 2, 16, 255
	case 'u':
		m.next()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		m.next()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		msg := "unknown escape sequence"
		if m.ch < 0 {
			msg = "escape sequence not terminated"
		}
		m.error(offs, msg)
		return false
	}

	var x uint32
	for n > 0 {
		d := uint32(digitVal(m.ch))
		if d >= base {
			msg := fmt.Sprintf("illegal character %#U in escape sequence", m.ch)
			if m.ch < 0 {
				msg = "escape sequence not terminated"
			}
			m.error(m.offset, msg)
			return false
		}
		x = x*base + d
		m.next()
		n--
	}

	if x > max || 0xD800 <= x && x < 0xE000 {
		m.error(offs, "escape sequence is invalid Unicode code point")
		return false
	}

	return true
}

func (m *Scanner) scanString() string {
	// '"' opening already consumed
	offs := m.offset - 1

	for {
		ch := m.ch
		if ch == '\n' || ch < 0 {
			m.error(offs, "string literal not terminated")
			break
		}
		m.next()
		if ch == '"' {
			break
		}
		if ch == '\\' {
			m.scanEscape('"')
		}
	}

	return string(m.src[offs:m.offset])
}

func (m *Scanner) scanRawString() string {
	// '`' opening already consumed
	offs := m.offset - 1

	hasCR := false
	for {
		ch := m.ch
		if ch < 0 {
			m.error(offs, "raw string literal not terminated")
			break
		}
		m.next()
		if ch == '`' {
			break
		}
		if ch == '\r' {
			hasCR = true
		}
	}

	lit := m.src[offs:m.offset]
	if hasCR {
		lit = stripCR(lit, false)
	}

	return string(lit)
}

func stripCR(b []byte, comment bool) []byte {
	c := make([]byte, len(b))
	i := 0
	for j, ch := range b {
		// In a /*-style comment, don't strip \r from *\r/ (incl.
		// sequences of \r from *\r\r...\r/) since the resulting
		// */ would terminate the comment too early unless the \r
		// is immediately following the opening /* in which case
		// it's ok because /*/ is not closed yet (issue #11151).
		if ch != '\r' || comment && i > len("/*") && c[i-1] == '*' && j+1 < len(b) && b[j+1] == '/' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}

func (m *Scanner) skipWhitespace() {
	for m.ch == ' ' || m.ch == '\t' || m.ch == '\n' || m.ch == '\r' {
		m.next()
	}
}
