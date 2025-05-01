# Sidecar

The sidecar is what is injected into GameServer pods. It functionally acts as a 
simple rest server, with two booleans being saved internally.

## Documentation
There is more info available about the sidecar [here](https://unfamousthomas.github.io/thesis-initial/service/).
## Testing
Currently, the automatic tests are setup for a few of the handlers. To run these tests:

```bash
    go test github.com/unfamousthomas/thesis-sidecar/internal/handlers
```