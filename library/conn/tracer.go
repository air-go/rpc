package conn

type ConnTracer interface {
	BeforeRead(*conn, []byte)
	AfterRead(*conn, []byte, int, error)
	BeforeWrite(*conn, []byte)
	AfterWrite(*conn, []byte, int, error)
	BeforeClose(*conn)
	AfterClose(*conn, error)
}
