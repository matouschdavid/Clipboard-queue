# cbq (Clipboard Queue)

`cbq` is a clipboard manager CLI tool that works as a **stack** or **queue**. It's designed to let you copy multiple things at once and then paste them back one after another in either FIFO (First In First Out) or LIFO (Last In First Out) order.

## Features

- **Clipboard Monitoring:** Runs in the background (or a dedicated terminal) and records everything you copy.
- **Queue Mode (Default):** Paste items in the same order you copied them (FIFO).
- **Stack Mode:** Paste items in the reverse order you copied them (LIFO).
- **Persistent Storage:** Your queue is saved locally, so you don't lose your data if you restart the monitor.

## Installation

### Using Homebrew

```bash
brew tap matouschdavid/Clipboard-queue
brew install cbq
```

*(Note: Tap is not yet public, but that's how it will work)*

### From Source

```bash
git clone https://github.com/matouschdavid/Clipboard-queue.git
cd Clipboard-queue
```
,search:

## Usage

### 1. Start the monitor
To start the background listener for hotkeys and clipboard changes, run:
```bash
cbq start
```
Keep this terminal running (or run in background).

### 2. Control the queue with Hotkeys (macOS)
Once the monitor is started, you can use these global hotkeys:

- **`Cmd + I`**: **Start/Activate** the queue. This clears any previous items and starts recording every `Cmd + C`.
- **`Cmd + C`**: Copy items normally. If `cbq` is active, it records them in order.
- **`Cmd + V`**: Paste items. If `cbq` is active, it pops the next item from the queue to your clipboard just before pasting.
- **`Cmd + R`**: **Reset/Clear** the queue and deactivate it.

### 3. Manual CLI Usage
You can still interact with the queue manually:

#### Check the queue status
```bash
cbq status
```

#### Pop items manually
To get the first item and put it back onto your system clipboard:
```bash
cbq pop
```
For reverse order (stack mode):
```bash
cbq pop --stack
```

#### Clear the queue
```bash
cbq clear
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
[MIT](https://choosealicense.com/licenses/mit/)
