// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	application "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

// FormatCacheDir is the directory, relative to the working directory, that
// holds precompiled format files (LaTeXOptions.FormatFile).
const FormatCacheDir = ".autopdf-fmt"

// beginDocument ends the static preamble when no \endofdump marker comes first.
const beginDocument = `\begin{document}`

// dumpMarkers end the part of the preamble that goes into the format file;
// mylatexformat stops reading at whichever comes first.
var dumpMarkers = []string{`\endofdump`, `\csname endofdump\endcsname`, beginDocument}

// splitPreamble returns the static preamble of content: everything before the
// first dump marker. ok is false when content has no \begin{document} or an
// empty preamble.
func splitPreamble(content string) (preamble string, ok bool) {
	if !strings.Contains(content, beginDocument) {
		return "", false
	}
	end := -1
	for _, marker := range dumpMarkers {
		if i := strings.Index(content, marker); i >= 0 && (end < 0 || i < end) {
			end = i
		}
	}
	if end <= 0 || strings.TrimSpace(content[:end]) == "" {
		return "", false
	}
	return content[:end], true
}

// FormatName is the cache key of a preamble compiled by engine: the engine's
// base name and a hash of engine and preamble.
func FormatName(engine, preamble string) string {
	base := filepath.Base(engine)
	sum := sha256.Sum256([]byte(base + "\x00" + preamble))
	return base + "-" + hex.EncodeToString(sum[:16])
}

// prepareFormat returns the -fmt value (relative to workingDir) of a format
// holding content's preamble, building it when the cache has none. Any
// failure is logged and yields "", which compiles without a format.
func (lca *LaTeXCompilerAdapter) prepareFormat(ctx context.Context, engine, workingDir, content string, debug bool) string {
	logger := lca.log()
	preamble, ok := splitPreamble(content)
	if !ok {
		logger.Warn(ctx, "format_file: no static preamble before \\begin{document}; compiling without a format file")
		return ""
	}
	name := FormatName(engine, preamble)
	fmtArg := filepath.Join(FormatCacheDir, name)
	cacheDir := filepath.Join(workingDir, FormatCacheDir)
	fmtPath := filepath.Join(cacheDir, name+".fmt")

	if lca.formatValid(ctx, workingDir, fmtPath, filepath.Join(cacheDir, name+".deps")) {
		logger.Info(ctx, "format_file: using cached format", application.NewLogField("format", fmtArg))
		return fmtArg
	}

	markerPath := filepath.Join(cacheDir, name+".failed")
	version := lca.engineVersion(ctx, engine, workingDir)
	if marker, err := lca.fileSystem.ReadFile(ctx, markerPath); err == nil {
		if string(marker) == version {
			logger.Info(ctx, "format_file: this preamble failed to dump with this engine before; compiling without a format file",
				application.NewLogField("marker", markerPath))
			return ""
		}
		_ = lca.fileSystem.Remove(ctx, markerPath)
	}

	if err := lca.buildFormat(ctx, engine, workingDir, name, preamble, debug); err != nil {
		if ctx.Err() != nil {
			return ""
		}
		logger.Warn(ctx, "format_file: building the format failed; compiling without it",
			application.NewLogField("format", fmtArg),
			application.NewLogField("error", err.Error()))
		_ = lca.fileSystem.WriteFile(ctx, markerPath, []byte(version), 0o644)
		return ""
	}
	logger.Info(ctx, "format_file: built format", application.NewLogField("format", fmtArg))
	return fmtArg
}

// buildFormat dumps preamble into <cache>/<name>.fmt with mylatexformat and
// records the working-directory files the preamble read in <name>.deps.
func (lca *LaTeXCompilerAdapter) buildFormat(ctx context.Context, engine, workingDir, name, preamble string, debug bool) error {
	cacheDir := filepath.Join(workingDir, FormatCacheDir)
	if err := lca.fileSystem.MkdirAll(ctx, cacheDir, 0o755); err != nil {
		return fmt.Errorf("create format cache: %w", err)
	}
	tmpJob := name + "-" + randomSuffix()
	staged := stagedSourcePath(workingDir, tmpJob)
	source := preamble + "\n" + beginDocument + `\end{document}` + "\n"
	if err := lca.fileSystem.WriteFile(ctx, staged, []byte(source), 0o644); err != nil {
		return fmt.Errorf("stage preamble: %w", err)
	}
	tmpFmt := filepath.Join(cacheDir, tmpJob+".fmt")
	tmpLog := filepath.Join(cacheDir, tmpJob+".log")
	tmpFls := filepath.Join(cacheDir, tmpJob+".fls")
	defer func() {
		if !debug {
			removeStaged(ctx, lca.fileSystem, staged)
			_ = lca.fileSystem.Remove(context.WithoutCancel(ctx), tmpLog)
		}
		_ = lca.fileSystem.Remove(context.WithoutCancel(ctx), tmpFls)
		_ = lca.fileSystem.Remove(context.WithoutCancel(ctx), tmpFmt)
	}()

	cmd := application.NewCommand(engine, FormatBuildArgs(engine, tmpJob, filepath.Base(staged)), workingDir).
		WithTimeout(2 * time.Minute).
		WithEnv(sandboxEnv())
	result, err := lca.executor.Execute(ctx, cmd)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	if err == nil && result.ExitCode != 0 {
		err = fmt.Errorf("exit status %d", result.ExitCode)
	}
	if err != nil {
		return fmt.Errorf("%s -ini: %w (log: %s)", engine, err, tmpLog)
	}
	if info, statErr := lca.fileSystem.Stat(ctx, tmpFmt); statErr != nil || info.Size() == 0 {
		return fmt.Errorf("%s -ini wrote no format file", engine)
	}

	fls, _ := lca.fileSystem.ReadFile(ctx, tmpFls)
	manifest, err := lca.depsManifest(ctx, workingDir, recordedInputs(string(fls), filepath.Base(staged)))
	if err != nil {
		return err
	}
	tmpDeps := filepath.Join(cacheDir, tmpJob+".deps")
	if err := lca.fileSystem.WriteFile(ctx, tmpDeps, []byte(manifest), 0o644); err != nil {
		return fmt.Errorf("write format dependencies: %w", err)
	}
	if err := lca.fileSystem.Rename(ctx, tmpDeps, filepath.Join(cacheDir, name+".deps")); err != nil {
		_ = lca.fileSystem.Remove(ctx, tmpDeps)
		return fmt.Errorf("publish format dependencies: %w", err)
	}
	if err := lca.fileSystem.Rename(ctx, tmpFmt, filepath.Join(cacheDir, name+".fmt")); err != nil {
		return fmt.Errorf("publish format: %w", err)
	}
	return nil
}

