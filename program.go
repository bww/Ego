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
  "io"
  "fmt"
  "reflect"
)

/**
 * A runtime context
 */
type Context interface {
  Variable(name string)(interface{}, error)
}

/**
 * A variable provider
 */
type VariableProvider func(name string)(interface{}, error)

/**
 * Executable context
 */
type context struct {
  stack   []interface{}
}

/**
 * Create a new context
 */
func newContext(f interface{}) *context {
  return &context{[]interface{}{f}}
}

/**
 * Push a frame
 */
func (c *context) push(f interface{}) *context {
  if l := len(c.stack); cap(c.stack) > l {
    c.stack = c.stack[:l+1]
    c.stack[l] = f
  }else{
    c.stack = append(c.stack, f)
  }
  return c
}

/**
 * Pop a frame
 */
func (c *context) pop() *context {
  if l := len(c.stack); l > 0 {
    c.stack = c.stack[:l-1]
  }
  return c
}

/**
 * Obtain a value
 */
func (c *context) get(n string) (interface{}, error) {
  if l := len(c.stack); l > 0 {
    return derefProp(c.stack[l-1], n)
  }else{
    return nil, fmt.Errorf("No context")
  }
}

/**
 * Executable context
 */
type runtime struct {
  stdout    io.Writer
}

/**
 * Executable
 */
type executable interface {
  src()(span)
  exec(*runtime, *context) error
}

/**
 * An expression
 */
type expression interface {
  src()(span)
  exec(*runtime, *context)(interface{}, error)
}

/**
 * An AST node
 */
type node struct {
  span      span
  token     *token
}

/**
 * Obtain the node src span
 */
func (n node) src() span {
  return n.span
}

/**
 * A program
 */
type program struct {
  containerNode
}

/**
 * A container node
 */
type containerNode struct {
  node
  subnodes []executable
}

/**
 * Append a subnode
 */
func (n *containerNode) add(s executable) {
  if n.subnodes == nil {
    n.subnodes = make([]executable, 0)
  }
  n.subnodes = append(n.subnodes, s)
}

/**
 * Execute
 */
func (n *containerNode) exec(runtime *runtime, context *context) error {
  if n.subnodes == nil {
    return nil // nothing to do
  }
  for _, e := range n.subnodes {
    err := e.exec(runtime, context)
    if err != nil {
      return err
    }
  }
  return nil
}

/**
 * A verbatim node
 */
type verbatimNode struct {
  node
}

/**
 * Execute
 */
func (n *verbatimNode) exec(runtime *runtime, context *context) error {
  _, err := runtime.stdout.Write([]byte(n.span.excerpt()))
  if err != nil {
    return err
  }
  return nil
}

/**
 * A meta node
 */
type metaNode struct {
  node
  child executable
}

/**
 * Execute
 */
func (n *metaNode) exec(runtime *runtime, context *context) error {
  fmt.Println("exec:meta")
  return n.child.exec(runtime, context)
}

/**
 * An if node
 */
type ifNode struct {
  node
  condition expression
  iftrue    executable
  iffalse   executable
}

/**
 * Execute
 */
func (n *ifNode) exec(runtime *runtime, context *context) error {
  
  res, err := n.condition.exec(runtime, context)
  if err != nil {
    return err
  }
  
  var istrue bool
  switch v := res.(type) {
    case bool:
      istrue = v
    default:
      return fmt.Errorf("Invalid type: %T", v)
  }
  
  if istrue {
    return n.iftrue.exec(runtime, context)
  }else if n.iffalse != nil {
    return n.iffalse.exec(runtime, context)
  }
  
  return nil
}

/**
 * A logical NOT node
 */
type logicalNotNode struct {
  node
  right expression
}

/**
 * Execute
 */
func (n *logicalNotNode) exec(runtime *runtime, context *context) (interface{}, error) {
  
  rvi, err := n.right.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  
  rv, err := asBool(rvi)
  if err != nil {
    return nil, err
  }
  
  return !rv, nil
}

/**
 * A logical OR node
 */
type logicalOrNode struct {
  node
  left, right expression
}

/**
 * Execute
 */
func (n *logicalOrNode) exec(runtime *runtime, context *context) (interface{}, error) {
  
  lvi, err := n.left.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  lv, err := asBool(lvi)
  if err != nil {
    return nil, err
  }
  
  if lv {
    return true, nil
  }
  
  rvi, err := n.right.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  rv, err := asBool(rvi)
  if err != nil {
    return nil, err
  }
  
  return rv, nil
}

/**
 * A logical AND node
 */
type logicalAndNode struct {
  node
  left, right expression
}

/**
 * Execute
 */
func (n *logicalAndNode) exec(runtime *runtime, context *context) (interface{}, error) {
  
  lvi, err := n.left.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  lv, err := asBool(lvi)
  if err != nil {
    return nil, err
  }
  
  if !lv {
    return false, nil
  }
  
  rvi, err := n.right.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  rv, err := asBool(rvi)
  if err != nil {
    return nil, err
  }
  
  return rv, nil
}

/**
 * An arithmetic expression node
 */
type arithmeticNode struct {
  node
  op          token
  left, right expression
}

/**
 * Execute
 */
func (n *arithmeticNode) exec(runtime *runtime, context *context) (interface{}, error) {
  
  lvi, err := n.left.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  lv, err := asNumber(lvi)
  if err != nil {
    return nil, err
  }
  
  rvi, err := n.right.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  rv, err := asNumber(rvi)
  if err != nil {
    return nil, err
  }
  
  switch n.op.which {
    case tokenAdd:
      return lv + rv, nil
    case tokenSub:
      return lv - rv, nil
    case tokenMul:
      return lv * rv, nil
    case tokenDiv:
      return lv / rv, nil
    case tokenMod: // truncates to int
      return int64(lv) % int64(rv), nil
    default:
      return nil, fmt.Errorf("Invalid operator: %v", n.op)
  }
  
}

