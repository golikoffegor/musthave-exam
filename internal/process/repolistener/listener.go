package repolistener

type Listener struct {
	NewOrderChan chan string
}

func NewListener() *Listener {
	return &Listener{
		NewOrderChan: make(chan string),
	}
}
