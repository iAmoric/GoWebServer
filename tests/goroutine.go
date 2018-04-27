package main

import (
    "fmt"
    "time"
)

func printnumbers(k int, l int) {
    for i := k; i <= l; i++ {
        time.Sleep(250 * time.Millisecond)
        fmt.Printf("printnumbers n°%d! ", i)
    }
}

func printletters() {
    for i := 'a'; i <= 'e'; i++ {
        time.Sleep(400 * time.Millisecond)
        fmt.Printf("%c ", i)
    }
}

func main() {
    go printnumbers(1, 5)
    go printnumbers(6, 10)
    time.Sleep(5000 * time.Millisecond)
    fmt.Println("Printing from main")
}
