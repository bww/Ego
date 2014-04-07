// 
// Copyright (c) 2014 Brian William Wolter, All rights reserved.
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
  "fmt"
)

/**
 * A parser
 */
type parser struct {
  scanner   *scanner
  la        [2]token
}

/**
 * Create a parser
 */
func newParser(s *scanner) *parser {
  return &parser{scanner:s}
}

/**
 * Obtain a look-ahead token without consuming it
 */
func (p *parser) peek(n int) token {
  var t token
  
  if n < len(p.la) {
    return p.la[n]
  }else if n >= cap(p.la) {
    panic("Look-ahead overrun")
  }
  
  for i := len(p.la); i < n; i++ {
    t = p.scanner.scan()
    p.la[i] = t
  }
  
  return t
}

/**
 * Consume the next token
 */
func (p *parser) next() token {
  if len(p.la) < 1 {
    return p.scanner.scan()
  }else{
    t := p.la[0]
    for i := 1; i < len(p.la); i++ { p.la[i-1] = p.la[i] }
    return t
  }
}

/**
 * Parse
 */
func (p *parser) parse() (*program, error) {
  prog := &program{}
  
  for {
    t := p.next()
    switch t.which {
      
      case tokenEOF:
        return prog, nil
        
      case tokenError:
        return nil, fmt.Errorf("Error: %v", t)
        
      case tokenVerbatim:
        prog.add(&verbatimNode{node{t.span, &t, nil}})
        
      case tokenMeta:
        if n, err := p.parseMeta(); err != nil {
          return nil, err
        }else{
          prog.add(n)
        }
        
      default:
        return nil, fmt.Errorf("Unsupported token: %v", t)
        
    }
  }
  
}

/**
 * Parse
 */
func (p *parser) parseMeta() (*node, error) {
  out := &metaNode{}
  t := p.next()
  switch t.which {
    
    case tokenIf:
      if n, err := p.parseIf(); err != nil {
        return nil, err
      }else{
        return out.add(n), nil
      }
      
    case tokenFor:
      if n, err := p.parseFor(); err != nil {
        return nil, err
      }else{
        return out.add(n), nil
      }
      
    default:
      return nil, fmt.Errorf("Illegal token in meta: %v", t)
      
  }
}

/**
 * Parse
 */
func (p *parser) parseIf() (*node, error) {
  out := &ifNode{}
  if n, err := p.parseExpression(); err != nil {
    return nil, err
  }else{
    return out.add(n), nil
  }
}

/**
 * Parse
 */
func (p *parser) parseFor() (*node, error) {
  return nil, nil
}

/**
 * Parse
 */
func (p *parser) parseExpression() (*node, error) {
  out := &exprNode{}
  //t0 := p.peek(0)
  t1 := p.peek(1)
  
  switch t1.which {
    case tokenBlock: // end of meta
    default:
      return nil, fmt.Errorf("Illegal token in expression: %v", t1)
  }
  
  return out.add(nil), nil
}


