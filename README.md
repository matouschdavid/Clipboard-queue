# cbq (Clipboard Queue)

`cbq` is a macOS clipboard manager that works as a **queue** or **stack**. Copy multiple things, then paste them back one by one in either FIFO or LIFO order — all via global hotkeys, no terminal interaction needed.

## Features

- **Queue mode (default):** Paste items in the same order you copied them (FIFO).
- **Stack mode:** Paste items in reverse order (LIFO).
- **Persistent storage:** Your queue survives restarts — state is saved to `~/.cbq/state.json`.
- **Browser copy buttons:** Clipboard changes made outside of `Cmd+C` (e.g. website "copy to clipboard" buttons) are captured automatically while the queue is active.
- **System notifications:** macOS notifications confirm when the queue is started or stopped.

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

Keep this running in the background (or set it up to launch on login). The version number is printed to the console on startup.

### 2. Global hotkeys

| Hotkey  | Action                                                       |
|---------|--------------------------------------------------------------|
| `Cmd+I` | **Activate** — clears the queue and starts recording copies  |
| `Cmd+C` | Copy as normal — all clipboard changes are captured automatically while active, including browser copy buttons |
| `Cmd+V` | Paste as normal — pops the next item from the queue while active |
| `Cmd+M` | **Toggle mode** — switches between Queue (FIFO) and Stack (LIFO) |
| `Cmd+R` | **Deactivate** — clears the queue and stops recording        |

### 3. Switch mode

Press `Cmd+M` at any time to toggle between Queue and Stack mode. A notification confirms the new mode. The setting is persisted in `~/.cbq/state.json`.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
