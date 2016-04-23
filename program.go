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

var (
  errBreak    = fmt.Errorf("break")
  errContinue = fmt.Errorf("continue")
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
func (c *context) get(s span, n string) (interface{}, error) {
  if l := len(c.stack); l > 0 {
    return derefProp(s, c.stack[l-1], n)
  }else{
    return nil, fmt.Errorf("No context")
  }
}

/**
 * Executable context
 */
type Runtime struct {
  Stdout    io.Writer
}

/**
 * Executable
 */
type executable interface {
  src()(span)
  exec(*Runtime, *context) error
}

/**
 * An expression
 */
type expression interface {
  src()(span)
  exec(*Runtime, *context)(interface{}, error)
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
type Program struct {
  containerNode
}

/**
 * Execute a program
 */
func (n *Program) Exec(runtime *Runtime, context interface{}) error {
  return n.exec(runtime, newContext(context))
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
func (n *containerNode) exec(runtime *Runtime, context *context) error {
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
func (n *verbatimNode) exec(runtime *Runtime, context *context) error {
  _, err := runtime.Stdout.Write([]byte(n.span.excerpt()))
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
func (n *metaNode) exec(runtime *Runtime, context *context) error {
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
func (n *ifNode) exec(runtime *Runtime, context *context) error {
  
  res, err := n.condition.exec(runtime, context)
  if err != nil {
    return err
  }
  
  var istrue bool
  switch v := res.(type) {
    case bool:
      istrue = v
    default:
      return runtimeErrorf(n.span, "Invalid type: %T", v)
  }
  
  if istrue {
    return n.iftrue.exec(runtime, context)
  }else if n.iffalse != nil {
    return n.iffalse.exec(runtime, context)
  }
  
  return nil
}

/**
 * A for node
 */
type forNode struct {
  node
  vars  []expression
  expr  expression
  loop  executable
}

/**
 * Execute
 */
func (n *forNode) exec(runtime *Runtime, context *context) error {
  
  items, err := n.expr.exec(runtime, context)
  if err != nil {
    return err
  }
  
  value := reflect.ValueOf(items)
  deref, _ := derefValue(value)
  switch deref.Kind() {
    case reflect.Array:
      return n.execArray(runtime, context, deref)
    case reflect.Slice:
      return n.execArray(runtime, context, deref)
    case reflect.Map:
      return n.execMap(runtime, context, deref)
    default:
      return runtimeErrorf(n.span, "Expression result is not iterable: %v", deref.Type())
  }
  
}

/**
 * Execute
 */
func (n *forNode) execArray(runtime *Runtime, context *context, val reflect.Value) error {
  frame := make(map[string]interface{})
  context.push(frame)
  defer context.pop()
  
  l := val.Len()
  for i := 0; i < l; i++ {
    v := val.Index(i)
    
    if len(n.vars) == 1 {
      frame[n.vars[0].(*identNode).ident] = v.Interface()
    }else{
      frame[n.vars[0].(*identNode).ident] = i
      frame[n.vars[1].(*identNode).ident] = v.Interface()
    }
    
    err := n.loop.exec(runtime, context)
    if err == errBreak {
      break
    }else if err == errContinue {
      continue
    }else if err != nil {
      return err
    }
  }
  
  return nil
}

/**
 * Execute
 */
func (n *forNode) execMap(runtime *Runtime, context *context, val reflect.Value) error {
  frame := make(map[string]interface{})
  context.push(frame)
  defer context.pop()
  
  keys := val.MapKeys()
  for _, k := range keys {
    v := val.MapIndex(k)
    
    if len(n.vars) == 1 {
      frame[n.vars[0].(*identNode).ident] = v.Interface()
    }else{
      frame[n.vars[0].(*identNode).ident] = k.Interface()
      frame[n.vars[1].(*identNode).ident] = v.Interface()
    }
    
    err := n.loop.exec(runtime, context)
    if err == errBreak {
      break
    }else if err == errContinue {
      continue
    }else if err != nil {
      return err
    }
  }
  
  return nil
}

/**
 * An expression node
 */
type exprNode struct {
  node
  expr  expression
}

/**
 * Execute
 */
func (n *exprNode) exec(runtime *Runtime, context *context) error {
  
  res, err := n.expr.exec(runtime, context)
  if err != nil {
    return err
  }
  
  var out string
  switch v := res.(type) {
    case string:
      out = v
    case []byte:
      out = string(v)
    default:
      out = fmt.Sprintf("%v", v)
  }
  
  _, err = runtime.Stdout.Write([]byte(out))
  if err != nil {
    return err
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
func (n *logicalNotNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  
  rvi, err := n.right.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  
  rv, err := asBool(n.right.src(), rvi)
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
func (n *logicalOrNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  
  lvi, err := n.left.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  lv, err := asBool(n.left.src(), lvi)
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
  rv, err := asBool(n.right.src(), rvi)
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
func (n *logicalAndNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  
  lvi, err := n.left.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  lv, err := asBool(n.left.src(), lvi)
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
  rv, err := asBool(n.right.src(), rvi)
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
func (n *arithmeticNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  
  lvi, err := n.left.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  lv, err := asNumber(n.left.src(), lvi)
  if err != nil {
    return nil, err
  }
  
  rvi, err := n.right.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  rv, err := asNumber(n.right.src(), rvi)
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
      return nil, runtimeErrorf(n.span, "Invalid operator: %v", n.op)
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
func (n *relationalNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  
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
  
  lv, err := asNumber(n.left.src(), lvi)
  if err != nil {
    return nil, err
  }
  rv, err := asNumber(n.right.src(), rvi)
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
      return nil, runtimeErrorf(n.span, "Invalid operator: %v", n.op)
  }
  
}

/**
 * An index (subscript) expression node
 */
type indexNode struct {
  node
  left, right expression
}

/**
 * Execute
 */
func (n *indexNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  
  val, err := n.left.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  
  sub, err := n.right.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  
  deref, _ := derefValue(reflect.ValueOf(val))
  switch deref.Kind() {
    case reflect.Array:
      return n.execArray(runtime, context, deref, reflect.ValueOf(sub))
    case reflect.Slice:
      return n.execArray(runtime, context, deref, reflect.ValueOf(sub))
    case reflect.Map:
      return n.execMap(runtime, context, deref, reflect.ValueOf(sub))
    default:
      return nil, runtimeErrorf(n.span, "Expression result is not indexable: %v", deref.Type())
  }
  
}

/**
 * Execute
 */
func (n *indexNode) execArray(runtime *Runtime, context *context, val reflect.Value, index reflect.Value) (interface{}, error) {
  
  i, err := asNumberValue(n.right.src(), index)
  if err != nil {
    return nil, err
  }
  
  l := val.Len()
  if int(i) < 0 || int(i) >= l {
    return nil, runtimeErrorf(n.span, "Index out-of-bounds: %v", i)
  }
  
  return val.Index(int(i)).Interface(), nil
}

/**
 * Execute
 */
func (n *indexNode) execMap(runtime *Runtime, context *context, val reflect.Value, key reflect.Value) (interface{}, error) {
  
  if !key.Type().AssignableTo(val.Type().Key()) {
    return nil, runtimeErrorf(n.span, "Expression result is not assignable to map key type: %v != %v", key.Type(), val.Type().Key())
  }
  
  return val.MapIndex(key).Interface(), nil
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
func (n *derefNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  
  v, err := n.left.exec(runtime, context)
  if err != nil {
    return nil, err
  }
  
  context.push(v)
  defer context.pop()
  
  var z interface{}
  switch v := n.right.(type) {
    case *identNode:
      z, err = context.get(n.span, v.ident)
    case *derefNode:
      z, err = v.exec(runtime, context)
    default:
      return nil, runtimeErrorf(n.span, "Invalid right operand to . (dereference): %T", v)
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
func (n *identNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  return context.get(n.span, n.ident)
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
func (n *literalNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  return n.value, nil
}

/**
 * A break expression node
 */
type breakNode struct {
  node
}

/**
 * Execute
 */
func (n *breakNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  return nil, errBreak
}

/**
 * A continue expression node
 */
type continueNode struct {
  node
}

/**
 * Execute
 */
func (n *continueNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  return nil, errContinue
}

/**
 * Obtain an interface value as a bool
 */
func asBool(s span, value interface{}) (bool, error) {
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
      return false, runtimeErrorf(s, "Cannot cast %v to bool", v.Type())
  }
}

/**
 * Obtain an interface value as a number
 */
func asNumber(s span, v interface{}) (float64, error) {
  return asNumberValue(s, reflect.ValueOf(v))
}

/**
 * Obtain an interface value as a number
 */
func asNumberValue(s span, v reflect.Value) (float64, error) {
  switch v.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
      return float64(v.Int()), nil
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
      return float64(v.Uint()), nil
    case reflect.Float32, reflect.Float64:
      return v.Float(), nil
    default:
      return 0, runtimeErrorf(s, "Cannot cast %v to numeric", v.Type())
  }
}

/**
 * Dereference
 */
func derefProp(s span, context interface{}, ident string) (interface{}, error) {
  
  switch v := context.(type) {
    case Context:
      return v.Variable(ident)
    case VariableProvider:
      return v(ident)
    case func(string)(interface{}, error):
      return v(ident)
    case map[string]interface{}:
      return v[ident], nil
  }
  
  val, _ := derefValue(reflect.ValueOf(context))
  switch val.Kind() {
    case reflect.Map:
      return derefMap(s, val, ident)
    case reflect.Struct:
      return derefMember(s, val, ident)
    default:
      return nil, runtimeErrorf(s, "Cannot dereference variable: %v", val.Type())
  }
  
}

/**
 * Execute
 */
func derefMap(s span, val reflect.Value, property string) (interface{}, error) {
  key := reflect.ValueOf(property)
  
  if !key.Type().AssignableTo(val.Type().Key()) {
    return nil, runtimeErrorf(s, "Expression result is not assignable to map key type: %v != %v", key.Type(), val.Type().Key())
  }
  
  return val.MapIndex(key).Interface(), nil
}

/**
 * Execute
 */
func derefMember(s span, val reflect.Value, property string) (interface{}, error) {
  var v reflect.Value
  
  if val.Kind() != reflect.Struct {
    return nil, runtimeErrorf(s, "Cannot dereference variable: %v", val.Type())
  }
  
  v = val.MethodByName(property)
  if v.IsValid() {
    r := v.Call(make([]reflect.Value,0))
    if r == nil {
      return nil, runtimeErrorf(s, "Method %v of %v did not return a value", v, val.Type())
    }else if l := len(r); l < 1 || l > 2 {
      return nil, runtimeErrorf(s, "Method %v of %v must return either (interface{}) or (interface{}, error)", v, val.Type())
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
          return nil, runtimeErrorf(s, "Method %v of %v must return either (interface{}) or (interface{}, error)", v, val.Type())
      }
    }
  }
  
  v = val.FieldByName(property)
  if v.IsValid() {
    return v.Interface(), nil
  }
  
  return nil, runtimeErrorf(s, "No suitable method or field '%v' of %v", property, val.Type())
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
 * Format a runtime error
 */
func runtimeErrorf(s span, f string, a ...interface{}) *runtimeError {
  return &runtimeError{fmt.Sprintf(f, a...), s, nil}
}

/**
 * Error
 */
func (e runtimeError) Error() string {
  if e.cause != nil {
    return fmt.Sprintf("%s: %v\n%v", e.message, e.cause, excerptCallout.FormatExcerpt(e.span))
  }else{
    return fmt.Sprintf("%s\n%v", e.message, excerptCallout.FormatExcerpt(e.span))
  }
}

