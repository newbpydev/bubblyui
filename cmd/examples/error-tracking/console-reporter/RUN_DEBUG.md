# Debug Instructions

## How to Run with Debug Logging

1. **Build the debug version:**
   ```bash
   cd /home/newbpydev/Development/Xoomby/bubblyui
   go build -o /tmp/console-reporter-debug ./cmd/examples/error-tracking/console-reporter/
   ```

2. **Run the application:**
   ```bash
   cd /tmp
   ./console-reporter-debug
   ```

3. **Test the panic:**
   - Press some number keys (e.g., `4`, `4`)
   - Press an operator (e.g., `+`)
   - Press more numbers
   - Press `=` to calculate
   - **Press `p` to trigger the panic**
   - Observe what happens

4. **Exit the app:**
   - If it freezes, press `Ctrl+C` to force quit
   - If it works, press `q` to quit normally

5. **Check the debug log:**
   ```bash
   cat /tmp/debug.log
   ```

## What to Look For in debug.log

The log will show:
- **MAIN**: Application lifecycle events
- **UPDATE**: Every Update() call with message types
- **KEY**: Key press handling
- **HANDLER**: Event handler execution
- **PANIC**: Panic reporting flow
- **ERROR**: Any errors encountered

## Expected Flow When Pressing 'p':

```
[UPDATE] Update called with msg type: tea.KeyMsg
[UPDATE] Received KeyMsg: p
[KEY] User pressed 'p' - about to trigger panic
[KEY] About to call Emit('panic', nil)
[HANDLER] panic event handler called
[HANDLER] About to panic!
[PANIC] ReportPanic called: component=Calculator, event=panic, panic=Intentional panic...
[PANIC] Sending errorMsg to program
[PANIC] errorMsg sent successfully
[KEY] Emit('panic', nil) returned successfully
[UPDATE] Update called with msg type: main.errorMsg
[UPDATE] Received errorMsg: Panic in 'Calculator.panic': Intentional panic...
```

## If It Freezes

Look for where the log stops. Common issues:
- Stops after "About to panic!" → panic not being caught properly
- Stops after "Sending errorMsg" → program.Send() blocking
- Stops after "Emit returned" → Update() not being called again
- No "Received errorMsg" → errorMsg not reaching Update()

## Debugging Tips

1. The log is flushed immediately after each write (Sync() called)
2. If the app crashes, the log will still contain all entries up to the crash
3. Check timestamps to see where delays occur
4. Look for missing expected log entries
