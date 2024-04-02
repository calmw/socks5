package main

func main() {
	//server := NewServer("0.0.0.0", WithPort(6666))
	server := NewServer("0.0.0.0", WithPort(6666))

	server.ListenAndServe()
}
