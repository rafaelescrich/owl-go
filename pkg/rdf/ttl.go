package rdf

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"
)

// token represents one ttl expression
type token struct {
	typ   int    // token type
	value string // value of token
}

// parser parses a ttl document
type parser struct {
	reader       *bufio.Reader     // reader of turtle document
	runes        []rune            // turtle document as rune slice
	posStatement int               // starting position of current statement
	prefix       map[string]string // prefixes
	base         string            // base IRI
	curSubject   string            // current Subject
	curPredicate string            // current Predicate
}

// DecodeTTL decodes a ttl input to rdf triples
func DecodeTTL(input io.Reader) (trip []Triple, err error) {
	p := &parser{reader: bufio.NewReader(input), prefix: make(map[string]string)}
	err = p.parseRunes()
	if err != nil {
		return
	}
	for {
		err = p.parseStatement()
		if err != nil {
			break
		}
	}
	fmt.Println(p.prefix)
	fmt.Println(p.base)
	// for i := range p.runes {
	// 	b := make([]byte, 3)
	// 	utf8.EncodeRune(b, p.runes[i])
	// 	fmt.Println(string(b))
	// }
	// fmt.Println(p.runes)
	return
}

// parseRunes parses all runes from the reader and omits empty lines and comments
func (p *parser) parseRunes() (err error) {
	for {
		var line []byte
		line, err = p.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if len(line) == 0 {
			continue
		}
		pos := 0
		// omit new line at end of slice
		for pos < len(line)-1 {
			var r rune
			var s int
			r, s = utf8.DecodeRune(line[pos:])
			if r == utf8.RuneError {
				err = errors.New("Rune error")
			}
			if pos == 0 && r == '#' {
				break
			}
			p.runes = append(p.runes, r)
			pos += s
		}
	}
	return
}

// parseStatement decodes one statement beginning from the position stored in parser
func (p *parser) parseStatement() (err error) {
	if len(p.runes) <= p.posStatement {
		return
	}
	length := 0
	switch p.runes[p.posStatement] {
	case '@':
		// @prefix or @base
		length, err = p.parseDirective(p.posStatement + 1)
		length++
	default:
		if p.isEqual(p.posStatement, "BASE ") || p.isEqual(p.posStatement, "PREFIX ") {
			// sparqlBase or sparqlPrefix
			length, err = p.parseDirective(p.posStatement + 1)
		} else {
			length, err = p.parseTriples(p.posStatement)
		}

		return
	}
	if err != nil {
		return
	}
	p.posStatement += length
	return
}

// parseDirective decodes one directive beginning from the specified position
func (p *parser) parseDirective(pos int) (length int, err error) {
	if len(p.runes) <= pos {
		err = errors.New("Invalid directive " + strconv.Itoa(pos))
		return
	}
	switch p.runes[pos] {
	case 'p':
		var prefix, iri string
		var tempLength int
		if p.isEqual(pos, "prefix") {
			length = 6
			length += p.consumeWP(pos + length)
			prefix, tempLength, err = p.parsePrefix(pos + length)
			if err != nil {
				return
			}
			length += tempLength
			length += p.consumeWP(pos + length)
			iri, tempLength, err = p.parseIRIRef(pos + length)
			if err != nil {
				return
			}
			length += tempLength
			p.prefix[prefix] = iri
		} else {
			err = errors.New("Invalid directive " + string(p.runes[pos]) + strconv.Itoa(pos))
			return
		}
	case 'b':
		var iri string
		var tempLength int
		if p.isEqual(pos, "base") {
			length = 4
			length += p.consumeWP(pos + length)
			iri, tempLength, err = p.parseIRIRef(pos + length)
			if err != nil {
				return
			}
			length += tempLength
			p.base = iri
		} else {
			err = errors.New("Invalid directive " + string(p.runes[pos]) + strconv.Itoa(pos))
			return
		}
	case 'P':
		var prefix, iri string
		var tempLength int
		if p.isEqual(pos, "PREFIX") {
			length = 6
			length += p.consumeWP(pos + length)
			prefix, tempLength, err = p.parsePrefix(pos + length)
			if err != nil {
				return
			}
			length += tempLength
			length += p.consumeWP(pos + length)
			iri, tempLength, err = p.parseIRIRef(pos + length)
			if err != nil {
				return
			}
			length += tempLength
			p.prefix[prefix] = iri
		} else {
			err = errors.New("Invalid directive " + string(p.runes[pos]) + strconv.Itoa(pos))
			return
		}
	case 'B':
		var iri string
		var tempLength int
		if p.isEqual(pos, "BASE") {
			length = 4
			length += p.consumeWP(pos + length)
			iri, tempLength, err = p.parseIRIRef(pos + length)
			if err != nil {
				return
			}
			length += tempLength
			p.base = iri
		} else {
			err = errors.New("Invalid directive " + string(p.runes[pos]) + strconv.Itoa(pos))
			return
		}
	default:
		err = errors.New("Invalid directive " + string(p.runes[pos]) + strconv.Itoa(pos))
		return
	}
	// consumer dot
	length += p.consumeWP(pos + length)
	if p.isEqual(pos+length, ".") {
		length++
	} else {
		err = errors.New("No dot")
		return
	}
	length += p.consumeWP(pos + length)
	return
}

