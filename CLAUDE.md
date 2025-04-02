# AutoPDF Specifications & Core Rules

AutoPDF is a CLI tool designed to generate PDFs (and optionally image formats) from LaTeX templates. The tool uses YAML configuration files to define variables, document parts, and other options, processes LaTeX templates with Go’s `text/template` package (using custom delimiters), compiles the resulting LaTeX into a PDF, and optionally converts that PDF to images. This document outlines all core rules, specifications, technologies, and folder structure to get started with development.

---

## Overview

- **Purpose:**  
  Generate PDFs from LaTeX templates by substituting user-defined variables specified in YAML configurations. Optionally, convert PDFs to image formats (PNG/JPG) and clean up auxiliary files generated during compilation.

- **Key Features:**  
  - **YAML-Based Configuration:** Define templates, variables, LaTeX engine, output settings, and conversion options.
  - **Custom Template Processing:**  
    Use Go’s `text/template` package with custom delimiters to avoid conflicts with LaTeX syntax.
  - **LaTeX Compilation:**  
    Invoke external LaTeX engines (like `pdflatex` or `xelatex`) to compile the processed template into a PDF.
  - **Auxiliary File Cleanup:**  
    Remove temporary files generated during LaTeX compilation.
  - **Optional PDF Conversion:**  
    Convert PDFs to image formats using external tools.

---

## Technologies

