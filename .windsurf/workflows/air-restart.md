# Air Restart Workflow

This workflow helps restart the Air development server with the latest configuration.

## Steps

1. Stop any running Air instances (Ctrl+C in the terminal where Air is running)
2. Clear the Air cache (if needed):
   ```bash
   air init
   ```
3. Start Air with the updated configuration:
   ```bash
   air
   ```
4. Verify that Air is watching the correct files by checking the initial output for the list of watched directories

## Troubleshooting

- If changes are still not detected, try:
  - Deleting the `tmp` directory and restarting Air
  - Running `air init` to regenerate the Air configuration
  - Checking for any processes using the same port (e.g., `lsof -i :8080` on Unix-like systems)
