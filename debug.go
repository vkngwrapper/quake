//go:build debug

package main

var CVarRenderRayDebug = CVar{
	Name:      "r_raydebug",
	StringVal: "0",
}

func InitDebug() {
	CVars.Register(&CVarRenderRayDebug)
}