- **Language:** Go
- **CLI Framework:** Bonzai Tree CLI ([rwxrob/bonzai](https://github.com/rwxrob/bonzai))
- **Configuration:** YAML (using [go-yaml/yaml](https://github.com/go-yaml/yaml))
- **Templating:** Go’s `text/template` package
- **Process Execution:** `os/exec` (for invoking LaTeX engines and conversion tools)
- **Image Conversion (Optional):** External tools (e.g., ImageMagick’s `convert` or `pdftoppm`)

---

## Template Processing Rules

- **Custom Delimiters:**  
  To avoid conflicts with LaTeX syntax, templates must be parsed with custom delimiters:
  ```go
  template.New(templateFile).Funcs(funcMap).Delims("delim[[", "]]")

Function Map (funcMap):
Define any custom helper functions needed during template processing and inject them via the function map.

Folder Structure
Following the golang-standards/project-layout guidelines and our design requirements, the project is structured as follows:

go
Copy
autopdf/
├── cmd/
│   └── autopdf/
│       └── main.go          // CLI entry point using Bonzai Tree; sole file in this directory
├── configs/                 // YAML configuration files
│   └── config.yaml          // Sample configuration file
├── internal/                // Private application logic (non-exported packages)
│   ├── cli/                 // CLI command implementations
│   │   ├── build.go         // 'build' command: processes template and compiles LaTeX
│   │   ├── clean.go         // 'clean' command: removes auxiliary files
│   │   └── convert.go       // 'convert' command: converts PDF to images
│   ├── config/              // YAML configuration parsing and validation
│   │   └── config.go
│   ├── template/            // Template processing using Go's text/template
│   │   └── engine.go        // Uses custom delimiters and funcMap for parsing templates
│   ├── compiler/            // LaTeX compilation routines
│   │   └── compiler.go      // Invokes pdflatex/xelatex via os/exec
│   ├── converter/           // Optional PDF-to-image conversion routines
│   │   └── converter.go
│   └── util/                // Utility functions (logging, error handling, etc.)
│       └── logger.go
├── pkg/                     // Public packages (if needed for external use)
├── docs/                    // Documentation files
├── scripts/                 // Build, installation, or deployment scripts
├── test/                    // Additional tests and test data
├── go.mod                   // Go module file
└── go.sum                   // Go checksum file

### Key Points in the Structure
cmd/autopdf/main.go:
The sole file in the cmd/autopdf directory; it initializes the Bonzai CLI command tree and delegates functionality to the internal packages.

### Internal Packages:
All business logic and functionality are kept in the internal/ directory, ensuring a clear separation between the CLI bootstrap and application logic.

### Modular Commands:

Build: Processes the YAML configuration, applies variables to the LaTeX template, and compiles the output.

Clean: Removes auxiliary files generated during the LaTeX build.

Convert: Optionally converts the compiled PDF into other formats (e.g., PNG, JPG).

### Command Specifications
Build Command:

Function: Read the YAML configuration, process the LaTeX template (using custom delimiters), compile the LaTeX to PDF, and log the output.

Location: internal/cli/build.go

Clean Command:

Function: Remove temporary auxiliary files such as .aux, .log, .toc, etc.

Location: internal/cli/clean.go

Convert Command:

Function: Convert the generated PDF into image formats if conversion is enabled in the configuration.

Location: internal/cli/convert.go

### Development Guidelines

#### Modularity:
Maintain a clear separation of the CLI entry point and the internal application logic.

#### Testing:
Place tests in the test/ directory and ensure all internal modules have robust unit and integration tests.

#### Logging & Error Handling:
Use the utilities in internal/util/logger.go for consistent logging and error management across modules.

#### Documentation:
Keep documentation updated in the docs/ directory and within this CLAUDE.md file as the project evolves.

#### Code Quality:
Follow idiomatic Go practices and ensure the code adheres to the established standards and guidelines.


## Examples to follow (Bonzai Tree)

- [kimono](https://github.com/rwxrob/kimono)
- [bonzai-example](https://github.com/rwxrob/bonzai-example)
- [vars](https://github.com/rwxrob/vars)
- [help](https://github.com/rwxrob/help)

``` go
package kimono

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/comp/completers/git"
	"github.com/rwxrob/bonzai/fn/each"
	"github.com/rwxrob/bonzai/futil"
	"github.com/rwxrob/bonzai/vars"
)

const (
	WorkScopeEnv   = `KIMONO_WORK_SCOPE`
	WorkScopeVar   = `work-scope`
	TagVerPartEnv  = `KIMONO_VERSION_PART`
	TagVerPartVar  = `tag-ver-part`
	TagShortenEnv  = `KIMONO_TAG_SHORTEN`
	TagShortenVar  = `tag-shorten`
	TagRmRemoteEnv = `KIMONO_TAG_RM_REMOTE`
	TagRmRemoteVar = `tag-rm-remote`
	TagPushEnv     = `KIMONO_PUSH_TAG`
	TagPushVar     = `tag-push`
	TidyScopeEnv   = `KIMONO_TIDY_SCOPE`
	TidyScopeVar   = `tidy-scope`
)

var Cmd = &bonzai.Cmd{
	Name:  `kimono`,
	Alias: `kmono|km`,
	Vers:  `v0.7.0`,
	Short: `manage golang monorepos`,
	Long: `
The kimono tool helps manage Go monorepos. It simplifies common monorepo
operations and workflow management.

# Features:
- Toggle go.work files on/off for local development
- Perform coordinated version tagging
- Keep go.mod files tidy across modules
- View dependency graphs and module information
- Track dependent modules and their relationships

# Commands:
- work:     Toggle go.work files for local development
- tidy:     run 'go get -u' and 'go mod tidy' across modules
- tag:      List and coordinate version tagging across modules
- deps:     List and manage module dependencies
- depsonme: List and manage module dependencies
- vars:     View and set configuration variables

Use 'kimono help <command> <subcommand>...' for detailed information
about each command.
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		workCmd,
		tidyCmd,
		tagCmd,
		dependenciesCmd,
		dependentsCmd,
		vars.Cmd,
		help.Cmd,
	},
	Def: help.Cmd,
}

var workCmd = &bonzai.Cmd{
	Name:  `work`,
	Alias: `w`,
	Short: `toggle go work files on or off`,
	Long: `
Work command toggles the state of Go workspace files (go.work) between
active (on) and inactive (off) modes. This is useful for managing
monorepo development by toggling Go workspace configurations. The scope
in which to toggle the work files can be configured using either the
'work-scope' variable or the 'KIMONO_WORK_SCOPE' environment variable.

# Arguments
  on  : Renames go.work.off to go.work, enabling the workspace.
  off : Renames go.work to go.work.off, disabling the workspace.

# Environment Variables

- KIMONO_WORK_SCOPE: module|repo|tree (Defaults to "module")
  Configures the scope in which to toggle.
  - module: Toggles the go.work file in the current module.
  - repo: Toggles all go.work files in the monorepo.
  - tree: Toggles go.work files in the directory tree starting from pwd.
`,
	Vars: bonzai.Vars{
		{
			K:     WorkScopeVar,
			V:     `module`,
			Env:   WorkScopeVar,
			Short: `Configures the scope in which to toggle work files`,
		},
	},
	NumArgs:  1,
	RegxArgs: `on|off`,
	Opts:     `on|off`,
	Comp:     comp.CmdsOpts,
	Cmds: []*bonzai.Cmd{
		workInitCmd,
		help.Cmd.AsHidden(),
		vars.Cmd.AsHidden(),
	},
	Do: func(x *bonzai.Cmd, args ...string) error {
		root := ``
		var err error
		var from, to string
		invArgsErr := fmt.Errorf("invalid arguments: %s", args[0])
		switch args[0] {
		case `on`:
			from = `go.work.off`
			to = `go.work`
		case `off`:
			from = `go.work`
			to = `go.work.off`
		default:
			return invArgsErr
		}
		// FIXME: the default here could come from Env or Vars.
		scope := vars.Fetch(WorkScopeEnv, WorkScopeVar, `module`)
		switch scope {
		case `module`:
			return WorkToggleModule(from, to)
		case `repo`:
			root, err = getGitRoot()
			if err != nil {
				return err
			}
		case `tree`:
			root, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		return WorkToggleRecursive(root, from, to)
	},
}

var workInitCmd = &bonzai.Cmd{
	Name:  `init`,
	Alias: `i`,
	Short: `new go.work in module using dependencies from monorepo`,
	Long: `
The "init" subcommand initializes a new Go workspace file (go.work) 
for the current module. It helps automate the creation of a workspace
file that includes relevant dependencies, streamlining monorepo
development.

# Arguments
  all:     Automatically generates a go.work file with all module
           dependencies from the monorepo.
  modules: Relative path(s) to modules, same as used with 'go work use'.

Run "work init all" to include all dependencies from the monorepo in a 
new go.work file. Alternatively, provide specific module paths to 
initialize a workspace tailored to those dependencies.
`,
	MinArgs:  1,
	RegxArgs: `all`,
	Cmds: []*bonzai.Cmd{
		help.Cmd.AsHidden(),
		vars.Cmd.AsHidden(),
	},
	Do: func(x *bonzai.Cmd, args ...string) error {
		if args[0] == `all` {
			return WorkGenerate()
		}
		return WorkInit(args...)
	},
}

var tagCmd = &bonzai.Cmd{
	Name:  `tag`,
	Alias: `t`,
	Short: `manage or list tags for the go module`,
	Long: `
The tag command helps with listing, smart tagging of modules in a
monorepo. This ensures that all modules are consistently tagged with the
appropriate module prefix and version numbers, facilitating error-free
version control and release management.
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		tagBumpCmd,
		tagListCmd,
		tagDeleteCmd,
		help.Cmd.AsHidden(),
		vars.Cmd.AsHidden(),
	},
	Def: tagListCmd,
}

var tagListCmd = &bonzai.Cmd{
	Name:  `list`,
	Alias: `l`,
	Short: `list the tags for the go module`,
	Long: `
The "list" subcommand displays the list of semantic version (semver)
tags for the current Go module. This is particularly useful for
inspecting version history or understanding the current state of version 
tags in your project.

# Behavior

By default, the command lists all tags that are valid semver tags and 
associated with the current module. The tags can be displayed in their 
full form or shortened by setting the KIMONO_TAG_SHORTEN env var.

# Environment Variables

- KIMONO_TAG_SHORTEN: (Defaults to "true")
  Determines whether to display tags in a shortened format, removing 
  the module prefix. It accepts any truthy value.

# Examples

List tags with the module prefix:

    $ export TAG_SHORTEN=false
    $ tag list

List tags in shortened form (default behavior):

    $ KIMONO_TAG_SHORTEN=1 tag list

The tags are automatically sorted in semantic version order.
`,
	Vars: bonzai.Vars{
		{K: TagShortenVar, V: `true`, Env: TagShortenEnv},
	},
	Do: func(x *bonzai.Cmd, args ...string) error {
		shorten := vars.Fetch(
			TagShortenEnv,
			TagShortenVar,
			false,
		)
		each.Println(TagList(shorten))
		return nil
	},
}

var tagDeleteCmd = &bonzai.Cmd{
	Name:  `delete`,
	Alias: `d|del|rm`,
	Short: `delete the given semver tag for the go module`,
	Long: `
The "delete" subcommand removes a specified semantic version (semver) 
tag. This operation is useful for cleaning up incorrect, outdated, or
unnecessary version tags.
By default, the "delete" command only removes the tag locally. To 
delete a tag both locally and remotely, set the TAG_RM_REMOTE 
environment variable or variable to "true". For example:


# Arguments
  tag: The semver tag to be deleted.

# Environment Variables

- TAG_RM_REMOTE: (Defaults to "false")
  Configures whether the semver tag should also be deleted from the 
  remote repository. Set to "true" to enable remote deletion.

# Examples

    $ tag delete v1.2.3
    $ TAG_RM_REMOTE=true tag delete submodule/v1.2.3

This command integrates with Git to manage semver tags effectively.
`,
	Vars: bonzai.Vars{
		{K: TagRmRemoteVar, V: `false`, Env: TagRmRemoteEnv},
	},
	NumArgs: 1,
	Comp:    comp.Combine{git.CompTags},
	Cmds:    []*bonzai.Cmd{help.Cmd.AsHidden(), vars.Cmd.AsHidden()},
	Do: func(x *bonzai.Cmd, args ...string) error {
		rmRemote := vars.Fetch(
			TagRmRemoteEnv,
			TagRmRemoteVar,
			false,
		)
		return TagDelete(args[0], rmRemote)
	},
}

var tagBumpCmd = &bonzai.Cmd{
	Name:  `bump`,
	Alias: `b|up|i|inc`,
	Short: `bumps semver tags based on given version part.`,
	Long: `
The "bump" subcommand increments the current semantic version (semver) 
tag of the Go module based on the specified version part. This command 
is ideal for managing versioning in a structured manner, following 
semver conventions.

# Arguments
  part: (Defaults to "patch") The version part to increment.
        Accepted values:
          - major (or M): Increments the major version (x.0.0).
          - minor (or m): Increments the minor version (a.x.0).
          - patch (or p): Increments the patch version (a.b.x).

# Environment Variables

- TAG_VER_PART: (Defaults to "patch")
  Specifies the default version part to increment when no argument is 
  passed.

- TAG_PUSH: (Defaults to "false")
  Configures whether the bumped tag should be pushed to the remote 
  repository after being created. Set to "true" to enable automatic 
  pushing. It accepts any truthy value.

# Examples

Increment the version tag locally:

    $ tag bump patch

Automatically push the incremented tag:

    $ TAG_PUSH=true tag bump minor
`,
	Vars: bonzai.Vars{
		{K: TagPushVar, V: `false`, Env: TagPushEnv},
		{K: TagVerPartVar, V: `false`, Env: TagVerPartEnv},
	},
	MaxArgs: 1,
	Opts:    `major|minor|patch|M|m|p`,
	Comp:    comp.CmdsOpts,
	Cmds:    []*bonzai.Cmd{help.Cmd.AsHidden(), vars.Cmd.AsHidden()},
	Do: func(x *bonzai.Cmd, args ...string) error {
		mustPush := vars.Fetch(TagPushEnv, TagPushVar, false)
		if len(args) == 0 {
			part := vars.Fetch(
				TagVerPartEnv,
				TagVerPartVar,
				`patch`,
			)
			return TagBump(optsToVerPart(part), mustPush)
		}
		part := optsToVerPart(args[0])
		return TagBump(part, mustPush)
	},
}

var tidyCmd = &bonzai.Cmd{
	Name:  `tidy`,
	Alias: `tidy|update`,
	Short: "tidy dependencies on all modules in repo",
	Long: `
The "tidy" command updates and tidies the Go module dependencies
across all modules in a monorepo or within a specific scope. This
is particularly useful for maintaining consistency and ensuring
that dependencies are up-to-date.

# Arguments:
  module|mod:          Tidy the current module only.
  repo:                Tidy all modules in the repository.
  deps|dependencies:   Tidy dependencies of the current module in the 
                       monorepo.
  depsonme|dependents: Tidy modules in the monorepo dependent on the 
                       current module.

# Environment Variables:

- KIMONO_TIDY_SCOPE: (Defaults to "module")
  Defines the scope of the tidy operation. Can be set to "module(mod)",
  "root", "dependencies(deps)", or "dependent(depsonme)".

The scope can also be configured using the "tidy-scope" variable or
the "KIMONO_TIDY_SCOPE" environment variable. If no argument is provided,
the default scope is "module".

# Examples:

    # Tidy all modules in the repository
    $ kimono tidy root

    # Tidy only dependencies of the current module in the monorepo
    $ kimono tidy deps

    # Tidy modules in the monorepo dependent on the current module
    $ kimono tidy depsonme

`,
	Vars: bonzai.Vars{
		{K: TidyScopeVar, V: `module`, Env: TidyScopeEnv},
	},
	MaxArgs: 1,
	Opts:    `module|mod|repo|deps|depsonme|dependencies|dependents`,
	Comp:    comp.Opts,
	Cmds:    []*bonzai.Cmd{help.Cmd.AsHidden(), vars.Cmd.AsHidden()},
	Do: func(x *bonzai.Cmd, args ...string) error {
		var scope string
		if len(args) == 0 {
			scope = vars.Fetch(
				TidyScopeEnv,
				TidyScopeVar,
				`module`,
			)
		} else {
			scope = args[0]
		}
		switch scope {
		case `module`:
			pwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return TidyAll(pwd)
		case `repo`:
			root, err := futil.HereOrAbove(".git")
			if err != nil {
				return err
			}
			return TidyAll(filepath.Dir(root))
		case `deps`, `dependencies`:
			TidyDependencies()
		case `depsonme`, `dependents`, `deps-on-me`:
			TidyDependents()
		}
		return nil
	},
}

var dependenciesCmd = &bonzai.Cmd{
	Name:  `dependencies`,
	Alias: `deps`,
	Short: `list or update dependencies`,
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd.AsHidden(),
		vars.Cmd.AsHidden(),
		dependencyListCmd,
		// dependencyUpdateCmd,
	},
	Def: help.Cmd,
}

var dependencyListCmd = &bonzai.Cmd{
	Name:  `list`,
	Alias: `on`,
	Short: `list the dependencies for the go module`,
	Long: `
The list subcommand provides a list of all dependencies for the Go
module. The scope of dependencies can be customized using the options
provided. By default, it lists all dependencies.
`,
	NoArgs: true,
	Cmds:   []*bonzai.Cmd{help.Cmd.AsHidden(), vars.Cmd.AsHidden()},
	Do: func(x *bonzai.Cmd, args ...string) error {
		deps, err := ListDependencies()
		if err != nil {
			return err
		}
		each.Println(deps)
		return nil
	},
}

var dependentsCmd = &bonzai.Cmd{
	Name:  `dependents`,
	Alias: `depsonme`,
	Short: `list or update dependents`,
	Comp:  comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd.AsHidden(),
		vars.Cmd.AsHidden(),
		dependentListCmd,
		// dependentUpdateCmd,
	},
	Def: help.Cmd,
}

