// Command sb runs commands in isolated Linux namespaces.
package main

import (
	"github.com/carlosgrillet/sandbox/cmd/sandbox"
)

func main() {
	sandbox.Bootstrap()
}
