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
  _"fmt"
)

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
  exec(*runtime, interface{}) error
}

/**
 * An AST node
 */
type node struct {
  span      span
  token     *token
  subnodes  []executable
}

/**
 * Add a node to this node's subnodes
 */
func (n *node) add(c executable) *node {
  n.subnodes = append(n.subnodes, c)
  return n
}

/**
 * Execute
 */
func (n *node) exec(runtime *runtime, context interface{}) error {
  for _, s := range n.subnodes {
    if err := s.exec(runtime, context); err != nil {
      return err
    }
  }
  return nil
}

/**
 * A program
 */
type program struct {
  node
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
func (n *verbatimNode) exec(runtime *runtime, context interface{}) error {
  if err := n.node.exec(runtime, context); err != nil {
    return err
  }else if _, err := runtime.stdout.Write([]byte(n.span.excerpt())); err != nil {
    return err
  }
  return nil
}

/**
 * A meta node
 */
type metaNode struct {
  node
}

/**
 * An if node
 */
type ifNode struct {
  node
}

/**
 * An expression node
 */
type exprNode struct {
  node
}

