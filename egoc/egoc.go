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

package main

import (
  "os"
  "fmt"
  "flag"
  "path"
  "ego"
)

var CMD string

func main() {
  CMD = path.Base(os.Args[0])
  
  cmdline         := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  fContext        := cmdline.String   ("context",   "",       "Path to the context data to be used when evaluating a source file. This file must be formatted as JSON.")
  fVerbose        := cmdline.Bool     ("verbose",   false,    "Be verbose.")
  fDebug          := cmdline.Bool     ("debug",     false,    "Debug the compiler and runtime (be very verbose).")
  cmdline.Parse(os.Args[1:])
  
  if *fVerbose { }
  if *fDebug {
    ego.DEBUG_TRACE_TOKEN = true
  }
  
  if *fContext == "" {
    fmt.Fprintf(os.Stderr, "%v: No context provided\n", CMD)
    return
  }
  
  context, err := readContext(*fContext)
  if err != nil {
    fmt.Fprintf(os.Stderr, "%v: Could not load context: %v\n", CMD, err)
    return
  }
  
  runtime := &ego.Runtime{
    Stdout: os.Stdout,
  }
  
  for _, p := range cmdline.Args() {
    
    src, err := readFile(p)
    if err != nil {
      fmt.Fprintf(os.Stderr, "%v: Could not read source: %v: %v\n", CMD, p, err)
      return
    }
    
    prog, err := ego.Compile(string(src))
    if err != nil {
      fmt.Fprintf(os.Stderr, "%v: Could not compile: %v: %v\n", CMD, p, err)
      return
    }
    
    err = prog.Exec(runtime, context)
    if err != nil {
      fmt.Fprintf(os.Stderr, "\n%v: Could not execute: %v: %v\n", CMD, p, err)
      return
    }
    
  }
  
}
