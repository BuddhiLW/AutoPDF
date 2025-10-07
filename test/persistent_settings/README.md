# Persistent Settings Example

This example demonstrates AutoPDF's persistent CLI settings functionality, showing how settings are saved and managed across sessions.

## What This Example Shows

- **Persistent CLI settings** that survive across sessions
- **Setting management commands** (on, off, switch, status)
- **Configuration storage** in user home directory
- **Automatic settings persistence** using Bonzai persisters
- **Cross-platform compatibility** with fallback options

## Files

- `config.yaml`: Configuration with settings demonstration
- `settings_document.tex`: LaTeX template explaining persistent settings
- `README.md`: This documentation

## Running the Example

```bash
cd test/persistent_settings
autopdf build settings_document.tex config.yaml
```

## Expected Output

- `settings_document.pdf`: Generated PDF explaining persistent settings

## Key Features Demonstrated

- ✅ **Persistent Settings**: Settings saved across CLI sessions
- ✅ **Setting Commands**: `verbose`, `clean`, `debug`, `force`
- ✅ **Configuration Storage**: Automatic file management
- ✅ **Status Checking**: View current settings
- ✅ **Toggle Operations**: Switch settings on/off
- ✅ **Cross-Platform**: Works on all operating systems

## Available Settings

### Verbose Settings
```bash
# Set specific verbosity level (0-4)
autopdf verbose 3

# Enable verbose logging
autopdf verbose on

# Disable verbose logging
autopdf verbose off
```

### Clean Settings
```bash
# Enable auxiliary file cleaning
autopdf clean on

# Disable cleaning
autopdf clean off

# Toggle cleaning
autopdf clean switch

# Check current status
autopdf clean status
```

### Debug Settings
```bash
# Enable debug information
autopdf debug on

# Disable debug
autopdf debug off

# Toggle debug
autopdf debug switch
```

### Force Settings
```bash
# Enable force mode (overwrite files)
autopdf force on

# Disable force mode
autopdf force off

# Toggle force mode
autopdf force switch
```

## Configuration Storage

Settings are automatically stored in:
- **Primary**: `~/.autopdf/config.yaml` (user home directory)
- **Fallback**: System temp directory if home directory unavailable
- **Format**: YAML configuration file
- **Permissions**: Secure file permissions (0755)

## Setting Levels

### Verbose Levels
- **0 (Silent)**: Only errors
- **1 (Basic)**: Warnings and above
- **2 (Detailed)**: Info and above
- **3 (Debug)**: Debug and above
- **4 (Maximum)**: All logs with full introspection

### Clean Options
- **Enabled**: Remove auxiliary LaTeX files after compilation
- **Disabled**: Keep auxiliary files
- **Toggle**: Switch between enabled/disabled

### Debug Options
- **Enabled**: Show debug information
- **Output**: stdout, file, or custom path
- **Toggle**: Switch debug on/off

### Force Options
- **Enabled**: Force operations and overwrite existing files
- **Disabled**: Respect existing files
- **Toggle**: Switch force mode

## Best Practices

- Use persistent settings for frequently used options
- Check current settings with status commands
- Reset to defaults when needed
- Export/import configurations for sharing
- Use appropriate verbosity levels for different tasks
