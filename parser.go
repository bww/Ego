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
  la        []token
}

/**
 * Create a parser
 */
func newParser(s *scanner) *parser {
  return &parser{s, make([]token, 0, 2)}
}

/**
 * Obtain a look-ahead token without consuming it
 */
func (p *parser) peek(n int) token {
  var t token
  
  if n < len(p.la) {
    return p.la[n]
  }else if n + 1 > cap(p.la) {
    panic("Look-ahead overrun")
  }
  
  l := len(p.la)
  p.la = p.la[:n + 1]
  for i := l; i <= n; i++ {
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
    l := len(p.la)
    for i := 1; i < l; i++ {
      p.la[i-1] = p.la[i]
    }
    p.la = p.la[:l-1]
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
    if DEBUG_TRACE_TOKEN {
      fmt.Printf("parse:t0: %+v\n", t)
    }
    switch t.which {
      
      case tokenEOF:
        return prog, nil
        
      case tokenError:
        return nil, fmt.Errorf("Error: %v", t)
        
      case tokenVerbatim:
        prog.add(&verbatimNode{node{t.span, &t}})
        
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
func (p *parser) parseMeta() (executable, error) {
  t := p.next()
  if DEBUG_TRACE_TOKEN {
    fmt.Printf("meta:t0: %+v\n", t)
  }
  switch t.which {
    
    case tokenEOF:
      return nil, fmt.Errorf("Unexpected end-of-input")
      
    case tokenError:
      return nil, fmt.Errorf("Error: %v", t)
      
    case tokenIf:
      if n, err := p.parseIf(t); err != nil {
        return nil, err
      }else{
        return &metaNode{node{t.span, &t}, n}, nil
      }
      
    case tokenFor:
      if n, err := p.parseFor(t); err != nil {
        return nil, err
      }else{
        return &metaNode{node{t.span, &t}, n}, nil
      }
      
    default:
      return nil, fmt.Errorf("Illegal token in meta: %v", t)
      
  }
}

/**
 * Parse
 */
func (p *parser) parseIf(t token) (executable, error) {
  if n, err := p.parseExpression(); err != nil {
    return nil, err
  }else{
    return &ifNode{node{t.span, &t}, n}, nil
  }
}

/**
 * Parse
 */
func (p *parser) parseFor(t token) (executable, error) {
  return nil, nil
}

/**
 * Parse
 */
func (p *parser) parseExpression() (expression, error) {
  return p.parseLogicalOr()
}

/**
 * Parse a logical or
 */
func (p *parser) parseLogicalOr() (expression, error) {
  
  left, err := p.parseLogicalAnd()
  if err != nil {
    return nil, err
  }
  
  op := p.peek(0)
  switch op.which {
    case tokenEOF:
      return nil, fmt.Errorf("Unexpected end-of-input")
    case tokenError:
      return nil, fmt.Errorf("Error: %v", op)
    case tokenLogicalOr:
      break // valid token
    default:
      return left, nil
  }
  
  p.next() // consume the operator
  right, err := p.parseLogicalOr()
  if err != nil {
    return nil, err
  }
  
  return &logicalOrNode{node{}, left, right}, nil
}

/**
 * Parse a logical and
 */
func (p *parser) parseLogicalAnd() (expression, error) {
  
  left, err := p.parseRelational()
  if err != nil {
    return nil, err
  }
  
  op := p.peek(0)
  switch op.which {
    case tokenEOF:
      return nil, fmt.Errorf("Unexpected end-of-input")
    case tokenError:
      return nil, fmt.Errorf("Error: %v", op)
    case tokenLogicalAnd:
      break // valid token
    default:
      return left, nil
  }
  
  p.next() // consume the operator
  right, err := p.parseLogicalAnd()
  if err != nil {
    return nil, err
  }
  
  return &logicalAndNode{node{}, left, right}, nil
}

/**
 * Parse a relational expression
 */
func (p *parser) parseRelational() (expression, error) {
  
  left, err := p.parseArithmeticL1()
  if err != nil {
    return nil, err
  }
  
  op := p.peek(0)
  switch op.which {
    case tokenEOF:
      return nil, fmt.Errorf("Unexpected end-of-input")
    case tokenError:
      return nil, fmt.Errorf("Error: %v", op)
    case tokenLess, tokenGreater, tokenEqual, tokenLessEqual, tokenGreaterEqual, tokenNotEqual:
      break // valid tokens
    default:
      return left, nil
  }
  
  p.next() // consume the operator
  right, err := p.parseRelational()
  if err != nil {
    return nil, err
  }
  
  return &relationalNode{node{}, op, left, right}, nil
}

/**
 * Parse an arithmetic expression
 */
func (p *parser) parseArithmeticL1() (expression, error) {
  
  left, err := p.parseArithmeticL2()
  if err != nil {
    return nil, err
  }
  
  op := p.peek(0)
  switch op.which {
    case tokenEOF:
      return nil, fmt.Errorf("Unexpected end-of-input")
    case tokenError:
      return nil, fmt.Errorf("Error: %v", op)
    case tokenAdd, tokenSub:
      break // valid tokens
    default:
      return left, nil
  }
  
  p.next() // consume the operator
  right, err := p.parseArithmeticL1()
  if err != nil {
    return nil, err
  }
  
  return &arithmeticNode{node{}, op, left, right}, nil
}

/**
 * Parse an arithmetic expression
 */
func (p *parser) parseArithmeticL2() (expression, error) {
  
  left, err := p.parseDeref()
  if err != nil {
    return nil, err
  }
  
  op := p.peek(0)
  switch op.which {
    case tokenEOF:
      return nil, fmt.Errorf("Unexpected end-of-input")
    case tokenError:
      return nil, fmt.Errorf("Error: %v", op)
    case tokenMul, tokenDiv, tokenMod:
      break // valid tokens
    default:
      return left, nil
  }
  
  p.next() // consume the operator
  right, err := p.parseArithmeticL2()
  if err != nil {
    return nil, err
  }
  
  return &arithmeticNode{node{}, op, left, right}, nil
}

/**
 * Parse a deref expression
 */
func (p *parser) parseDeref() (expression, error) {
  
  left, err := p.parsePrimary()
  if err != nil {
    return nil, err
  }
  
  op := p.peek(0)
  switch op.which {
    case tokenEOF:
      return nil, fmt.Errorf("Unexpected end-of-input")
    case tokenError:
      return nil, fmt.Errorf("Error: %v", op)
    case tokenDot:
      break // valid token
    default:
      return left, nil
  }
  
  p.next() // consume the operator
  right, err := p.parseDeref()
  if err != nil {
    return nil, err
  }
  
  switch v := right.(type) {
    case *identNode, *derefNode:
      return &derefNode{node{}, left, v}, nil
    default:
      return nil, fmt.Errorf("Expected identifier: %v (%T)", right)
  }
  
}

/**
 * Parse a primary expression
 */
func (p *parser) parsePrimary() (expression, error) {
  t := p.next()
  switch t.which {
    case tokenEOF:
      return nil, fmt.Errorf("Unexpected end-of-input")
    case tokenError:
      return nil, fmt.Errorf("Error: %v", t)
    case tokenLParen:
      return p.parseParen()
    case tokenIdentifier:
      return &identNode{node{}, t.value.(string)}, nil
    case tokenNumber, tokenString:
      return &literalNode{node{}, t.value}, nil
    case tokenTrue:
      return &literalNode{node{}, true}, nil
    case tokenFalse:
      return &literalNode{node{}, false}, nil
    case tokenNil:
      return &literalNode{node{}, nil}, nil
    default:
      return nil, fmt.Errorf("Illegal token in primary expression: %v", t)
  }
}

/**
 * Parse a (sub-expression)
 */
func (p *parser) parseParen() (expression, error) {
  
  e, err := p.parseExpression()
  if err != nil {
    return nil, err
  }
  
  t := p.next()
  if t.which != tokenRParen {
    return nil, fmt.Errorf("Expected ')' but found %v", t)
  }
  
  return e, nil
}

