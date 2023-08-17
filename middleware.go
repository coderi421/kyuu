package kyuu

type Middleware func(next HandleFunc) HandleFunc
