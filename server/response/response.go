package response

type Response struct {
	Body        []byte
	ContentType string
	Cookie      map[string]string
	Headers     map[string]string
	StatusCode  int
}
