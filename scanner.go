// 
// Copyright (c) 2014-2016 Brian W. Wolter, All rights reserved.
// Ego - an embedded Go parser / compiler
// 
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
// 
//   * Redistributions of source code must retain the above copyright notice, this
//     list of conditions and the following disclaimer.
// 
//   * Redistributions in binary form must reproduce the above copyright notice,
//     this list of conditions and the following disclaimer in the documentation
//     and/or other materials provided with the distribution.
//     
//   * Neither the names of Brian W. Wolter nor the names of the contributors may
//     be used to endorse or promote products derived from this software without
//     specific prior written permission.
//     
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
// INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
// BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
// LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE
// OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
// OF THE POSSIBILITY OF SUCH DAMAGE.
// 
// --
// 
// This scanner incorporates routines from the Go package text/scanner:
// http://golang.org/src/pkg/text/scanner/scanner.go
// 
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// 
// http://golang.org/LICENSE
// 

package ego

import (
  "fmt"
  "math"
  "strings"
  "strconv"
  "unicode"
  "unicode/utf8"
)

/**
 * A text span
 */
type span struct {
  text      string
  offset    int
  length    int
}

/**
 * Span (unquoted) excerpt
 */
func (s span) excerpt() string {
  max := float64(len(s.text))
  return s.text[int(math.Max(0, math.Min(max, float64(s.offset)))):int(math.Min(max, float64(s.offset+s.length)))]
}

/**
 * Span (quoted) excerpt
 */
func (s span) String() string {
  return strconv.Quote(s.excerpt())
}

/**
 * Create a new span that encompasses all the provided spans. The underlying text is taken from the first span.
 */
func encompass(a ...span) span {
  var t string
  min, max := 0, 0
  for i, e := range a {
    if i == 0 {
      min, max = e.offset, e.offset + e.length
      t = e.text
    }else{
      if e.offset < min {
        min = e.offset
      }
      if e.offset + e.length > max {
        max = e.offset + e.length
      }
    }
  }
  return span{t, min, max - min}
}

/**
 * Numeric type
 */
type numericType int

const (
  numericInteger numericType = iota
  numericFloat
)

/**
 * Token type
 */
type tokenType int

/**
 * Token types
 */
const (
  
  tokenError tokenType  = iota
  tokenEOF
  tokenVerbatim
  tokenMeta
  tokenBlock
  tokenClose
  tokenAtem
  
  tokenString
  tokenNumber
  tokenIdentifier
  
  tokenIf
  tokenElse
  tokenFor
  tokenBreak
  tokenContinue
  
  tokenTrue
  tokenFalse
  tokenNil
  tokenRange
  
  tokenLParen           = '('
  tokenRParen           = ')'
  tokenLBracket         = '['
  tokenRBracket         = ']'
  tokenDot              = '.'
  tokenComma            = ','
  tokenSemi             = ';'
  tokenColon            = ':'
  tokenAdd              = '+'
  tokenSub              = '-'
  tokenMul              = '*'
  tokenDiv              = '/'
  tokenMod              = '%'
  tokenAssign           = '='
  tokenLess             = '<'
  tokenGreater          = '>'
  tokenBang             = '!'
  tokenAmp              = '&'
  tokenPipe             = '|'
  
  tokenPrefixAdd        = 1 << 16
  tokenInc              = tokenPrefixAdd | '+'
  
  tokenPrefixSub        = 1 << 17
  tokenDec              = tokenPrefixSub | '-'
  
  tokenPrefixAmp        = 1 << 18
  tokenLogicalAnd       = tokenPrefixAmp | '&'
  
  tokenPrefixPipe       = 1 << 19
  tokenLogicalOr        = tokenPrefixPipe | '|'
  
  tokenSuffixEqual      = 1 << 20
  tokenEqual            = tokenSuffixEqual | '='
  tokenAddEqual         = tokenSuffixEqual | '+'
  tokenSubEqual         = tokenSuffixEqual | '-'
  tokenMulEqual         = tokenSuffixEqual | '*'
  tokenDivEqual         = tokenSuffixEqual | '/'
  tokenLessEqual        = tokenSuffixEqual | '<'
  tokenGreaterEqual     = tokenSuffixEqual | '>'
  tokenNotEqual         = tokenSuffixEqual | '!'
  tokenBitwiseAndEqual  = tokenSuffixEqual | '&'
  tokenBitwiseOrEqual   = tokenSuffixEqual | '|'
  tokenAssignSpecial    = tokenSuffixEqual | ':'
  
)

