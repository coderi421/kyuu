package errhdl

import (
	"github.com/coderi421/kyuu"
	"net/http"
	"testing"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := NewMiddlewareBuilder()
	builder.AddCode(http.StatusNotFound, []byte(`
<html>
	<body>
		<h1>404</h1>
	</body>
</html>
`)).
		AddCode(http.StatusBadRequest, []byte(`
<html>
	<body>
		<h1>500</h1>
	</body>
</html>
`))
	server := kyuu.NewHTTPServer()
	server.Use(builder.Build())
	server.Start(":8081")
}
