package arbor

// Store is a data structure that holds ChatMessages and allows them
// to be easily looked up by their identifiers. It is safe for
// concurrent use.
type Store struct {
	m        map[string]*ChatMessage
	add      chan *ChatMessage
	request  chan string
	response chan *ChatMessage
}

// NewStore creates a Store that is ready to be used.
func NewStore() *Store {
	s := &Store{
		m:        make(map[string]*ChatMessage),
		add:      make(chan *ChatMessage),
		request:  make(chan string),
		response: make(chan *ChatMessage),
	}
	go s.dispatch()
	return s
}

func (s *Store) dispatch() {
	for {
		select {
		case msg := <-s.add:
			s.m[msg.UUID] = msg
		case id := <-s.request:
			value := s.m[id]
			s.response <- value
		}
	}
}

// Get retrieves the message with a UUID from the store.
func (s *Store) Get(uuid string) *ChatMessage {
	s.request <- uuid
	return <-s.response
}

// Add inserts the given message into the store.
func (s *Store) Add(msg *ChatMessage) {
	s.add <- msg
}