/**
 * Token type string
 */
func (t tokenType) String() string {
  switch t {
    case tokenError:
      return "error"
    case tokenEOF:
      return "EOF"
    case tokenVerbatim:
      return "verbatim"
    case tokenMeta:
      return "@"
    case tokenBlock:
      return "{..."
    case tokenClose:
      return "...}"
    case tokenAtem:
      return "#"
    case tokenString:
      return "string"
    case tokenNumber:
      return "number"
    case tokenIdentifier:
      return "ident"
    case tokenIf:
      return "if"
    case tokenElse:
      return "else"
    case tokenFor:
      return "for"
    case tokenBreak:
      return "break"
    case tokenContinue:
      return "continue"
    case tokenTrue:
      return "true"
    case tokenFalse:
      return "false"
    case tokenNil:
      return "nil"
    case tokenRange:
      return "range"
    case tokenInc:
      return "++"
    case tokenDec:
      return "--"
    case tokenLogicalAnd:
      return "&&"
    case tokenLogicalOr:
      return "||"
    case tokenEqual:
      return "=="
    case tokenAddEqual:
      return "+="
    case tokenSubEqual:
      return "+="
    case tokenMulEqual:
      return "*="
    case tokenDivEqual:
      return "/="
    case tokenLessEqual:
      return "<="
    case tokenGreaterEqual:
      return ">="
    case tokenNotEqual:
      return "!="
    case tokenAssignSpecial:
      return ":="
    case tokenBitwiseAndEqual:
      return "&="
    case tokenBitwiseOrEqual:
      return "|="
    default:
      if t < 128 {
        return fmt.Sprintf("'%v'", string(t))
      }else{
        return fmt.Sprintf("%U", t)
      }
  }
}

/**
 * Token stuff
 */
const (
  meta = '@'
  eof  = -1
)

/**
 * A token
 */
type token struct {
  span      span
  which     tokenType
  value     interface{}
}

/**
 * Stringer
 */
func (t token) String() string {
  switch t.which {
    case tokenError:
      return fmt.Sprintf("<%v %v %v>", t.which, t.span, t.value)
    default:
      return fmt.Sprintf("<%v %v>", t.which, t.span)
  }
}

/**
 * A scanner action
 */
type scannerAction func(*scanner) scannerAction

/**
 * A scanner error
 */
type scannerError struct {
  message   string
  span      span
  cause     error
}

/**
 * Error
 */
func (s *scannerError) Error() string {
  if s.cause != nil {
    return fmt.Sprintf("%s: %v\n%v", s.message, s.cause, excerptCallout.FormatExcerpt(s.span))
  }else{
    return fmt.Sprintf("%s\n%v", s.message, excerptCallout.FormatExcerpt(s.span))
  }
}

const (
  mtypeExpr     = iota
  mtypeControl  = iota
)

/**
 * A scanner
 */
type scanner struct {
  text    string
  index   int
  width   int // current rune width
  start   int // token start position
  depth   int // meta depth
  tokens  chan token
  state   scannerAction
  paren   int
  mtype   int
}

/**
 * Create a scanner
 */
func newScanner(text string) *scanner {
  t := make(chan token, 64 /* several tokens may be produced in one iteration */)
  return &scanner{text, 0, 0, 0, 0, t, startAction, 0, 0}
}

/**
 * Scan and produce a token
 */
func (s *scanner) scan() token {
  for {
    select {
      case t := <- s.tokens:
        return t
      default:
        if s.state == nil {
          return token{span{s.text, len(s.text), 0}, tokenEOF, nil}
        }else{
          s.state = s.state(s)
        }
    }
  }
}

/**
 * Create an error
 */
func (s *scanner) errorf(where span, cause error, format string, args ...interface{}) *scannerError {
  return &scannerError{fmt.Sprintf(format, args...), where, cause}
}

/**
 * Emit a token
 */
func (s *scanner) emit(t token) {
  s.tokens <- t
  s.start = t.span.offset + t.span.length
}

/**
 * Emit an error and return a nil action
 */
func (s *scanner) error(err *scannerError) scannerAction {
  s.tokens <- token{err.span, tokenError, err}
  return nil
}

/**
 * Obtain the next rune from input without consuming it
 */
