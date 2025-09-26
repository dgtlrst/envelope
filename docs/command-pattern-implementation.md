# Command Pattern Implementation Plan

## Overview

This document outlines the implementation plan for introducing the Command Pattern to abstract complex editor operations that involve both buffer and cursor manipulation, while maintaining clean separation of concerns.

## Architecture Goals

- Keep `Buffer` and `Cursor` as separate, focused components
- Abstract complex operations (like backspace) into reusable commands
- Enable extensibility for future complex operations
- Support undo/redo functionality
- Allow for macro recording and key binding customization

## Directory Structure

```
pkg/edit/commands/
├── command.go          // Command interface and base types
├── text_commands.go    // Basic text operations
├── navigation_commands.go // Cursor movement commands
├── complex_commands.go // Multi-step operations
└── macro_commands.go   // Command sequences and macros
```

## Core Types to Implement

### Base Command Interface

```go
type Command interface {
    Execute() *OperationResult
    Undo() *OperationResult
    CanExecute() bool
    Description() string
}

type OperationResult struct {
    Success bool
    NewCursorPos *CursorPosition
    UndoCommand Command
    ErrorMessage string
}

type CursorPosition struct {
    Line, Col int
}
```

### Command Context

```go
type CommandContext struct {
    Buffer Buffer
    Cursor Cursor
}

type BaseCommand struct {
    Context *CommandContext
}
```

## Implementation Steps

### Step 1: Create Base Command Infrastructure
1. Create `pkg/edit/commands/` directory
2. Implement `Command` interface in `command.go`
3. Create `OperationResult` and `CursorPosition` types
4. Implement `CommandContext` and `BaseCommand`

### Step 2: Implement Simple Commands
Create basic commands that replace current app-layer logic:

#### Text Commands (`text_commands.go`)
- `InsertRuneCommand` - Insert single character
- `BackspaceCommand` - Delete character before cursor
- `DeleteCommand` - Delete character at cursor
- `NewLineCommand` - Insert new line and position cursor

#### Navigation Commands (`navigation_commands.go`)
- `MoveLeftCommand`
- `MoveRightCommand` 
- `MoveUpCommand`
- `MoveDownCommand`

### Step 3: Implement Complex Commands
Create commands for advanced operations:

#### Complex Commands (`complex_commands.go`)
- `DeleteWordCommand` - Delete word (Ctrl+Backspace)
- `DeleteLineCommand` - Delete entire line (Ctrl+K)
- `IndentCommand` - Smart indentation (Tab at line start)
- `CommentToggleCommand` - Toggle line/block comments (Ctrl+/)
- `DuplicateLineCommand` - Duplicate current line (Ctrl+D)
- `JoinLinesCommand` - Join current line with next (Ctrl+J)

#### Advanced Text Operations
- `FindReplaceCommand` - Find and replace text (Ctrl+H)
- `GoToLineCommand` - Jump to specific line (Ctrl+G)
- `SelectWordCommand` - Select word under cursor (Ctrl+W)
- `SelectLineCommand` - Select entire line (Ctrl+L)

### Step 4: Implement Macro System
Create support for command sequences:

#### Macro Commands (`macro_commands.go`)
- `MacroCommand` - Execute sequence of commands
- `MacroRecorder` - Record user actions as commands
- `MacroPlayer` - Replay recorded command sequences

### Step 5: Integration with App Layer
1. Modify `app.go` to use command dispatcher
2. Create `CommandDispatcher` that maps keystrokes to commands
3. Replace existing keystroke handling with command execution
4. Implement undo/redo stack using command history

## Command Examples

### Simple Command Example
```go
type BackspaceCommand struct {
    BaseCommand
}

func (cmd *BackspaceCommand) Execute() *OperationResult {
    // Handle cursor at beginning of line (merge with previous)
    // Handle normal character deletion
    // Return new cursor position
}
```

### Complex Command Example
```go
type DeleteWordCommand struct {
    BaseCommand
    Direction WordDirection // Forward/Backward
}

func (cmd *DeleteWordCommand) Execute() *OperationResult {
    // Find word boundaries
    // Delete characters in range
    // Position cursor appropriately
    // Return undo command
}
```

### Macro Command Example
```go
type MacroCommand struct {
    BaseCommand
    Commands []Command
}

func (cmd *MacroCommand) Execute() *OperationResult {
    // Execute commands in sequence
    // Aggregate results
    // Handle partial failures
}
```

## Benefits of This Approach

1. **Extensibility** - Easy to add new commands without modifying existing code
2. **Testability** - Each command can be unit tested independently
3. **Undo/Redo** - Commands naturally support reverse operations
4. **Macros** - Command sequences enable macro recording/playback
5. **Key Binding** - Easy to remap keys to different commands
6. **Separation of Concerns** - App layer becomes pure command dispatch
7. **Reusability** - Commands can be used in different contexts (UI, scripting, etc.)

## Future Enhancements

- **Command History** - Track all executed commands for undo/redo
- **Command Scripting** - Allow commands to be defined in configuration files
- **Plugin System** - Enable third-party commands
- **Command Palette** - UI for discovering and executing commands
- **Keyboard Shortcuts** - Customizable key bindings
- **Command Validation** - Pre-execution validation and error handling