var dependentListCmd = &bonzai.Cmd{
	Name:  `list`,
	Alias: `onme`,
	Short: `list the dependents of the go module`,
	Long: `
The list subcommand provides a list of all modules or packages that
depend on the current Go module. This is useful to determine the
downstream impact of changes made to the current module.
`,
	Comp: comp.Cmds,
	Do: func(x *bonzai.Cmd, args ...string) error {
		deps, err := ListDependents()
		if err != nil {
			return err
		}
		if len(deps) == 0 {
			fmt.Println(`None`)
			return nil
		}
		each.Println(deps)
		return nil
	},
}

func optsToVerPart(x string) VerPart {
	switch x {
	case `major`, `M`:
		return Major
	case `minor`, `m`:
		return Minor
	case `patch`, `p`:
		return Patch
	}
	return Minor
}

func argIsOr(args []string, is string, fallback bool) bool {
	if len(args) == 0 {
		return fallback
	}
	return args[0] == is
}

func getGitRoot() (string, error) {
	root, err := futil.HereOrAbove(".git")
	if err != nil {
		return "", err
	}
	return filepath.Dir(root), nil
}
```

using yaml config file

``` go

package inyaml

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rogpeppe/go-internal/lockedfile"
	"github.com/rwxrob/bonzai/futil"
	"gopkg.in/yaml.v3"
)

