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

package ego

import (
  "io"
  "os"
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
  return &context{[]interface{}{stdlib, f}}
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
  return c.sget(s, n, c.stack)
}

/**
 * Obtain a value
 */
func (c *context) sget(s span, n string, k []interface{}) (interface{}, error) {
  l := len(k)
  if l < 1 {
    return nil, nil
  }
  
  v, err := derefProp(s, k[l-1], n)
  if err != nil {
    return nil, err
  }
  
  if v == nil && l > 1 {
    return c.sget(s, n, k[:l-1])
  }else{
    return v, nil
  }
}

/**
 * Executable context
 */
type Runtime struct {
  Stdout    io.Writer
  attrs     map[string]interface{}
}

/**
 * Get an attribute.
 */
func (r *Runtime) Attr(k string) interface{} {
  if r.attrs != nil {
    return r.attrs[k]
  }else{
    return nil
  }
}

/**
 * Set an attribute.
 */
func (r *Runtime) SetAttr(k string, v interface{}) {
  if r.attrs == nil {
    r.attrs = make(map[string]interface{})
  }
  r.attrs[k] = v
}

/**
 * Execution state
 */
type State struct {
  Runtime   *Runtime
  Context   interface{}
}

var typeOfState   = reflect.TypeOf(&State{})
var typeOfRuntime = reflect.TypeOf(&Runtime{})
var typeOfError   = reflect.TypeOf((*error)(nil)).Elem()

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
func (n *Program) Exec(rt *Runtime, cxt interface{}) error {
  if rt.Stdout == nil {
    rt.Stdout = os.Stdout
  }
  switch v := cxt.(type) {
    case *context:
      return n.exec(rt, v)
    default:
      return n.exec(rt, newContext(v))
  }
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
      return runtimeErrorf(n.expr.src(), "Expression result is not iterable: %v", displayType(deref))
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
  if res == nil {
    return nil // no output on nil
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
    case *derefNode, *indexNode, *invokeNode:
      z, err = n.right.exec(runtime, context)
    default:
      return nil, runtimeErrorf(n.span, "Invalid right operand to . (dereference): %T", v)
  }
  
  return z, err
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
  
  prop := reflect.ValueOf(sub)
  if prop.Kind() == reflect.Invalid {
    return nil, runtimeErrorf(n.right.src(), "Subscript expression is nil")
  }
  
  context.push(val)
  defer context.pop()
  
  deref, _ := derefValue(reflect.ValueOf(val))
  switch deref.Kind() {
    case reflect.Array:
      return n.execArray(runtime, context, deref, prop)
    case reflect.Slice:
      return n.execArray(runtime, context, deref, prop)
    case reflect.Map:
      return n.execMap(runtime, context, deref, prop)
    default:
      return nil, runtimeErrorf(n.span, "Expression result is not indexable: %v", displayType(deref))
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
 * A function invocation expression node
 */
type invokeNode struct {
  node
  left, right expression
  params      []expression
}

/**
 * Execute
 */
func (n *invokeNode) exec(runtime *Runtime, context *context) (interface{}, error) {
  var liv interface{}
  var err error
  
  var name string
  switch v := n.right.(type) {
    case *identNode:
      name = v.ident
    default:
      return nil, runtimeErrorf(n.span, "Invalid node type for function call: %T", v)
  }
  
  if n.left != nil {
    liv, err = n.left.exec(runtime, context)
    if err != nil {
      return nil, err
    }
  }
  
  var f reflect.Value
  if liv != nil {
    lrv := reflect.ValueOf(liv)
    f = lrv.MethodByName(name)
    if !f.IsValid() {
      return nil, runtimeErrorf(n.span, "No such method '%v' for type %v or method is not exported", name, lrv.Type())
    }
  }else{
    liv, err = context.get(n.span, name)
    if err != nil {
      return nil, err
    }else if liv == nil {
      return nil, runtimeErrorf(n.span, "No such function '%v'", name)
    }
    f = reflect.ValueOf(liv)
    if f.Kind() != reflect.Func {
      return nil, runtimeErrorf(n.span, "Variable '%v' is not a function", name)
    }
  }
  
  ft := f.Type()
  lp := len(n.params)
  
  cout := ft.NumOut()
  if cout > 2 {
    return nil, runtimeErrorf(n.span, "Function %v returns %v values (expected: 0, 1 or 2)", name, cout)
  }
  
  in, extra := 0, 1
  args := make([]reflect.Value, 0)
  cin := ft.NumIn()
  if cin != lp {
    if cin - extra != lp /* allow for runtime parameter */ {
      return nil, runtimeErrorf(n.span, "Function %v takes %v arguments but is given %v", name, cin, lp)
    }
    if ft.In(in) != typeOfState {
      return nil, runtimeErrorf(n.span, "Function %v takes %v arguments but is given %v; first native argument must receive %v", name, cin - extra, lp, typeOfState)
    }
    args = append(args, reflect.ValueOf(&State{runtime, context}))
    in++
  }
  
  for _, e := range n.params {
    v, err := e.exec(runtime, context)
    if err != nil {
      return nil, err
    }
    t := ft.In(in)
    var a reflect.Value
    if v == nil { // we need a typed zero value if the value is nil
      a = reflect.Zero(t)
    }else{
      a = reflect.ValueOf(v)
    }
    if !a.IsValid() {
      return nil, runtimeErrorf(e.src(), "Invalid parameter")
    }
    if !a.Type().AssignableTo(t) {
      return nil, runtimeErrorf(e.src(), "Cannot use %v as %v", displayType(a), t.String())
    }
    args = append(args, a)
    in++
  }
  
  r := f.Call(args)
  if r == nil {
    return nil, runtimeErrorf(n.span, "Function %v did not return a value", name)
  }else if l := len(r); l > 2 {
    return nil, runtimeErrorf(n.span, "Function %v must return either (void), (interface{}) or (interface{}, error)", name)
  }else if l == 0 {
    return nil, nil
  }else if l == 1 {
    if ft.Out(0) == typeOfError {
      if !r[0].IsNil() {
        return nil, r[0].Interface().(error)
      }else{
        return nil, nil
      }
    }else{
      return r[0].Interface(), nil
    }
  }else if l == 2 {
    r0 := r[0].Interface()
    r1 := r[1].Interface()
    if r1 == nil {
      return r0, nil
    }else if e, ok := r1.(error); ok {
      return r0, e
    }else{
      return nil, runtimeErrorf(n.span, "Function %v must return either (void), (interface{}) or (interface{}, error)", name)
    }
  }
  
  return nil, nil
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
      return false, runtimeErrorf(s, "Cannot cast %v to bool", displayType(v))
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
      return 0, runtimeErrorf(s, "Cannot cast %v to numeric", displayType(v))
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
      return nil, runtimeErrorf(s, "Cannot dereference variable: %v", displayType(val))
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
  
  res := val.MapIndex(key)
  if !res.IsValid() {
    return nil, nil // not found
  }
  
  return res.Interface(), nil
}

/**
 * Execute
 */
func derefMember(s span, val reflect.Value, property string) (interface{}, error) {
  var v reflect.Value
  
  if val.Kind() != reflect.Struct {
    return nil, runtimeErrorf(s, "Cannot dereference variable: %v", displayType(val))
  }
  
  v = val.MethodByName(property)
  if v.IsValid() {
    t := v.Type()
    
    if n := t.NumIn(); n != 0 {
      return nil, runtimeErrorf(s, "Method %v of %v takes %v arguments (expected: 0)", v, displayType(val), n)
    }
    if n := t.NumOut(); n < 1 || n > 2 {
      return nil, runtimeErrorf(s, "Method %v of %v returns %v values (expected: 1 or 2)", v, displayType(val), n)
    }
    
    r := v.Call(make([]reflect.Value,0))
    if r == nil {
      return nil, runtimeErrorf(s, "Method %v of %v did not return a value", v, displayType(val))
    }else if l := len(r); l < 1 || l > 2 {
      return nil, runtimeErrorf(s, "Method %v of %v must return either (interface{}) or (interface{}, error)", v, displayType(val))
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
          return nil, runtimeErrorf(s, "Method %v of %v must return either (interface{}) or (interface{}, error)", v, displayType(val))
      }
    }
    
  }
  
  v = val.FieldByName(property)
  if v.IsValid() {
    if !v.CanInterface() {
      return nil, runtimeErrorf(s, "Cannot access %v of %v", property, displayType(val))
    }
    return v.Interface(), nil
  }
  
  return nil, nil
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
 * Obtain the presentation type of a value
 */
func displayType(v reflect.Value) string {
  if v.Kind() == reflect.Invalid {
    return "<nil>"
  }else{
    return v.Type().String()
  }
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

