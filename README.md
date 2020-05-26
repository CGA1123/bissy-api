# bissy-api

### Ideas

#### A slack <-> github PR state sync
  - When a PR is posted, listen to it and its events, update the message with reactions based on state
    - :merged: :ci-pass: :ci-fail: :approved: :changes-requested: :comments:

#### A github <-> kanbanize state sync

#### slackerduty rebuild

#### trevor but cached, with okta signin and configurable refresh rate
  Query is:
    - query id
    - query name
    - query sql
    - query refresh rate

  API:
    - POST query
    - PATCH query/:id
    - DELETE query
    - PURGE query/:id
    - GET query/:id
    - GET query/:id/result?gsheet=true # include a last_updated somehow?

  - Start with basic auth
  - Add okta
  - Allow configuring
