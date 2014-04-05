// 
// Copyright (c) 2014 Brian William Wolter, All rights reserved.
// Go Framer
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
//   * Neither the names of Brian William Wolter, Wolter Group New York, nor the
//     names of its contributors may be used to endorse or promote products derived
//     from this software without specific prior written permission.
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

package ego

import (
  _ "fmt"
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
 * Token type
 */
type tokenType int

/**
 * Token types
 */
const (
  tokenError tokenType = iota
  tokenEOF
  tokenVerbatim
  tokenMeta
)

/**
 * Token type string
 */
func (t tokenType) String() string {
  switch t {
    case tokenError:
      return "Error"
    case tokenEOF:
      return "EOF"
    case tokenVerbatim:
      return "Verbatim"
    case tokenMeta:
      return "@"
    default:
      return "Unknown"
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
 * A scanner
 */
type scanner struct {
  text      string
  index     int
  width     int // current rune width
  start     int // token start position
  tokens    chan token
}

/**
 * A scanner action
 */
type scannerAction func(*scanner) scannerAction

/**
 * Create a scanner
 */
func newScanner(text string, tokens chan token) *scanner {
  return &scanner{text, 0, 0, 0, tokens}
}

/**
 * Scan tokens and produce them on our token channel
 */
func (s *scanner) scan() {
  for state := startAction; state != nil; {
    state = state(s)
  }
  close(s.tokens)
}

/**
 * Emit a token
 */
func (s *scanner) emit(t token) {
  s.tokens <- t
  s.start = t.span.offset + t.span.length
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
 * Unconsume the previous rune from input (this can be called only once
 * per invocation of next())
 */
func (s *scanner) backup() {
  s.index -= s.width
}

/**
 * Start action
 */
func startAction(s *scanner) scannerAction {
  
  for {
    if s.index < len(s.text) && s.text[s.index] == meta {
      if s.index > s.start {
        s.emit(token{span{s.text, s.start, s.index - s.start}, tokenVerbatim, s.text[s.start:s.index]})
      }
      return metaAction  // next state
    }
    if s.next() == eof {
      break
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
 * Meta action
 */
func metaAction(s *scanner) scannerAction {
  s.emit(token{span{s.text, s.index, 1}, tokenMeta, "@"})
  s.next() // skip the '@' delimiter
  return startAction
}

