package main

func main() {
	server, err := New()
	if err != nil {
		panic(err)
	}

	server.ListenAndServe("0.0.0.0:6666")
}