// Persister represents a key-value storage system using the YAML format.
//
// Features:
//
// - Values are stored as strings and loaded directly into a flat key-value map.
// - The file format is compatible with standard Persister parsers.
//
// # Usage
//
//		     storage := &inyaml.Persister{File: "data.yaml"}
//	    	 storage.Setup()
//		     storage.Set("key", "value")
//		     value := storage.Get("key")
type Persister struct {
	File string // consider someplace in futil.UserStateDir()
}

func NewUserConfig(name, file string) *Persister {
	this := new(Persister)
	dir, err := futil.UserConfigDir()
	if err != nil {
		panic(err)
	}
	f := filepath.Join(dir, name, file)
	err = futil.Touch(f)
	if err != nil {
		panic(err)
	}
	this.File = f
	return this
}

func NewUserCache(name, file string) *Persister {
	this := new(Persister)
	dir, err := futil.UserCacheDir()
	if err != nil {
		panic(err)
	}
	f := filepath.Join(dir, name, file)
	err = futil.Touch(f)
	if err != nil {
		panic(err)
	}
	this.File = f
	return this
}

func NewUserState(name, file string) *Persister {
	this := new(Persister)
	dir, err := futil.UserStateDir()
	if err != nil {
		panic(err)
	}
	f := filepath.Join(dir, name, file)
	err = futil.Touch(f)
	if err != nil {
		panic(err)
	}
	this.File = f
	return this
}