func (s *scanner) peek() rune {
  r := s.next()
  s.backup()
  return r
}

/**
 * Consume the next rune from input
 */
func (s *scanner) next() rune {
  
  if s.index >= len(s.text) {
    s.width = 0
    return eof
  }
  
  r, w := utf8.DecodeRuneInString(s.text[s.index:])
  s.index += w
  s.width  = w
  
  return r
}

/**
 * Match ahead
 */
func (s *scanner) match(text string) bool {
  return s.matchAt(s.index, text)
}

/**
 * Match ahead
 */
func (s *scanner) matchAt(index int, text string) bool {
  i := index
  if i < 0 {
    return false
  }
  
  for n := 0; n < len(text); {
    if i >= len(s.text) {
      return false
    }
    
    r, w := utf8.DecodeRuneInString(s.text[i:])
    i += w
    c, z := utf8.DecodeRuneInString(text[n:])
    n += z
    
    if r != c {
      return false
    }
  }
  
  return true
}

/**
 * Find the next occurance of any character in the specified string
 */
func (s *scanner) findFrom(index int, any string, invert bool) int {
  i := index
  if !invert {
    return strings.IndexAny(s.text[i:], any)
  }else{
    for {
      
      if i >= len(s.text) {
        return -1
      }
      
      r, w := utf8.DecodeRuneInString(s.text[i:])
      if !strings.ContainsRune(any, r) {
        return i
      }else{
        i += w
      }
      
    }
  }
}

/**
 * Shuffle the token start to the current index
 */
func (s *scanner) ignore() {
  s.start = s.index
}

/**
 * Move directly to an index
 */
func (s *scanner) move(index int) {
  s.start, s.index = index, index
}

/**
 * Unconsume the previous rune from input (this can be called only once
 * per invocation of next())
 */
func (s *scanner) backup() {
  s.index -= s.width
}

/**
 * Skip past a rune that was previously peeked
 */
func (s *scanner) skip() {
  s.index += s.width
}

/**
 * Skip past a rune that was previously peeked and ignore it
 */
func (s *scanner) skipAndIgnore() {
  s.skip()
  s.ignore()
}

/**
 * Start action
 */
func startAction(s *scanner) scannerAction {
  
  for {
    
    if s.index < len(s.text) {
      r := s.text[s.index]
      switch {
        
        case r == meta:
          if s.index > s.start {
            s.emit(token{span{s.text, s.start, s.index - s.start}, tokenVerbatim, s.text[s.start:s.index]})
          }
          return preludeAction
          
        case r == '}' && s.depth > 0:
          if s.index > s.start {
            s.emit(token{span{s.text, s.start, s.index - s.start}, tokenVerbatim, s.text[s.start:s.index]})
          }
          return closeAction
          
      }
    }
    
    if r := s.next(); r == eof {
      break
    }else if r == '\\' {
      var t span
      switch r := s.next(); {
        
        case r == eof:
          break
          
        case r == '\\':
          t = span{s.text, s.start, s.index - s.start - 1}    // verbatim up to first '\', ignore second (literal '\')
          s.emit(token{t, tokenVerbatim, t.excerpt()})
          s.ignore()
          
        case r == '@' || r == '{' || r == '}':
          if s.index - 2 > s.start {
            t = span{s.text, s.start, s.index - s.start - 2}  // verbatim up to '\', exclusive (we know the widths of runes '\' and r)
            s.emit(token{t, tokenVerbatim, t.excerpt()})
          }
          
          t = span{s.text, s.index - 1, 1}                    // emit r, continue (literal r)
          s.emit(token{t, tokenVerbatim, t.excerpt()})
          s.ignore()
          
      }
    }
    
  }
  
  // emit the last verbatim block, if we have one
  if s.index > s.start {
    s.emit(token{span{s.text, s.start, s.index - s.start}, tokenVerbatim, s.text[s.start:s.index]})
  }
  
  // emit end of input
  s.emit(token{span{s.text, len(s.text), 0}, tokenEOF, nil})
  
  // we're done
  return nil
}

/**
 * Prelude action. This introduces a meta expression or control structure.
 */
func preludeAction(s *scanner) scannerAction {
  s.emit(token{span{s.text, s.index, 1}, tokenMeta, "@"})
  s.next() // skip the '@' delimiter
  
  // if the meta begins with an open parenthesis it is an expression, otherwise
  // it is a control structure (if, for, etc)
  if s.next() == '(' {
    s.mtype = mtypeExpr
  }else{
    s.mtype = mtypeControl
  }
  s.backup()
  
  return metaAction
}

