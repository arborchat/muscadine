package main

import "path"

var (
	// UserDataPath is the path within a user's home directory in which application
	// data should be stored. Muscadine uses this to determine the default
	// location of its logs and chat history files.
	UserDataPath = path.Join(".local", "share", "muscadine")
)
