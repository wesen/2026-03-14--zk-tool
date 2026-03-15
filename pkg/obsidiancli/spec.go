package obsidiancli

// OutputKind describes how command stdout should be parsed.
type OutputKind string

const (
	OutputRaw      OutputKind = "raw"
	OutputJSON     OutputKind = "json"
	OutputLineList OutputKind = "line-list"
	OutputKeyValue OutputKind = "key-value"
)

// CommandSpec describes one Obsidian CLI command surface.
type CommandSpec struct {
	Name        string
	Output      OutputKind
	Description string
}

var (
	// SpecVersion reports the CLI or application version.
	SpecVersion = CommandSpec{Name: "version", Output: OutputRaw, Description: "Report Obsidian version"}

	// File and content operations.
	SpecFilesList   = CommandSpec{Name: "files", Output: OutputLineList, Description: "List files in the vault"}
	SpecFileRead    = CommandSpec{Name: "read", Output: OutputRaw, Description: "Read a single file"}
	SpecFileCreate  = CommandSpec{Name: "create", Output: OutputRaw, Description: "Create a file"}
	SpecFileAppend  = CommandSpec{Name: "append", Output: OutputRaw, Description: "Append to a file"}
	SpecFilePrepend = CommandSpec{Name: "prepend", Output: OutputRaw, Description: "Prepend to a file"}
	SpecFileMove    = CommandSpec{Name: "move", Output: OutputRaw, Description: "Move a file"}
	SpecFileRename  = CommandSpec{Name: "rename", Output: OutputRaw, Description: "Rename a file"}
	SpecFileTrash   = CommandSpec{Name: "delete", Output: OutputRaw, Description: "Delete a file via trash"}
	SpecFileDelete  = CommandSpec{Name: "delete", Output: OutputRaw, Description: "Delete a file permanently"}

	// Search and graph operations.
	SpecSearch          = CommandSpec{Name: "search", Output: OutputLineList, Description: "Search vault content"}
	SpecSearchContext   = CommandSpec{Name: "search:context", Output: OutputJSON, Description: "Search vault content with context"}
	SpecLinksBacklinks  = CommandSpec{Name: "backlinks", Output: OutputJSON, Description: "Return backlinks for a note"}
	SpecLinksOutgoing   = CommandSpec{Name: "links", Output: OutputLineList, Description: "Return outgoing links for a note"}
	SpecLinksOrphans    = CommandSpec{Name: "orphans", Output: OutputLineList, Description: "Return orphan notes"}
	SpecLinksDeadEnds   = CommandSpec{Name: "deadends", Output: OutputLineList, Description: "Return dead-end notes"}
	SpecLinksUnresolved = CommandSpec{Name: "unresolved", Output: OutputJSON, Description: "Return unresolved links"}

	// Metadata and task operations.
	SpecTagsList         = CommandSpec{Name: "tags", Output: OutputJSON, Description: "List tags"}
	SpecTagsRename       = CommandSpec{Name: "tags:rename", Output: OutputRaw, Description: "Rename a tag"}
	SpecPropertiesList   = CommandSpec{Name: "properties", Output: OutputJSON, Description: "List properties"}
	SpecPropertiesGet    = CommandSpec{Name: "properties:get", Output: OutputRaw, Description: "Read one property"}
	SpecPropertiesSet    = CommandSpec{Name: "properties:set", Output: OutputRaw, Description: "Set one property"}
	SpecPropertiesDelete = CommandSpec{Name: "properties:delete", Output: OutputRaw, Description: "Delete one property"}
	SpecTasksList        = CommandSpec{Name: "tasks", Output: OutputJSON, Description: "List tasks"}
	SpecTasksToggle      = CommandSpec{Name: "tasks:toggle", Output: OutputRaw, Description: "Toggle a task"}
	SpecTasksDone        = CommandSpec{Name: "tasks:done", Output: OutputRaw, Description: "Mark a task done"}

	// Daily notes and templates.
	SpecDailyOpen      = CommandSpec{Name: "daily:open", Output: OutputRaw, Description: "Open today's daily note"}
	SpecDailyRead      = CommandSpec{Name: "daily:read", Output: OutputRaw, Description: "Read today's daily note"}
	SpecDailyAppend    = CommandSpec{Name: "daily:append", Output: OutputRaw, Description: "Append to today's daily note"}
	SpecDailyPrepend   = CommandSpec{Name: "daily:prepend", Output: OutputRaw, Description: "Prepend to today's daily note"}
	SpecDailyPath      = CommandSpec{Name: "daily:path", Output: OutputRaw, Description: "Return today's daily note path"}
	SpecTemplatesList  = CommandSpec{Name: "templates:list", Output: OutputLineList, Description: "List templates"}
	SpecTemplateRead   = CommandSpec{Name: "template:read", Output: OutputRaw, Description: "Read a template"}
	SpecTemplateInsert = CommandSpec{Name: "template:insert", Output: OutputRaw, Description: "Insert a template"}

	// Application-level operations.
	SpecPluginsList   = CommandSpec{Name: "plugins", Output: OutputJSON, Description: "List plugins"}
	SpecPluginEnable  = CommandSpec{Name: "plugin:enable", Output: OutputRaw, Description: "Enable a plugin"}
	SpecPluginDisable = CommandSpec{Name: "plugin:disable", Output: OutputRaw, Description: "Disable a plugin"}
	SpecPluginInstall = CommandSpec{Name: "plugin:install", Output: OutputRaw, Description: "Install a plugin"}
	SpecPluginReload  = CommandSpec{Name: "plugin:reload", Output: OutputRaw, Description: "Reload a plugin"}
	SpecThemeSet      = CommandSpec{Name: "theme:set", Output: OutputRaw, Description: "Set the active theme"}
	SpecVaultsList    = CommandSpec{Name: "vaults", Output: OutputLineList, Description: "List available vaults"}
	SpecVaultInfo     = CommandSpec{Name: "vault", Output: OutputKeyValue, Description: "Return active vault metadata"}
	SpecVaultOpen     = CommandSpec{Name: "vault:open", Output: OutputRaw, Description: "Open a vault"}
	SpecEval          = CommandSpec{Name: "eval", Output: OutputRaw, Description: "Evaluate JavaScript in Obsidian"}
)