// FormatBuildArgs is the argument list of the sandboxed -ini run that dumps
// stagedPreamble (relative to the working directory) as jobName.fmt.
func FormatBuildArgs(engine, jobName, stagedPreamble string) []string {
	return []string{
		"-ini",
		"-interaction=nonstopmode",
		noShellEscape,
		"-recorder",
		"-jobname=" + jobName,
		"-output-directory=" + FormatCacheDir,
		"&" + filepath.Base(engine),
		"mylatexformat.ltx",
		stagedPreamble,
	}
}

// formatValid reports whether fmtPath exists and every working-directory
// file listed in depsPath still has the recorded content.
func (lca *LaTeXCompilerAdapter) formatValid(ctx context.Context, workingDir, fmtPath, depsPath string) bool {
	info, err := lca.fileSystem.Stat(ctx, fmtPath)
	if err != nil || info.Size() == 0 {
		return false
	}
	manifest, err := lca.fileSystem.ReadFile(ctx, depsPath)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(strings.TrimSpace(string(manifest)), "\n") {
		if line == "" {
			continue
		}
		sum, path, found := strings.Cut(line, "  ")
		if !found {
			return false
		}
		data, err := lca.fileSystem.ReadFile(ctx, filepath.Join(workingDir, path))
		if err != nil || hashHex(data) != sum {
			return false
		}
	}
	return true
}

// depsManifest lists "sha256  path" for each recorded input.
func (lca *LaTeXCompilerAdapter) depsManifest(ctx context.Context, workingDir string, inputs []string) (string, error) {
	var b strings.Builder
	for _, path := range inputs {
		data, err := lca.fileSystem.ReadFile(ctx, filepath.Join(workingDir, path))
		if err != nil {
			return "", fmt.Errorf("hash format dependency %s: %w", path, err)
		}
		fmt.Fprintf(&b, "%s  %s\n", hashHex(data), path)
	}
	return b.String(), nil
}

// recordedInputs returns the relative INPUT paths of a -recorder .fls file,
// without the staged preamble and the format cache itself. Absolute paths
// belong to the TeX distribution and are not tracked.
func recordedInputs(fls, stagedPreamble string) []string {
	seen := map[string]bool{}
	var out []string
	for _, line := range strings.Split(fls, "\n") {
		path, ok := strings.CutPrefix(strings.TrimSpace(line), "INPUT ")
		if !ok || filepath.IsAbs(path) {
			continue
		}
		path = filepath.Clean(path)
		if path == stagedPreamble || path == FormatCacheDir ||
			strings.HasPrefix(path, FormatCacheDir+string(filepath.Separator)) ||
			strings.HasPrefix(path, "..") || seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

// discardFormat removes a format a compile could not use, so the next
// compile rebuilds it.
func (lca *LaTeXCompilerAdapter) discardFormat(ctx context.Context, workingDir, fmtArg string) {
	base := filepath.Join(workingDir, fmtArg)
	_ = lca.fileSystem.Remove(context.WithoutCancel(ctx), base+".fmt")
	_ = lca.fileSystem.Remove(context.WithoutCancel(ctx), base+".deps")
}

// engineVersion is the first line of `engine --version`, "" when unknown.
func (lca *LaTeXCompilerAdapter) engineVersion(ctx context.Context, engine, workingDir string) string {
	result, err := lca.executor.Execute(ctx,
		application.NewCommand(engine, []string{"--version"}, workingDir).WithTimeout(30*time.Second))
	if err != nil || result.ExitCode != 0 {
		return ""
	}
	first, _, _ := strings.Cut(result.Stdout, "\n")
	return strings.TrimSpace(first)
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func randomSuffix() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%08x", time.Now().UnixNano()&0xffffffff)
	}
	return hex.EncodeToString(b[:])
}
