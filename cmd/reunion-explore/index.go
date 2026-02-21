package main

import "github.com/kevin-cantwell/reunion-explore/index"

// Re-export index types so existing code compiles unchanged.
type Index = index.Index

var BuildIndex = index.BuildIndex
var FormatName = index.FormatName
