# cbq (Clipboard Queue)

`cbq` is a macOS clipboard manager that works as a **queue** or **stack**. Copy multiple things, then paste them back one by one in either FIFO or LIFO order — all via global hotkeys, no terminal interaction needed.

## Features

- **Queue mode (default):** Paste items in the same order you copied them (FIFO).
- **Stack mode:** Paste items in reverse order (LIFO).
- **Persistent storage:** Your queue survives restarts — state is saved to `~/.cbq/state.json`.

## Installation

### Using Homebrew

```bash
brew tap matouschdavid/cbq https://github.com/matouschdavid/Clipboard-queue.git
brew install cbq
```

### From Source

```bash
git clone https://github.com/matouschdavid/Clipboard-queue.git
cd Clipboard-queue
go install
```

## Usage

### 1. Start the monitor

```bash
cbq
```

Keep this running in the background (or set it up to launch on login).

### 2. Global hotkeys

| Hotkey  | Action                                                       |
|---------|--------------------------------------------------------------|
| `Cmd+I` | **Activate** — clears the queue and starts recording copies  |
| `Cmd+C` | Copy as normal — recorded to the queue while active          |
| `Cmd+V` | Paste as normal — pops the next item from the queue while active |
| `Cmd+R` | **Deactivate** — clears the queue and stops recording        |

### 3. Switch mode

The mode is persisted in `~/.cbq/state.json`. Edit the `"is_stack"` field directly:

- `false` — queue mode (FIFO, default)
- `true` — stack mode (LIFO)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