// Setup ensures that the persistence file exists and is ready for use.
// It opens the file in read-only mode, creating it if it doesn't already exist,
// and applies secure file permissions (0600) to restrict access.
// The file is immediately closed after being verified or created.
// If the file cannot be opened or created, an error is returned.
func (p *Persister) Setup() error {
	f, err := lockedfile.OpenFile(p.File, os.O_RDONLY|os.O_CREATE, 0600)
	defer f.Close()
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the value associated with the given key from the persisted
// data. The method loads the data from the persistence file in a way
// that is both safe for concurrency and locked against use by other
// programs using the lockedfile package (like go binary itself does).
// If the key is not present in the data, a nil value will be
// returned.
func (p *Persister) Get(key string) string {
	data := p.loadFile()
	return data[key]
}

// Set stores a key-value pair in the persisted data. The method loads
// the existing data from the persistence file, updates the data with
// the new key-value pair, and saves it back to the file.
func (p *Persister) Set(key, value string) {
	data := p.loadFile()
	data[key] = value
	p.saveFile(data)
}

func (p *Persister) loadFile() map[string]string {
	data := make(map[string]string)
	f, err := lockedfile.OpenFile(p.File, os.O_RDONLY, 0600)
	defer f.Close()
	if err != nil {
		return data
	}
	content, err := io.ReadAll(f)
	if err != nil || len(content) == 0 {
		return data
	}
	yaml.Unmarshal(content, &data)
	return data
}

func (p *Persister) saveFile(data map[string]string) {
	f, err := lockedfile.OpenFile(p.File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		return
	}
	content, _ := yaml.Marshal(data)
	f.Write(content)
}
```

``` go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/arara/internal/pkg/vars"
	"github.com/rwxrob/bonzai/persisters/inyaml"
	bonzaiVars "github.com/rwxrob/bonzai/vars"
	"gopkg.in/yaml.v3"
)

