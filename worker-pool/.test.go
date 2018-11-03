package main

func main() {
	var ch = make(chan bool)
	go func() {
		close(ch)
	}()
	<-ch
	print(1)
}
