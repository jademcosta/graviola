package app_test

const configOneGroupWithOneRemote = `
api:
  port: 8091
  timeout: 3m

log:
  level: error

storages:
  merge_strategy:
    type: keep_biggest
  groups:
    - name: "the solo group"
      on_query_fail: fail_all
      remotes:
        - name: "the server 1"
          address: "http://localhost:9090"
`
