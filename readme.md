# envelope

[tips](https://leg100.github.io/en/posts/building-bubbletea-programs/)


1. Log all bubbletea messages to a file, and tail that file in another terminal
2. Automatically re-build and re-run your app whenever you make changes, so you can see changes in real-time (similar to livereload with web apps). See my scripts in `./hacks`.
3. Check the bubbletea examples in their repo. Run and tinker with them.
4. Understand the Update() loop really well. Appreciate that it is single-threaded and use that to your advantage. Keep any I/O or slow code out of the loop; instead put it in a tea.Cmd, which bubbletea runs in a goroutine.
5. Some of the "bubbles", i.e. their default components, are very useful, particularly the lower-level ones such as "viewport" and "spinner", whereas their higher-level bubbles such as the "table", "list" and "help" I found to be too uncustomisable to be useful.
6. Getting the layout right in a full screen TUI can be very difficult. You have to measure the heights and widths of everything carefully to prevent lines wrapping or escaping the visible area.
7. If your code panics, bubbletea is meant to rescue the panic but in practice it doesn't do that and instead it can leave your terminal in a half baked state. When that happens type `reset`.

The key to writing a fast model is to offload expensive operations to a tea.Cmd:

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
        // don't do this:
        // time.Sleep(time.Minute)

        // do this:
        return m, func() tea.Msg {
            time.Sleep(time.Minute)
        }
	}
	return m, nil
}
```