// parsePrefix parses one prefix beginning from the specified position
func (p *parser) parsePrefix(pos int) (prefix string, length int, err error) {
	if len(p.runes) <= pos {
		err = errors.New("Prefix error " + strconv.Itoa(pos))
	}
	prefix, length, err = p.parseUntil(pos, ':')
	length++
	return
}

// parseTriples parses all tripls in a statement
func (p *parser) parseTriples(pos int) (length int, err error) {
	if len(p.runes) <= pos {
		err = errors.New("Invalid triples " + strconv.Itoa(pos))
		return
	}
	// iri, _, _ := p.parseIRI(pos)
	// fmt.Println(iri)
	return
}

// parseIRI parses the next iri
func (p *parser) parseIRI(pos int) (iri string, length int, err error) {
	if len(p.runes) <= pos {
		err = errors.New("IRI error " + strconv.Itoa(pos))
	}
	if p.runes[pos] == '<' {
		iri, length, err = p.parseIRIRef(pos)
	} else {
		iri, length, err = p.parsePrefixedName(pos)
	}
	return
}

// parseIRI parses IRIRef <iri>
func (p *parser) parseIRIRef(pos int) (iri string, length int, err error) {
	if len(p.runes) <= pos {
		err = errors.New("IRI error " + strconv.Itoa(pos))
	}
	if p.runes[pos] != '<' {
		err = errors.New("No IRI: " + string(p.runes[pos]) + strconv.Itoa(pos))
		return
	}
	iri, length, err = p.parseUntil(pos+1, '>')
	length += 2
	return
}

// parsePrefixedName parses prefixed name prefix:name
func (p *parser) parsePrefixedName(pos int) (iri string, length int, err error) {
	if len(p.runes) <= pos {
		err = errors.New("IRI error " + strconv.Itoa(pos))
	}
	var prefix string
	prefix, length, err = p.parsePrefix(pos)
	if err != nil {
		return
	}
	ok := false
	if iri, ok = p.prefix[prefix]; !ok {
		err = errors.New("No such prefix " + prefix)
	}
	var name string
	var tempLength int
	name, tempLength, err = p.parseUntil(pos+length, ' ')
	if err != nil {
		return
	}
	iri = iri + name
	length += tempLength
	return
}

// isEqual checks if runes at position equal specified string
func (p *parser) isEqual(pos int, comp string) (ok bool) {
	ok = false
	compRune, err := toRunes([]byte(comp))
	if err != nil {
		return
	}
	if len(p.runes) <= pos+len(compRune) {
		return
	}
	for i := range compRune {
		if compRune[i] != p.runes[pos+i] {
			return
		}
	}
	ok = true
	return
}

// toRunes transforms a byte slice to a rune slice
func toRunes(in []byte) (out []rune, err error) {
	pos := 0
	for pos < len(in) {
		r, s := utf8.DecodeRune(in[pos:])
		if r == utf8.RuneError {
			err = errors.New("Rune error")
		}
		out = append(out, r)
		pos += s
	}
	return
}

// consumeWP returns number of consecutive white spaces
func (p *parser) consumeWP(pos int) (num int) {
	num = 0
	for {
		if len(p.runes) <= pos {
			break
		}
		if p.runes[pos] == ' ' {
			num++
			pos++
		} else {
			break
		}
	}
	return
}

// parseUntil returns a string from current position to next occurance of specified rune
func (p *parser) parseUntil(pos int, delim rune) (res string, length int, err error) {
	length = 0
	var r []rune
	for {
		if len(p.runes) <= pos+length {
			err = errors.New("No delimiter")
			return
		}
		if p.runes[pos+length] == delim {
			break
		} else {
			r = append(r, p.runes[pos+length])
			length++
		}
	}
	res = string(r)
	return
}