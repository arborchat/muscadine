// Package archive implements a serializable structure for storing arbor history.
// The `Archive` type provides a simple in-memory store with functions to add, query,
// and persist into a file-like object.
// The `Manager` type wraps an archive and an path and provides functions to store the
// archive's contents into that path and to read the archives contents from that path.
package archive