/**
 * An relational expression node
 */
type relationalNode struct {
  node
  op          token
  left, right expression
}

/**
 * Execute
 */
func (n *relationalNode) exec(runtime *runtime, context *context) (interface{}, error) {
  
  lvi, err := n.left.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  rvi, err := n.right.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  
  switch n.op.which {
    case tokenEqual:
      return lvi == rvi, nil
    case tokenNotEqual:
      return lvi != rvi, nil
  }
  
  lv, err := asNumber(lvi)
  if err != nil {
    return nil, err
  }
  rv, err := asNumber(rvi)
  if err != nil {
    return nil, err
  }
  
  switch n.op.which {
    case tokenLess:
      return lv < rv, nil
    case tokenGreater:
      return lv > rv, nil
    case tokenLessEqual:
      return lv <= rv, nil
    case tokenGreaterEqual:
      return lv >= rv, nil
    default:
      return nil, fmt.Errorf("Invalid operator: %v", n.op)
  }
  
}

/**
 * A dereference expression node
 */
type derefNode struct {
  node
  left, right expression
}

/**
 * Execute
 */
func (n *derefNode) exec(runtime *runtime, context *context) (interface{}, error) {
  
  v, err := n.left.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  
  context.push(v)
  defer context.pop()
  
  var z interface{}
  switch v := n.right.(type) {
    case *identNode:
      z, err = context.get(v.ident)
    case *derefNode:
      z, err = v.exec(runtime, context)
    default:
      return nil, fmt.Errorf("Invalid right operand to . (dereference): %v (%T)", v, v)
  }
  
  return z, err
}

/**
 * An identifier expression node
 */
type identNode struct {
  node
  ident string
}

/**
 * Execute
 */
func (n *identNode) exec(runtime *runtime, context *context) (interface{}, error) {
  return context.get(n.ident)
}

/**
 * A literal expression node
 */
type literalNode struct {
  node
  value interface{}
}

/**
 * Execute
 */
func (n *literalNode) exec(runtime *runtime, context *context) (interface{}, error) {
  return n.value, nil
}

/**
 * Obtain an interface value as a bool
 */
func asBool(value interface{}) (bool, error) {
  v := reflect.ValueOf(value)
  switch v.Kind() {
    case reflect.Bool:
      return v.Bool(), nil
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
      return v.Int() != 0, nil
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
      return v.Uint() != 0, nil
    case reflect.Float32, reflect.Float64:
      return v.Float() != 0, nil
    default:
      return false, fmt.Errorf("Cannot cast %v (%T) to bool", value, value)
  }
}

/**
 * Obtain an interface value as a number
 */
func asNumber(value interface{}) (float64, error) {
  v := reflect.ValueOf(value)
  switch v.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
      return float64(v.Int()), nil
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
      return float64(v.Uint()), nil
    case reflect.Float32, reflect.Float64:
      return v.Float(), nil
    default:
      return 0, fmt.Errorf("Cannot cast %v (%T) to numeric", value, value)
  }
}

/**
 * Dereference
 */
func derefProp(context interface{}, ident string) (interface{}, error) {
  switch v := context.(type) {
    case Context:
      return v.Variable(ident)
    case VariableProvider:
      return v(ident)
    case func(string)(interface{}, error):
      return v(ident)
    case map[string]interface{}:
      return v[ident], nil
    default:
      return derefMember(context, ident)
  }
}

/**
 * Execute
 */
func derefMember(context interface{}, property string) (interface{}, error) {
  var v reflect.Value
  
  value := reflect.ValueOf(context)
  deref, _ := derefValue(value)
  if deref.Kind() != reflect.Struct {
    return nil, fmt.Errorf("Cannot dereference context: %v (%T)", context, context)
  }
  
  v = value.MethodByName(property)
  if v.IsValid() {
    r := v.Call(make([]reflect.Value,0))
    if r == nil {
      return nil, fmt.Errorf("Method %v of %v (%T) did not return a value", v, value, value)
    }else if l := len(r); l < 1 || l > 2 {
      return nil, fmt.Errorf("Method %v of %v (%T) must return either (interface{}) or (interface{}, error)", v, value, value)
    }else if l == 1 {
      return r[0].Interface(), nil
    }else if l == 2 {
      r0 := r[0].Interface()
      r1 := r[1].Interface()
      if r1 == nil {
        return r0, nil
      }
      switch e := r1.(type) {
        case error:
          return r0, e
        default:
          return nil, fmt.Errorf("Method %v of %v (%T) must return either (interface{}) or (interface{}, error)", v, value, value)
      }
    }
  }
  
  v = deref.FieldByName(property)
  if v.IsValid() {
    return v.Interface(), nil
  }
  
  return nil, fmt.Errorf("No suitable method or field '%v' of %v (%T)", property, value.Interface(), value.Interface())
}

/**
 * Dereference a value
 */
func derefValue(value reflect.Value) (reflect.Value, int) {
  v := value
  c := 0
  for ; v.Kind() == reflect.Ptr; {
    v = v.Elem()
    c++
  }
  return v, c
}

/**
 * A runtime error
 */
type runtimeError struct {
  message   string
  span      span
  cause     error
}

/**
 * Error
 */
func (e runtimeError) Error() string {
  if e.cause != nil {
    return fmt.Sprintf("@[%d+%d] %s: %v\n%v", e.span.offset, e.span.length, e.message, e.cause, e.span.excerpt())
  }else{
    return fmt.Sprintf("@[%d+%d] %s\n%v", e.span.offset, e.span.length, e.message, e.span.excerpt())
  }
}

