# packages/

Shared libraries used by more than one app (`@servienta/*`). A library is extracted here only when
a second consumer exists — never speculatively. Candidates when they become real: the license file
format (admin issues, engine verifies), the engine HTTP-contract client (console consumes).