/**
 * Meta action.
 */
func metaAction(s *scanner) scannerAction {
  
  for {
    switch r := s.next(); {
      
      case r == eof:
        return s.error(s.errorf(span{s.text, s.index, 1}, nil, "Unexpected end-of-input"))
        
      case unicode.IsSpace(r):
        s.ignore()
        
      case r == '{': // open verbatim
        s.backup()
        return blockAction
        
      case r == '"':
        // consume the open '"'
        return stringAction
        
      case r >= '0' && r <= '9':
        s.backup()
        return numberAction
        
      case r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'):
        s.backup()
        return identifierAction
        
      case r == '(':
        s.paren++
        s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(r), string(r)})
        return metaAction
        
      case r == ')':
        s.paren--
        s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(r), string(r)})
        if s.paren == 0 && s.mtype == mtypeExpr {
          return startAction
        }else{
          return metaAction
        }
        
      case r == '[' || r == ']' || r == '.' || r == ',' || r == ';':
        s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(r), string(r)})
        return metaAction
        
      case r == '&':
        if n := s.next(); n == '=' {
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(tokenSuffixEqual | r), string(r)})
        }else if n == '&' {
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(tokenPrefixAmp | r), string(r)})
        }else{
          s.backup()
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(r), string(r)})
        }
        return metaAction
        
      case r == '|':
        if n := s.next(); n == '=' {
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(tokenSuffixEqual | r), string(r)})
        }else if n == '|' {
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(tokenPrefixPipe | r), string(r)})
        }else{
          s.backup()
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(r), string(r)})
        }
        return metaAction
        
      case r == '+':
        if n := s.next(); n == '=' {
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(tokenSuffixEqual | r), string(r)})
        }else if n == '+' {
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(tokenPrefixAdd | r), string(r)})
        }else{
          s.backup()
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(r), string(r)})
        }
        return metaAction
      
      case r == '-':
        if n := s.next(); n == '=' {
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(tokenSuffixEqual | r), string(r)})
        }else if n == '-' {
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(tokenPrefixSub | r), string(r)})
        }else{
          s.backup()
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(r), string(r)})
        }
        return metaAction
      
      case r == ':':
        if n := s.next(); n == '=' {
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(tokenSuffixEqual | r), string(r)})
        }else{
          s.backup()
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(r), string(r)})
        }
        return metaAction
      
      case r == '=' || r == '!' || r == '<' || r == '>' || r == ':' || r == '*' || r == '/':
        if n := s.next(); n == '=' {
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(tokenSuffixEqual | r), string(r)})
        }else{
          s.backup()
          s.emit(token{span{s.text, s.start, s.index - s.start}, tokenType(r), string(r)})
        }
        return metaAction
        
      default:
        return s.error(s.errorf(span{s.text, s.index, 1}, nil, "Syntax error in meta"))
        
    }
  }
  
  return startAction
}

/**
 * Block { ... }
 */
func blockAction(s *scanner) scannerAction {
  s.depth++ // increment the meta depth
  s.emit(token{span{s.text, s.index, 1}, tokenBlock, nil})
  s.next()  // skip the '{' delimiter
  return startAction
}

/**
 * Close a block. This closes a meta expression or control structure. As a special case, the
 * dangling 'else' following what is otherwise a closed structure begins a new block.
 */
func closeAction(s *scanner) scannerAction {
  
  s.emit(token{span{s.text, s.index, 1}, tokenClose, nil})
  s.next()  // skip the '}' delimiter
  s.depth-- // decrement the meta depth
  
  f := s.findFrom(s.index, " \n\r\t\v", true)
  if s.matchAt(f, "else") {
    s.move(f)
    return identifierAction
  }else if s.matchAt(f, "\\else") {
    // move past the \ escape, which should be ignored in this specific case
    s.move(f+1)
  }
  
  return startAction
}

/**
 * Quoted string
 */
func stringAction(s *scanner) scannerAction {
  if v, err := s.scanString('"', '\\'); err != nil {
    s.error(s.errorf(span{s.text, s.index, 1}, err, "Invalid string"))
  }else{
    s.emit(token{span{s.text, s.start, s.index - s.start}, tokenString, v})
  }
  return metaAction
}

/**
 * Number string
 */
