# Notes

## Go

[Example](https://github.com/golang-standards/project-layout)

pkg/:
Contains reusable libraries or utilities that can be imported by multiple parts of your project — including cmd/, internal/, or even other projects if you want.
Think: helpers, clients, wrappers, utility functions that are generally usable.

internal/:
Holds application-specific code that is private to your project. Packages here cannot be imported by external projects (enforced by Go).
This is where your core business logic (e.g., matchmaking algorithms, queue handling) usually lives.

cmd/:
Contains main packages — the executable entry points. For example, cmd/server/main.go is the actual server startup script that wires everything together and launches the app.
