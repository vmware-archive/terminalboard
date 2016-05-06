pipeline_status can include
- success (succeeded)
- failure (failed, errored)
- stopped (paused, aborted)

currently_running is boolean
- true means it is started and should flash
- false means not running and is solid

next_build is null = false
next_build not null, pending = false
next_build not null, started = true

like we discussed, we could add more to the hashes if we get there