func numberAction(s *scanner) scannerAction {
  if v, _, err := s.scanNumber(); err != nil {
    s.error(s.errorf(span{s.text, s.index, 1}, err, "Invalid number"))
  }else{
    s.emit(token{span{s.text, s.start, s.index - s.start}, tokenNumber, v})
  }
  return metaAction
}

/**
 * Identifier
 */
func identifierAction(s *scanner) scannerAction {
  
  v, err := s.scanIdentifier()
  if err != nil {
    s.error(s.errorf(span{s.text, s.index, 1}, err, "Invalid identifier"))
  }
  
  t := span{s.text, s.start, s.index - s.start}
  switch v {
    case "if":
      s.emit(token{t, tokenIf, v})
    case "else":
      s.emit(token{t, tokenElse, v})
    case "for":
      s.emit(token{t, tokenFor, v})
    case "break":
      s.emit(token{t, tokenBreak, v})
    case "continue":
      s.emit(token{t, tokenContinue, v})
    case "true":
      s.emit(token{t, tokenTrue, v})
    case "false":
      s.emit(token{t, tokenFalse, v})
    case "nil":
      s.emit(token{t, tokenNil, nil})
    case "range":
      s.emit(token{t, tokenRange, v})
    default:
      s.emit(token{t, tokenIdentifier, v})
  }
  
  return metaAction
}

/***
 ***  SCANNING PRIMITIVES
 ***/

/**
 * Scan a delimited token with escape sequences. The opening delimiter is
 * expected to have already been consumed.
 */
func (s *scanner) scanString(quote, escape rune) (string, error) {
  var unquoted string
  
  for {
    switch r := s.next(); {
      
      case r == eof:
        return "", s.errorf(span{s.text, s.start, s.index - s.start}, nil, "Unexpected end-of-input")
        
      case r == escape:
        if e, err := s.scanEscape(quote, escape); err != nil {
          return "", s.errorf(span{s.text, s.start, s.index - s.start}, err, "Invalid escape sequence")
        }else{
          unquoted += string(e)
        }
        
      case r == quote:
        return unquoted, nil
        
      default:
        unquoted += string(r)
        
    }
  }
  
  return "", s.errorf(span{s.text, s.start, s.index - s.start}, nil, "Unexpected end-of-input")
}

/**
 * Scan an identifier
 */
func (s *scanner) scanIdentifier() (string, error) {
  start := s.index
	for r := s.next(); r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r); {
		r = s.next()
	}
	s.backup() // unget the last character
	return s.text[start:s.index], nil
}

/**
 * Scan a digit value
 */
func digitValue(ch rune) int {
	switch {
    case '0' <= ch && ch <= '9':
      return int(ch - '0')
    case 'a' <= ch && ch <= 'f':
      return int(ch - 'a' + 10)
    case 'A' <= ch && ch <= 'F':
      return int(ch - 'A' + 10)
	}
	return 16 // too big
}

/**
 * Decimla digit?
 */
func isDecimal(ch rune) bool {
  return '0' <= ch && ch <= '9'
}

/**
 * Scan digits
 */
func (s *scanner) scanDigits(base, n int) (string, error) {
  start := s.index
	for r := s.next(); n > 0 && digitValue(r) < base; {
		r = s.next(); n--
	}
	if n > 0 {
		return "", s.errorf(span{s.text, start, s.index - start}, nil, "Not enough digits")
	}else{
	  return s.text[start:s.index-1], nil
	}
}

/**
 * Scan digits
 */
func (s *scanner) scanDecimal(base, n int) (int64, error) {
  if d, err := s.scanDigits(base, n); err != nil {
    return 0, err
  }else{
    return strconv.ParseInt(d, base, 64)
  }
}

/**
 * Scan digits as a rune
 */
func (s *scanner) scanRune(base, n int) (rune, error) {
  if d, err := s.scanDecimal(base, n); err != nil {
    return 0, err
  }else{
    return rune(d), nil
  }
}

/**
 * Scan an escape
 */
