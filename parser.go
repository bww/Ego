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
 * Consume the next token asserting that it is one of the provided token types
 */
func (p *parser) nextAssert(valid ...tokenType) (token, error) {
  t := p.next()
  switch t.which {
    case tokenEOF:
      return token{}, fmt.Errorf("Unexpected end-of-input")
    case tokenError:
      return token{}, fmt.Errorf("Error: %v", t)
  }
  for _, v := range valid {
    if t.which == v {
      return t, nil
    }
  }
  return token{}, invalidTokenError(t, valid...)
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
        if n, err := p.parseMeta(t); err != nil {
          return nil, err
        }else{
          prog.add(n)
        }
        
      default:
        return nil, invalidTokenError(t, tokenVerbatim, tokenMeta, tokenEOF)
        
    }
  }
  
}

/**
 * Parse
 */
func (p *parser) parseMeta(t token) (executable, error) {
  t = p.next()
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
        return n, nil
      }
      
    case tokenFor:
      if n, err := p.parseFor(t); err != nil {
        return nil, err
      }else{
        return n, nil
      }
      
    case tokenBlock:
      if n, err := p.parseBlock(t); err != nil {
        return nil, err
      }else{
        return n, nil
      }
      
    case tokenLParen:
      if n, err := p.parseInterpolate(t); err != nil {
        return nil, err
      }else{
        return n, nil
      }
      
    default:
      return nil, invalidTokenError(t, tokenIf, tokenFor, tokenBlock, '(')
      
  }
}

/**
 * Parse a block containing verbatim and meta content
 */
func (p *parser) parseBlock(t token) (executable, error) {
  b := &containerNode{}
  
  outer: for {
    t := p.next()
    if DEBUG_TRACE_TOKEN {
      fmt.Printf("block:t0: %+v\n", t)
    }
    switch t.which {
      
      case tokenEOF:
        return nil, fmt.Errorf("Unexpected end-of-input")
        
      case tokenError:
        return nil, fmt.Errorf("Error: %v", t)
        
      case tokenClose:
        break outer // close the block
        
      case tokenVerbatim:
        b.add(&verbatimNode{node{t.span, &t}})
        
      case tokenMeta:
        if n, err := p.parseMeta(t); err != nil {
          return nil, err
        }else{
          b.add(n)
        }
        
      default:
        return nil, invalidTokenError(t, tokenVerbatim, tokenMeta)
        
    }
  }
  
  return b, nil
}

/**
 * Parse
 */
func (p *parser) parseIf(t token) (executable, error) {
  
  cond, err := p.parseExpression()
  if err != nil {
    return nil, err
  }
  
  t, err = p.nextAssert(tokenBlock)
  if err != nil {
    return nil, err
  }
  
  iftrue, err := p.parseBlock(t)
  if err != nil {
    return nil, err
  }
  
  var iffalse executable
  t = p.peek(0)
  if t.which == tokenElse {
    p.next() // consume 'else'
    iffalse, err = p.parseMeta(t)
    if err != nil {
      return nil, err
    }
  }
  
  if iffalse != nil {
    return &ifNode{node{encompass(t.span, cond.src(), iftrue.src(), iffalse.src()), &t}, cond, iftrue, iffalse}, nil
  }else{
    return &ifNode{node{encompass(t.span, cond.src(), iftrue.src()), &t}, cond, iftrue, iffalse}, nil
  }
}

/**
 * Parse
 */
func (p *parser) parseFor(t token) (executable, error) {
  
  vars, err := p.parseIdentList()
  if err != nil {
    return nil, err
  }
  
  lspan := []span{t.span}
  for _, e := range vars {
    lspan = append(lspan, e.src())
  }
  
  if len(vars) < 1 || len(vars) > 2 {
    return nil, &parserError{fmt.Sprintf("Incorrect variable count: %d", len(vars)), encompass(lspan...), nil}
  }
  
  t, err = p.nextAssert(tokenAssignSpecial)
  if err != nil {
    return nil, err
  }
  
  t, err = p.nextAssert(tokenRange)
  if err != nil {
    return nil, err
  }
  
  expr, err := p.parseExpression()
  if err != nil {
    return nil, err
  }
  
  t, err = p.nextAssert(tokenBlock)
  if err != nil {
    return nil, err
  }
  
  loop, err := p.parseBlock(t)
  if err != nil {
    return nil, err
  }
  
  lspan = append(lspan, loop.src())
  return &forNode{node{encompass(lspan...), &t}, vars, expr, loop}, nil
}