// Config represents the global arara configuration
type Config struct {
	Namespaces []string          `yaml:"namespaces"`
	Configs    map[string]NSInfo `yaml:"configs"`
}

// DotfilesConfig represents a local dotfiles configuration (arara.yaml)
type DotfilesConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Env         map[string]string `yaml:"env,omitempty"`
	Namespace   string            `yaml:"namespace"`

	Dependencies []string `yaml:"dependencies,omitempty"`

	Setup struct {
		BackupDirs  []string `yaml:"backup_dirs"`
		CoreLinks   []Link   `yaml:"core_links"`
		ConfigLinks []Link   `yaml:"config_links"`
	} `yaml:"setup"`

	Build struct {
		Steps []Step `yaml:"steps"`
	} `yaml:"build"`

	Scripts struct {
		Install []Script `yaml:"install,omitempty"`
	} `yaml:"scripts,omitempty"`
}

type Link struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

type Step struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Command     string        `yaml:"command,omitempty"`
	Commands    []string      `yaml:"commands,omitempty"`
	Compat      *CompatConfig `yaml:"compat,omitempty"`
}

type Script struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Path        string        `yaml:"path"`
	Compat      *CompatConfig `yaml:"compat,omitempty"`
}

// String implements fmt.Stringer for interactive selection
func (s Script) String() string {
	return fmt.Sprintf("%s: %s", s.Name, s.Description)
}