func (s *scanner) scanEscape(quote, esc rune) (rune, error) {
  start := s.index
  r := s.next()
	switch r {
    case 'a':
      return '\a', nil
    case 'b':
      return '\b', nil
    case 'f':
      return '\f', nil
    case 'n':
      return '\n', nil
    case 'r':
      return '\r', nil
    case 't':
      return '\t', nil
    case 'v':
      return '\v', nil
    case esc, quote:
      return r, nil
    case '0', '1', '2', '3', '4', '5', '6', '7':
      return s.scanRune(8, 3)
    case 'x':
      return s.scanRune(16, 2)
    case 'u':
      return s.scanRune(16, 4)
    case 'U':
      return s.scanRune(16, 8)
    default:
      return 0, s.errorf(span{s.text, start, s.index - start}, nil, "Invalid escape sequence")
	}
}

/**
 * Scan a number mantissa
 */
func (s *scanner) scanMantissa(ch rune) rune {
	for isDecimal(ch) {
		ch = s.next()
	}
	return ch
}

/**
 * Scan a number fraction
 */
func (s *scanner) scanFraction(ch rune) rune {
	if ch == '.' {
		ch = s.scanMantissa(s.next())
	}
	return ch
}

/**
 * Scan a number exponent
 */
func (s *scanner) scanExponent(ch rune) rune {
	if ch == 'e' || ch == 'E' {
		ch = s.next()
		if ch == '-' || ch == '+' {
			ch = s.next()
		}
		ch = s.scanMantissa(ch)
	}
	return ch
}

/**
 * Scan a number
 */
func (s *scanner) scanNumber() (float64, numericType, error) {
  start := s.index
  ch := s.next()
  
	if ch == '0' {
		// int or float
		ch = s.next()
		if ch == 'x' || ch == 'X' {
		  
			// hexadecimal int
			ch = s.next()
			
			hasMantissa := false
			for digitValue(ch) < 16 {
				ch = s.next()
				hasMantissa = true
			}
			
			s.backup() // unscan the stop rune
			
			if !hasMantissa {
				return 0, 0, s.errorf(span{s.text, start, s.index - start}, nil, "Illegal hexadecimal number")
			}
			
			if v, err := strconv.ParseInt(s.text[start+2:s.index], 16, 64); err != nil {
				return 0, 0, s.errorf(span{s.text, start, s.index - start}, err, "Could not parse number")
			}else{
			  return float64(v), numericInteger, nil
			}
			
		} else {
		  
			// octal int or float
			has8or9 := false
			for isDecimal(ch) {
				if ch > '7' { has8or9 = true }
				ch = s.next()
			}
			
			s.backup() // unscan the stop rune
			
			// octal int
			if has8or9 {
				s.errorf(span{s.text, start, s.index - start}, nil, "Illegal octal number")
			}
			
      t := s.text[start+1:s.index]
      if t == "" { // no more text, this is a zero
        return 0, numericInteger, nil
			}else if v, err := strconv.ParseInt(t, 8, 64); err != nil {
				return 0, 0, s.errorf(span{s.text, start, s.index - start}, err, "Could not parse number")
			}else{
			  return float64(v), numericInteger, nil
			}
			
		}
	}
	
	// decimal int or float
	ch = s.scanMantissa(ch)
	
	// float
	if ch == '.' || ch == 'e' || ch == 'E' {
		// float
		ch = s.scanFraction(ch)
		ch = s.scanExponent(ch)
    // unscan the non-numeric rune
    s.backup()
    if v, err := strconv.ParseFloat(s.text[start:s.index], 64); err != nil {
      return 0, 0, s.errorf(span{s.text, start, s.index - start}, err, "Could not parse number")
    }else{
      return v, numericInteger, nil
    }
	}else{
    // unscan the non-numeric rune
    s.backup()
	}
	
	// integer
  if v, err := strconv.ParseInt(s.text[start:s.index], 10, 64); err != nil {
    return 0, 0, s.errorf(span{s.text, start, s.index - start}, err, "Could not parse number")
  }else{
    return float64(v), numericInteger, nil
  }
  
}

/*
func (s *scanner) scanChar() {
	if s.scanString('\'') != 1 {
		s.error("illegal char literal")
	}
}

func (s *scanner) scanComment(ch rune) rune {
	// ch == '/' || ch == '*'
	if ch == '/' {
		// line comment
		ch = s.next() // read character after "//"
		for ch != '\n' && ch >= 0 {
			ch = s.next()
		}
		return ch
	}

	// general comment
	ch = s.next() // read character after "/*"
	for {
		if ch < 0 {
			s.error("comment not terminated")
			break
		}
		ch0 := ch
		ch = s.next()
		if ch0 == '*' && ch == '/' {
			ch = s.next()
			break
		}
	}
	return ch
}
*/