/**
 * Parse ab expression interpolation
 */
func (p *parser) parseInterpolate(t token) (executable, error) {
  
  expr, err := p.parseParen()
  if err != nil {
    return nil, err
  }
  
  return &exprNode{node{expr.src(), &t}, expr}, nil
}

/**
 * Parse
 */
func (p *parser) parseExpression() (expression, error) {
  return p.parseLogicalNot()
}

/**
 * Parse a logical not
 */
func (p *parser) parseLogicalNot() (expression, error) {
  var t token
  
  op := p.peek(0)
  switch op.which {
    case tokenEOF:
      return nil, fmt.Errorf("Unexpected end-of-input")
    case tokenError:
      return nil, fmt.Errorf("Error: %v", op)
    case tokenBang:
      t = p.next() // consume the '!'
      break // valid token
  }
  
  right, err := p.parseLogicalOr()
  if err != nil {
    return nil, err
  }
  
  if t.which == tokenBang {
    return &logicalNotNode{node{encompass(op.span, right.src()), &op}, right}, nil
  }else{
    return right, nil
  }
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
  
  return &logicalOrNode{node{encompass(op.span, left.src(), right.src()), &op}, left, right}, nil
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
  
  return &logicalAndNode{node{encompass(op.span, left.src(), right.src()), &op}, left, right}, nil
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
  
  return &relationalNode{node{encompass(op.span, left.src(), right.src()), &op}, op, left, right}, nil
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
  
  return &arithmeticNode{node{encompass(op.span, left.src(), right.src()), &op}, op, left, right}, nil
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
  
  return &arithmeticNode{node{encompass(op.span, left.src(), right.src()), &op}, op, left, right}, nil
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
      return &derefNode{node{encompass(op.span, left.src()), &op}, left, v}, nil
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
      return &identNode{node{t.span, &t}, t.value.(string)}, nil
    case tokenNumber, tokenString:
      return &literalNode{node{t.span, &t}, t.value}, nil
    case tokenTrue:
      return &literalNode{node{t.span, &t}, true}, nil
    case tokenFalse:
      return &literalNode{node{t.span, &t}, false}, nil
    case tokenNil:
      return &literalNode{node{t.span, &t}, nil}, nil
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

/**
 * Parse an identifier list
 */
func (p *parser) parseIdentList() ([]expression, error) {
  list := make([]expression, 0)
  
  for {
    var t token
    
    t = p.next()
    if t.which != tokenIdentifier {
      return nil, fmt.Errorf("Expected ident but found %v", t)
    }
    
    list = append(list, &identNode{node{t.span, &t}, t.value.(string)})
    
    t = p.peek(0)
    if t.which != tokenComma {
      break
    }else{
      p.next() // consume the comma
    }
    
  }
  
  return list, nil
}

/**
 * A parser error
 */
type parserError struct {
  message   string
  span      span
  cause     error
}

/**
 * Error
 */
func (e parserError) Error() string {
  if e.cause != nil {
    return fmt.Sprintf("@[%d+%d] %s: %v\n%v", e.span.offset, e.span.length, e.message, e.cause, e.span.excerpt())
  }else{
    return fmt.Sprintf("@[%d+%d] %s\n%v", e.span.offset, e.span.length, e.message, e.span.excerpt())
  }
}

/**
 * Invalid token error
 */
func invalidTokenError(t token, e ...tokenType) error {
  
  m := fmt.Sprintf("Invalid token: %v", t.which)
  if e != nil && len(e) > 0 {
    m += " (expected: "
    for i, t := range e {
      if i > 0 { m += ", " }
      m += fmt.Sprintf("%v", t)
    }
    m += ")"
  }
  
  return &parserError{m, t.span, nil}
}
