package doc

import (
	"embed"

	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed *
var docFS embed.FS

// AddDocToHelpSystem loads embedded markdown help entries into the Glazed help system.
func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}
