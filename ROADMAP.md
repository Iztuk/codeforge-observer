# Observer Roadmap

## Persistence

Right now host config and findings are in memory only

**Goal**

- survive daemon restarts
- make dynamic host management durable
- ensure findings are saved to the database

**Implementation**

- SQLite database

**Done Signal**

Hosts

- Add host
- Restart daemon
- Host is still present

Findings

- Record a finding
- Ensure finding is written to the database

## Resource Validation Layer

Add another layer of validation based on the Resources contract after API contract audit

**Goals**

- Compare Resources contract with the Request/Response payload
- Create and log Findings based on the result

**Done Signal**

- Request resource validation works for basic field rules
- Response readable-field validation works
- Findings cleanly distinguish where it is being emitted

## Runtime Control Plane Reloading

Expand on add/remove/list hosts

**Goals**

- Attach API/Resource contract
- Reload host configurations
- Show host details

**Done Signal**

- Hosts and contracts can be managed live through the socket interface

## Findings Output / Reporting

**Goals**

- Persist or emit findings in a queryable way
- Distinguish API vs resource findings clearly
- Support future summaries and analytics (REST Endpoint | gRPC streaming)

**Done Signal**

- Findings can be stored, filtered, and reviewed by host/path/code/stage
