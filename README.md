# piper
Chain external commands in Go

## docs
[godoc](https://pkg.go.dev/github.com/noxer/piper)

## description
This package enables you to chain `*exec.Cmd`s and to pipe the output from one process into another without having to manually set up the chain.

## usage
This example executes the command `ls -al`, pipes the output to `grep main` and prints the result. It is the equivalent of running `ls -al | grep main` in your console.

    package main

    import (
    	"fmt"
    
    	"github.com/noxer/piper"
    )
    
    func main() {
    
    	p := piper.Command("ls", "-al").Command("grep", "main")
    	o, _ := p.CombinedOutput() // You should check the error!
    	fmt.Println(string(o))
    
    }

The `*piper.Chain` exposes almost the same API as a single `*exec.Cmd` so you can use it as a drop-in replacement most of the time.