type CompatConfig struct {
	OS     string        `yaml:"os,omitempty"`
	Arch   string        `yaml:"arch,omitempty"`
	Shell  string        `yaml:"shell,omitempty"`
	PkgMgr string        `yaml:"pkgmgr,omitempty"`
	Kernel string        `yaml:"kernel,omitempty"`
	Custom []interface{} `yaml:"custom,omitempty"`
}

func LoadConfig(path string) (*DotfilesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config DotfilesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Only validate namespace if it's a local config and we're not in a test environment
	if filepath.Base(path) == "arara.yaml" && os.Getenv("TEST_MODE") != "1" {
		// Load global config to validate namespace
		gc, err := NewGlobalConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load global config: %w", err)
		}

		// Validate namespace exists
		if config.Namespace != "" {
			found := false
			for _, ns := range gc.Config.Namespaces {
				if ns == config.Namespace {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("undefined namespace: %s", config.Namespace)
			}
		}
	}

	return &config, nil
}

func GetConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "arara")
}

// Marshal returns the YAML representation of the config
func (c *DotfilesConfig) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

// GlobalConfig represents the global arara configuration
type GlobalConfig struct {
	Config
	persister *inyaml.Persister
}

// Save persists the global config
func (gc *GlobalConfig) Save() error {
	data, err := yaml.Marshal(gc.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	gc.persister.Set("config", string(data))
	return nil
}

// GetDotfilesPath returns the path to the active namespace's dotfiles
func GetDotfilesPath() (string, error) {
	gc, err := NewGlobalConfig()
	if err != nil {
		return "", err
	}

	ns := gc.GetActiveNamespace()
	if ns == nil {
		return "", fmt.Errorf("no active namespace")
	}

	return ns.Path, nil
}

// GetActiveNamespace returns the active namespace's configuration
func (gc *GlobalConfig) GetActiveNamespace() *NamespaceConfig {
	activeNS := bonzaiVars.Fetch(vars.ActiveNamespaceEnv, vars.ActiveNamespaceVar, "")
	if activeNS == "" {
		return nil
	}
	if info, ok := gc.Configs[activeNS]; ok {
		return &NamespaceConfig{
			Name:     activeNS,
			Path:     info.Path,
			LocalBin: info.LocalBin,
		}
	}
	return nil
}

type NamespaceConfig struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	LocalBin string `yaml:"local-bin"`
}

type NSInfo struct {
	Path     string   `yaml:"path"`
	LocalBin string   `yaml:"local-bin"`
	Dirs     []string `yaml:"backup_dirs"`
}

var NewGlobalConfig = func() (*GlobalConfig, error) {
	persister := inyaml.NewUserConfig("arara", "config.yaml")

	gc := &GlobalConfig{
		persister: persister,
		Config: Config{
			// Initialize only what's needed for namespace management
			Namespaces: make([]string, 0),
			Configs:    make(map[string]NSInfo),
		},
	}

	// Load existing config
	if err := gc.load(); err != nil {
		return nil, err
	}

	return gc, nil
}

func (gc *GlobalConfig) load() error {
	data := gc.persister.Get("config")
	if data == "" {
		return nil
	}

	if err := yaml.Unmarshal([]byte(data), &gc.Config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}

// Add namespace management methods to GlobalConfig
func (gc *GlobalConfig) AddNamespace(name, path, localBin string) error {
	// Check if namespace already exists
	for _, ns := range gc.Namespaces {
		if ns == name {
			return fmt.Errorf("namespace %s already exists", name)
		}
	}

	// Add namespace
	gc.Namespaces = append(gc.Namespaces, name)
	gc.Configs[name] = NSInfo{
		Path:     path,
		LocalBin: localBin,
	}

	return gc.Save()
}
``` 