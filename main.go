package main

func main() {
	args := newArgs()
	args.Hello()
	srv := Server{args: &args}
	srv.RunForever()
}
