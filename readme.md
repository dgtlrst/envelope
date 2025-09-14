# envelope

[tips](https://leg100.github.io/en/posts/building-bubbletea-programs/)

project
├── cmd                      # command-related files
│   └── app                  # Application entry point
│       └── main.go          # main application logic
├── internal                 # internal codebase
│   ├── handlers             # HTTP request handlers (controllers)
│   │   └── user_handler.go  # user-specific handler
│   ├── services             # business logic (service layer)
│   │   └── user_service.go  # user-specific service
│   ├── repositories         # data access (repository layer)
│   │   └── user_repo.go     # user-specific repository
│   └── models               # data models (entities)
│       └── user.go          # user model
├── pkg                      # shared utilities or helpers
├── configs                  # configuration files
├── go.mod                   # go module definition
└── go.sum                   # go module checksum file
