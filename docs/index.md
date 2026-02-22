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
brew tap vibe-coding/cbq
brew install cbq
```

*(Note: Tap is not yet public, but that's how it will work)*

### From Source

```bash
git clone https://github.com/vibe-coding/cbq.git
cd cbq
go install
```

## Usage

### 1. Start the monitor
To start recording your clipboard, run:
```bash
cbq start
```

### 2. Copy multiple items
Just use your regular `Cmd+C` or `Right-click > Copy` on several pieces of text.

### 3. Paste the items back
To get the first item you copied and put it back onto your system clipboard:
```bash
cbq pop
```
Then use `Cmd+V` as usual. Repeat this to get the next items.

To paste items in reverse order (stack mode):
```bash
cbq pop --stack
# or
cbq pop -s
```

### 4. Check the queue status
To see what's currently in your queue:
```bash
cbq status
```

### 5. Clear the queue
To start fresh:
```bash
cbq clear
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
[MIT](https://choosealicense.com/licenses/mit/)
